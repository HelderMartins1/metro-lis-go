package push

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/HelderMartins1/metro-lis-go/models"
)

const (
	tick        = 15 * time.Second
	statusEvery = 60 * time.Second
	// A followed train is announced this long before it reaches the station.
	trainWarning = 120
	// A train alert that never fired (the train was cancelled, or the data
	// stopped) is dropped after this long.
	trainAlertLife = time.Hour
)

var lisbon = mustLocation("Europe/Lisbon")

func mustLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.FixedZone("WET", 0)
	}
	return loc
}

// Run checks for things to announce until ctx ends.
func (p *Service) Run(ctx context.Context) {
	w := &worker{p: p, lastStatus: map[string]string{}}
	t := time.NewTicker(tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			if p.store.len() == 0 {
				continue // nobody to tell: do not ask the Metro API anything
			}
			w.trains()
			w.reminders(now.In(lisbon))
			if now.Sub(w.statusAt) >= statusEvery {
				w.status()
				w.statusAt = now
			}
		}
	}
}

type worker struct {
	p          *Service
	statusAt   time.Time
	lastStatus map[string]string // line id -> last state seen; empty until the first check
	dests      map[string]string
	destsAt    time.Time
}

type arrival struct {
	destination string
	train       string
	seconds     int
}

// arrivals reads one station's answer as a flat list of trains.
func arrivals(entries []map[string]interface{}) []arrival {
	var out []arrival
	for _, e := range entries {
		dest := str(e["destino"])
		for _, k := range [][2]string{{"comboio", "tempoChegada1"}, {"comboio2", "tempoChegada2"}, {"comboio3", "tempoChegada3"}} {
			if secs, ok := seconds(e[k[1]]); ok && str(e[k[0]]) != "" {
				out = append(out, arrival{dest, str(e[k[0]]), secs})
			}
		}
	}
	return out
}

func (w *worker) destinations() map[string]string {
	if w.dests == nil || time.Since(w.destsAt) > time.Hour {
		if d, err := w.p.src.Destinations(); err == nil {
			w.dests, w.destsAt = d, time.Now()
		}
	}
	return w.dests
}

// trains announces followed trains that are about to arrive.
func (w *worker) trains() {
	type due struct {
		sub Subscription
		msg Message
	}
	var stations []string
	w.p.store.each(func(sub *Subscription) bool {
		for _, t := range sub.Trains {
			if !slices.Contains(stations, t.Station) {
				stations = append(stations, t.Station)
			}
		}
		return false
	})
	if len(stations) == 0 {
		return
	}

	boards := map[string][]arrival{}
	for _, code := range stations {
		if entries, err := w.p.src.Trains(code); err == nil {
			boards[code] = arrivals(entries)
		}
	}
	dests := w.destinations()

	var send []due
	w.p.store.each(func(sub *Subscription) bool {
		kept := sub.Trains[:0]
		for _, t := range sub.Trains {
			board, answered := boards[t.Station]
			if time.Since(t.CreatedAt) > trainAlertLife {
				continue
			}
			if !answered {
				kept = append(kept, t) // no data this time: try again next tick
				continue
			}
			i := slices.IndexFunc(board, func(a arrival) bool { return a.train == t.Train && a.destination == t.Destination })
			if i < 0 {
				// Gone from the board: it already passed (or was taken off). Nothing to say.
				if time.Since(t.CreatedAt) > 2*tick {
					continue
				}
				kept = append(kept, t)
				continue
			}
			if board[i].seconds > trainWarning {
				kept = append(kept, t)
				continue
			}
			send = append(send, due{*sub, trainMessage(t, board[i].seconds, dests)})
		}
		changed := len(kept) != len(sub.Trains)
		sub.Trains = kept
		return changed
	})
	for _, d := range send {
		w.p.deliver(d.sub, d.msg)
	}
}

func trainMessage(t TrainAlert, secs int, dests map[string]string) Message {
	station := stationName(t.Station)
	when := "está a chegar"
	if secs > 30 {
		when = fmt.Sprintf("chega dentro de %d min", (secs+59)/60)
	}
	title := "O teu comboio " + when
	if dest := dests[t.Destination]; dest != "" {
		title = fmt.Sprintf("Comboio para %s %s", dest, when)
	}
	return Message{
		Title: title,
		Body:  station + ". Toca para veres os tempos.",
		URL:   "/#/" + t.Station,
		Tag:   "train-" + t.ID,
	}
}

