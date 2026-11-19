package push

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Subscription is one device that turned notifications on, with what it asked for.
type Subscription struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
	// Stations whose lines' disruptions this device wants to hear about.
	Disruptions []string       `json:"disruptions"`
	Reminders   []Reminder     `json:"reminders"`
	Trains      []TrainAlert   `json:"trains"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	sent        map[string]int // reminder id -> day it last went out (yyyymmdd); not saved
}

// Reminder sends a station's next trains at a time of day, on some weekdays.
type Reminder struct {
	ID      string `json:"id"`
	Station string `json:"station"`
	Time    string `json:"time"` // "08:15", Lisbon time
	Days    []int  `json:"days"` // 1 = Monday … 7 = Sunday
}

// TrainAlert fires once, when a train is about to reach a station.
type TrainAlert struct {
	ID          string    `json:"id"`
	Station     string    `json:"station"`
	Destination string    `json:"destination"` // destination id, as the API gives it
	Train       string    `json:"train"`
	CreatedAt   time.Time `json:"createdAt"`
}

// store keeps the subscriptions in memory and in one JSON file.
type store struct {
	mu   sync.Mutex
	path string
	subs map[string]*Subscription // by endpoint
}

func openStore(path string) (*store, error) {
	s := &store{path: path, subs: map[string]*Subscription{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, os.MkdirAll(filepath.Dir(path), 0o700)
	}
	if err != nil {
		return nil, err
	}
	var list []*Subscription
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	for _, sub := range list {
		s.subs[sub.Endpoint] = sub
	}
	return s, nil
}

// update runs fn on the subscription for endpoint (nil when there is none)
// and saves. fn returns the subscription to keep, or nil to delete it.
func (s *store) update(endpoint string, fn func(*Subscription) *Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sub := fn(s.subs[endpoint]); sub != nil {
		sub.UpdatedAt = time.Now().UTC()
		s.subs[endpoint] = sub
	} else {
		delete(s.subs, endpoint)
	}
	return s.saveLocked()
}

// each runs fn on every subscription under the lock; fn reports whether it changed one.
func (s *store) each(fn func(*Subscription) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := false
	for _, sub := range s.subs {
		if fn(sub) {
			changed = true
		}
	}
	if changed {
		_ = s.saveLocked()
	}
}

func (s *store) get(endpoint string) (Subscription, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.subs[endpoint]
	if !ok {
		return Subscription{}, false
	}
	return *sub, true
}

func (s *store) len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subs)
}

// saveLocked writes to a temporary file and renames it, so a crash mid-write
// never leaves a half-written file behind.
func (s *store) saveLocked() error {
	list := make([]*Subscription, 0, len(s.subs))
	for _, sub := range s.subs {
		list = append(list, sub)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
