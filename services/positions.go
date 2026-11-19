package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/HelderMartins1/metro-lis-go/config"
	"github.com/HelderMartins1/metro-lis-go/models"
)

// Source is where live data comes from: the Metro API itself, or another
// instance of this server.
type Source interface {
	Trains(station string) ([]map[string]interface{}, error)
	Destinations() (map[string]string, error) // destination id -> station name
	Status() (map[string]string, error)       // line id -> its state ("normal", "condicionada", …)
}

// statusKeys maps the API's Portuguese line names to line ids.
var statusKeys = map[string]string{"azul": "blue", "amarela": "yellow", "verde": "green", "vermelha": "red"}

func byLineID(status map[string]interface{}) map[string]string {
	out := make(map[string]string, len(status))
	for name, id := range statusKeys {
		out[id] = toString(status[name])
	}
	return out
}

// Train is one train on the line. It is NextSeconds away from the station at
// Next (an index into the line's stations), on a stretch that takes about
// SegmentSeconds; clients count NextSeconds down between refreshes.
type Train struct {
	ID             string `json:"id"`
	Direction      string `json:"direction"` // "up": towards the last station, "down": towards the first
	Destination    string `json:"destination"`
	Next           int    `json:"next"`
	NextSeconds    int    `json:"nextSeconds"`
	SegmentSeconds int    `json:"segmentSeconds"`
}

type Positions struct {
	Line      string    `json:"line"`
	UpdatedAt time.Time `json:"updatedAt"`
	Trains    []Train   `json:"trains"`
}

// Without two observations of a stretch, assume a typical one.
const defaultSegment = 100

// Positions are shared by every client for this long, so the load on the
// Metro API does not grow with the number of people watching.
const positionsTTL = 15 * time.Second

type cached struct {
	mu   sync.Mutex
	at   time.Time
	data *Positions
}

type PositionsService struct {
	src   Source
	lines sync.Map // line id -> *cached
	dests struct {
		sync.Mutex
		at   time.Time
		byID map[string]string
	}
}

func NewPositions(src Source) *PositionsService {
	return &PositionsService{src: src}
}

func (p *PositionsService) Get(lineID string) (*Positions, error) {
	line, ok := models.LineByID(lineID)
	if !ok {
		return nil, fmt.Errorf("unknown line %q", lineID)
	}

	v, _ := p.lines.LoadOrStore(lineID, &cached{})
	c := v.(*cached)
	// Holding the lock while computing makes concurrent requests wait for one fetch.
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data != nil && time.Since(c.at) < positionsTTL {
		return c.data, nil
	}

	dests, err := p.destinations()
	if err != nil {
		return nil, err
	}
	data, err := compute(line, dests, p.fetchAll(line))
	if err != nil {
		return nil, err
	}
	c.data, c.at = data, time.Now()
	return data, nil
}

func (p *PositionsService) destinations() (map[string]string, error) {
	p.dests.Lock()
	defer p.dests.Unlock()
	if p.dests.byID != nil && time.Since(p.dests.at) < time.Hour {
		return p.dests.byID, nil
	}
	byID, err := p.src.Destinations()
	if err != nil {
		return nil, fmt.Errorf("destinations: %w", err)
	}
	p.dests.byID, p.dests.at = byID, time.Now()
	return byID, nil
}

// fetchAll asks for every station's times, a few at a time. A station that
// fails is left out: its trains reappear on the next refresh.
func (p *PositionsService) fetchAll(line models.Line) [][]map[string]interface{} {
	out := make([][]map[string]interface{}, len(line.Stations))
	sem := make(chan struct{}, 6)
	var wg sync.WaitGroup
	for i, s := range line.Stations {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if entries, err := p.src.Trains(s.Code); err == nil {
				out[i] = entries
			}
		}()
	}
	wg.Wait()
	return out
}

type sighting struct {
	station int
	seconds int
}

