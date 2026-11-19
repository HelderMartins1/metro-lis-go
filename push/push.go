// Package push sends Web Push notifications: a followed train about to
// arrive, a disruption on a favourite station's line, and daily reminders.
package push

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/HelderMartins1/metro-lis-go/models"
	"github.com/HelderMartins1/metro-lis-go/services"
	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/go-chi/chi/v5"
)

// Limits per device, so one subscription cannot grow without bound.
const (
	maxReminders   = 20
	maxTrainAlerts = 10
	maxBody        = 16 << 10
)

type keys struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

type Service struct {
	src      services.Source
	store    *store
	keys     keys
	contact  string
	stations map[string]bool
	// sender delivers one message; tests swap it to see what would go out.
	sender func(Subscription, Message) error
}

// New opens (or creates) the push data in dir: the VAPID keys and the
// subscriptions. contact is the address push services can reach the sender
// at (an https URL or an e-mail address).
func New(src services.Source, dir, contact string) (*Service, error) {
	k, err := loadKeys(filepath.Join(dir, "vapid.json"))
	if err != nil {
		return nil, fmt.Errorf("vapid keys: %w", err)
	}
	st, err := openStore(filepath.Join(dir, "subscriptions.json"))
	if err != nil {
		return nil, fmt.Errorf("subscriptions: %w", err)
	}
	stations := map[string]bool{}
	for _, l := range models.Lines {
		for _, s := range l.Stations {
			stations[s.Code] = true
		}
	}
	p := &Service{src: src, store: st, keys: k, contact: contact, stations: stations}
	p.sender = p.send
	return p, nil
}

// loadKeys reads the server's VAPID key pair, making one the first time.
// The pair must not change: every subscription is tied to the public key.
func loadKeys(path string) (keys, error) {
	var k keys
	data, err := os.ReadFile(path)
	if err == nil {
		return k, json.Unmarshal(data, &k)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return k, err
	}
	k.Private, k.Public, err = webpush.GenerateVAPIDKeys()
	if err != nil {
		return k, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return k, err
	}
	data, _ = json.Marshal(k)
	return k, os.WriteFile(path, data, 0o600)
}

// Message is what the service worker turns into a notification.
type Message struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"` // the app address a tap opens
	Tag   string `json:"tag"` // a newer notification with the same tag replaces the older
}

var errGone = errors.New("subscription expired")

func (p *Service) send(sub Subscription, m Message) error {
	payload, _ := json.Marshal(m)
	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpush.Keys{P256dh: sub.Keys.P256dh, Auth: sub.Keys.Auth},
	}, &webpush.Options{
		Subscriber:      p.contact,
		VAPIDPublicKey:  p.keys.Public,
		VAPIDPrivateKey: p.keys.Private,
		TTL:             300,
		Urgency:         webpush.UrgencyHigh,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return errGone
	case resp.StatusCode >= 300:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return fmt.Errorf("push service: %s %s", resp.Status, body)
	}
	return nil
}

// deliver sends m and forgets the device if its push service says it is gone.
func (p *Service) deliver(sub Subscription, m Message) {
	err := p.sender(sub, m)
	if errors.Is(err, errGone) {
		_ = p.store.update(sub.Endpoint, func(*Subscription) *Subscription { return nil })
		return
	}
	if err != nil {
		log.Printf("push: %v", err)
	}
}

// Handler serves the push API, mounted at /api/push.
func (p *Service) Handler() http.Handler {
	r := chi.NewRouter()
	r.Get("/key", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"publicKey": p.keys.Public})
	})
	r.Put("/subscription", p.putSubscription)
	r.Post("/subscription/delete", p.deleteSubscription)
	r.Post("/trains", p.addTrain)
	r.Post("/trains/delete", p.deleteTrain)
	return r
}

type subscriptionJSON struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

var clock = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