// status announces changes in a line's state to everyone whose favourite stations it serves.
func (w *worker) status() {
	status, err := w.p.src.Status()
	if err != nil {
		return
	}
	first := len(w.lastStatus) == 0
	type change struct{ line, state string }
	var changes []change
	for id, state := range status {
		norm := strings.ToLower(strings.TrimSpace(state))
		if norm == "" {
			continue
		}
		if prev, ok := w.lastStatus[id]; !first && ok && prev != norm {
			changes = append(changes, change{id, strings.TrimSpace(state)})
		}
		w.lastStatus[id] = norm
	}
	// The first check only learns how things are; it announces nothing.
	if first || len(changes) == 0 {
		return
	}

	for _, c := range changes {
		line, _ := models.LineByID(c.line)
		serves := map[string]bool{}
		for _, s := range line.Stations {
			serves[s.Code] = true
		}
		msg := statusMessage(line, c.state)
		var subs []Subscription
		w.p.store.each(func(sub *Subscription) bool {
			if slices.ContainsFunc(sub.Disruptions, func(code string) bool { return serves[code] }) {
				subs = append(subs, *sub)
			}
			return false
		})
		for _, sub := range subs {
			w.p.deliver(sub, msg)
		}
	}
}

func statusMessage(line models.Line, state string) Message {
	m := Message{URL: "/#/linha/" + slug(line.ID), Tag: "status-" + line.ID}
	if strings.EqualFold(state, "normal") || strings.EqualFold(state, "ok") {
		m.Title = fmt.Sprintf("Linha %s: circulação normal", line.Name)
		m.Body = "A linha voltou ao normal."
		return m
	}
	m.Title = fmt.Sprintf("Linha %s: %s", line.Name, strings.ToLower(state))
	m.Body = "Pode haver atrasos nas tuas estações. Toca para veres os comboios na linha."
	return m
}

// reminders sends each reminder due this minute, once a day.
func (w *worker) reminders(now time.Time) {
	hhmm := now.Format("15:04")
	day := now.Year()*10000 + int(now.Month())*100 + now.Day()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	type due struct {
		sub Subscription
		rem Reminder
	}
	var send []due
	w.p.store.each(func(sub *Subscription) bool {
		for _, r := range sub.Reminders {
			if r.Time != hhmm || !slices.Contains(r.Days, weekday) || sub.sent[r.ID] == day {
				continue
			}
			if sub.sent == nil {
				sub.sent = map[string]int{}
			}
			sub.sent[r.ID] = day
			send = append(send, due{*sub, r})
		}
		return false
	})
	if len(send) == 0 {
		return
	}

	dests := w.destinations()
	boards := map[string][]arrival{}
	for _, d := range send {
		code := d.rem.Station
		if _, ok := boards[code]; !ok {
			entries, err := w.p.src.Trains(code)
			if err != nil {
				log.Printf("push: reminder for %s: %v", code, err)
			}
			boards[code] = arrivals(entries)
		}
		w.p.deliver(d.sub, reminderMessage(code, boards[code], dests))
	}
}

func reminderMessage(code string, board []arrival, dests map[string]string) Message {
	station := stationName(code)
	// The soonest train per destination, skipping trains that end here.
	soonest := map[string]int{}
	var order []string
	for _, a := range board {
		name := dests[a.destination]
		if name == "" || name == station {
			continue
		}
		if prev, ok := soonest[name]; !ok || a.seconds < prev {
			if !ok {
				order = append(order, name)
			}
			soonest[name] = a.seconds
		}
	}
	m := Message{Title: station, URL: "/#/" + code, Tag: "reminder-" + code}
	if len(order) == 0 {
		m.Body = "Sem tempos de espera neste momento."
		return m
	}
	parts := make([]string, 0, len(order))
	for _, name := range order {
		mins := "a chegar"
		if s := soonest[name]; s > 30 {
			mins = fmt.Sprintf("%d min", (s+59)/60)
		}
		parts = append(parts, fmt.Sprintf("%s: %s", name, mins))
	}
	m.Body = strings.Join(parts, " · ")
	return m
}

func stationName(code string) string {
	for _, l := range models.Lines {
		for _, s := range l.Stations {
			if s.Code == code {
				return s.Name
			}
		}
	}
	return code
}

// slug is the line's name in the app's #/linha/… address.
func slug(id string) string {
	return map[string]string{"blue": "azul", "yellow": "amarela", "green": "verde", "red": "vermelha"}[id]
}

func str(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	}
	return ""
}

func seconds(v interface{}) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case string:
		n, err := strconv.Atoi(t)
		return n, err == nil
	}
	return 0, false
}