func compute(line models.Line, dests map[string]string, perStation [][]map[string]interface{}) (*Positions, error) {
	index := make(map[string]int, len(line.Stations))
	for i, s := range line.Stations {
		index[s.Name] = i
	}

	answered := 0
	// direction -> train id -> where it was seen and how far away it was.
	seen := map[string]map[string][]sighting{"up": {}, "down": {}}
	destOf := map[string]string{}
	for i, entries := range perStation {
		if entries != nil {
			answered++
		}
		for _, e := range entries {
			name := dests[toString(e["destino"])]
			d, onLine := index[name]
			if !onLine || d == i {
				continue // another line's train, or one that ends here
			}
			dir := "up"
			if d < i {
				dir = "down"
			}
			for _, k := range [][2]string{{"comboio", "tempoChegada1"}, {"comboio2", "tempoChegada2"}, {"comboio3", "tempoChegada3"}} {
				id, secs := toString(e[k[0]]), toSeconds(e[k[1]])
				if id == "" || secs < 0 {
					continue
				}
				seen[dir][id] = append(seen[dir][id], sighting{i, secs})
				destOf[dir+id] = name
			}
		}
	}
	if answered == 0 {
		return nil, fmt.Errorf("no station on the %s line answered", line.ID)
	}

	segments := segmentTimes(len(line.Stations), seen)
	last := len(line.Stations) - 1

	var trains []Train
	for dir, byID := range seen {
		for id, sightings := range byID {
			next := sightings[0]
			for _, s := range sightings[1:] {
				if s.seconds < next.seconds {
					next = s
				}
			}
			// A train still to reach the station its run starts from is on the
			// other track (or waiting to leave); the other direction shows it.
			if (dir == "up" && next.station == 0) || (dir == "down" && next.station == last) {
				continue
			}
			from := next.station - 1
			if dir == "down" {
				from = next.station + 1
			}
			trains = append(trains, Train{
				ID:             id,
				Direction:      dir,
				Destination:    destOf[dir+id],
				Next:           next.station,
				NextSeconds:    next.seconds,
				SegmentSeconds: segments[min(from, next.station)],
			})
		}
	}
	sort.Slice(trains, func(a, b int) bool {
		if trains[a].Direction != trains[b].Direction {
			return trains[a].Direction < trains[b].Direction
		}
		return trains[a].ID < trains[b].ID
	})

	return &Positions{Line: line.ID, UpdatedAt: time.Now().UTC(), Trains: trains}, nil
}

// segmentTimes estimates how long each stretch (station i to i+1) takes, as
// the median of what every train seen at both ends says about it.
func segmentTimes(n int, seen map[string]map[string][]sighting) []int {
	samples := make([][]int, n)
	for _, byID := range seen {
		for _, sightings := range byID {
			at := map[int]int{}
			for _, s := range sightings {
				at[s.station] = s.seconds
			}
			for i := 0; i+1 < n; i++ {
				a, okA := at[i]
				b, okB := at[i+1]
				if okA && okB && a != b {
					d := b - a
					if d < 0 {
						d = -d
					}
					samples[i] = append(samples[i], d)
				}
			}
		}
	}
	out := make([]int, n)
	for i, s := range samples {
		if len(s) == 0 {
			out[i] = defaultSegment
			continue
		}
		sort.Ints(s)
		out[i] = s[len(s)/2]
	}
	return out
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return ""
	}
}

// toSeconds reads an arrival time, which the API sends as a number or as a
// string ("--" when there is none). It returns -1 when there is no time.
func toSeconds(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case string:
		if n, err := strconv.Atoi(t); err == nil {
			return n
		}
	}
	return -1
}

// APISource reads straight from the Metro API.
type APISource struct{ Cfg *config.Config }

func (s APISource) Trains(station string) ([]map[string]interface{}, error) {
	data, err := GetTrains(s.Cfg, station)
	if err != nil {
		return nil, err
	}
	return asEntries(data), nil
}

func (s APISource) Destinations() (map[string]string, error) {
	byLine, err := GetDestinations(s.Cfg)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, list := range byLine {
		for _, d := range list.([]interface{}) {
			m, _ := d.(map[string]interface{})
			out[toString(m["id_destino"])] = toString(m["nome_destino"])
		}
	}
	return out, nil
}

func (s APISource) Status() (map[string]string, error) {
	status, err := GetLines(s.Cfg)
	if err != nil {
		return nil, err
	}
	return byLineID(status), nil
}

// RemoteSource reads from another instance of this server.
type RemoteSource struct{ Base *url.URL }

var remoteClient = &http.Client{Timeout: 10 * time.Second}

func (s RemoteSource) getJSON(path string, v interface{}) error {
	resp, err := remoteClient.Get(s.Base.JoinPath("api", path).String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func (s RemoteSource) Trains(station string) ([]map[string]interface{}, error) {
	var data interface{}
	if err := s.getJSON(station+"/trains", &data); err != nil {
		return nil, err
	}
	return asEntries(data), nil
}

func (s RemoteSource) Destinations() (map[string]string, error) {
	var byLine map[string][]map[string]interface{}
	if err := s.getJSON("destinations", &byLine); err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, list := range byLine {
		for _, d := range list {
			out[toString(d["id_destino"])] = toString(d["nome_destino"])
		}
	}
	return out, nil
}

func (s RemoteSource) Status() (map[string]string, error) {
	var status map[string]interface{}
	if err := s.getJSON("lines", &status); err != nil {
		return nil, err
	}
	return byLineID(status), nil
}

func asEntries(data interface{}) []map[string]interface{} {
	list, _ := data.([]interface{})
	out := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		if m, ok := item.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out
}