func (p *Service) putSubscription(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Subscription subscriptionJSON `json:"subscription"`
		Disruptions  []string         `json:"disruptions"`
		Reminders    []Reminder       `json:"reminders"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	s := body.Subscription
	if u, err := url.Parse(s.Endpoint); err != nil || u.Scheme != "https" || s.Keys.P256dh == "" || s.Keys.Auth == "" {
		fail(w, "subscription is not a push subscription")
		return
	}
	for _, code := range body.Disruptions {
		if !p.stations[code] {
			fail(w, "unknown station "+code)
			return
		}
	}
	if len(body.Reminders) > maxReminders {
		fail(w, fmt.Sprintf("at most %d reminders", maxReminders))
		return
	}
	for i, rem := range body.Reminders {
		if !p.stations[rem.Station] || !clock.MatchString(rem.Time) || len(rem.Days) == 0 {
			fail(w, "a reminder needs a station, a time like 08:15 and at least one day")
			return
		}
		for _, d := range rem.Days {
			if d < 1 || d > 7 {
				fail(w, "days go from 1 (Monday) to 7 (Sunday)")
				return
			}
		}
		if rem.ID == "" {
			body.Reminders[i].ID = newID()
		}
	}

	err := p.store.update(s.Endpoint, func(sub *Subscription) *Subscription {
		if sub == nil {
			sub = &Subscription{Endpoint: s.Endpoint}
		}
		sub.Keys.P256dh, sub.Keys.Auth = s.Keys.P256dh, s.Keys.Auth
		sub.Disruptions = body.Disruptions
		sub.Reminders = body.Reminders
		return sub
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not save"})
		return
	}
	sub, _ := p.store.get(s.Endpoint)
	writeJSON(w, http.StatusOK, sub)
}

func (p *Service) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Endpoint string `json:"endpoint"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	_ = p.store.update(body.Endpoint, func(*Subscription) *Subscription { return nil })
	w.WriteHeader(http.StatusNoContent)
}

func (p *Service) addTrain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Endpoint    string `json:"endpoint"`
		Station     string `json:"station"`
		Destination string `json:"destination"`
		Train       string `json:"train"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	if !p.stations[body.Station] || body.Destination == "" || body.Train == "" {
		fail(w, "a train alert needs a station, a destination and a train")
		return
	}
	alert := TrainAlert{
		ID: newID(), Station: body.Station, Destination: body.Destination,
		Train: body.Train, CreatedAt: time.Now().UTC(),
	}
	var missing, full bool
	err := p.store.update(body.Endpoint, func(sub *Subscription) *Subscription {
		if sub == nil {
			missing = true
			return nil
		}
		// One alert per platform: following another train there replaces it.
		kept := sub.Trains[:0]
		for _, t := range sub.Trains {
			if t.Station != alert.Station || t.Destination != alert.Destination {
				kept = append(kept, t)
			}
		}
		if len(kept) >= maxTrainAlerts {
			full = true
			sub.Trains = kept
			return sub
		}
		sub.Trains = append(kept, alert)
		return sub
	})
	switch {
	case missing:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "turn notifications on first"})
	case full:
		fail(w, fmt.Sprintf("at most %d trains at a time", maxTrainAlerts))
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not save"})
	default:
		writeJSON(w, http.StatusOK, alert)
	}
}

func (p *Service) deleteTrain(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Endpoint string `json:"endpoint"`
		ID       string `json:"id"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	_ = p.store.update(body.Endpoint, func(sub *Subscription) *Subscription {
		if sub == nil {
			return nil
		}
		kept := sub.Trains[:0]
		for _, t := range sub.Trains {
			if t.ID != body.ID {
				kept = append(kept, t)
			}
		}
		sub.Trains = kept
		return sub
	})
	w.WriteHeader(http.StatusNoContent)
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func readJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody)).Decode(v); err != nil {
		fail(w, "the request body is not valid JSON")
		return false
	}
	return true
}

func fail(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
