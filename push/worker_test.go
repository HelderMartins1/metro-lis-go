package push

import (
	"strings"
	"testing"
	"time"
)

type fakeSource struct {
	status map[string]string
	boards map[string][]map[string]interface{}
}

func (f *fakeSource) Trains(code string) ([]map[string]interface{}, error) {
	return f.boards[code], nil
}
func (f *fakeSource) Destinations() (map[string]string, error) {
	return map[string]string{"33": "Reboleira", "42": "Santa Apolónia"}, nil
}
func (f *fakeSource) Status() (map[string]string, error) { return f.status, nil }

type sent struct {
	endpoint string
	msg      Message
}

func newTestService(t *testing.T, src *fakeSource) (*Service, *[]sent) {
	t.Helper()
	p, err := New(src, t.TempDir(), "https://example.test")
	if err != nil {
		t.Fatal(err)
	}
	var out []sent
	p.sender = func(sub Subscription, m Message) error {
		out = append(out, sent{sub.Endpoint, m})
		return nil
	}
	return p, &out
}

func subscribe(t *testing.T, p *Service, endpoint string, fn func(*Subscription)) {
	t.Helper()
	err := p.store.update(endpoint, func(*Subscription) *Subscription {
		sub := &Subscription{Endpoint: endpoint}
		fn(sub)
		return sub
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStatusChangesReachOnlyAffectedStations(t *testing.T) {
	src := &fakeSource{status: map[string]string{"blue": "normal", "green": "normal"}}
	p, out := newTestService(t, src)
	subscribe(t, p, "https://push.test/blue-fan", func(s *Subscription) { s.Disruptions = []string{"RB"} })
	subscribe(t, p, "https://push.test/green-fan", func(s *Subscription) { s.Disruptions = []string{"TE"} })
	w := &worker{p: p, lastStatus: map[string]string{}}

	w.status()
	if len(*out) != 0 {
		t.Fatalf("the first check should only learn the state, sent %v", *out)
	}

	src.status = map[string]string{"blue": "Condicionada", "green": "normal"}
	w.status()
	if len(*out) != 1 || (*out)[0].endpoint != "https://push.test/blue-fan" {
		t.Fatalf("want one message to the blue line's fan, got %v", *out)
	}
	if !strings.Contains((*out)[0].msg.Title, "Azul") {
		t.Errorf("title %q should name the line", (*out)[0].msg.Title)
	}

	w.status()
	if len(*out) != 1 {
		t.Errorf("an unchanged state should not be announced again, got %d messages", len(*out))
	}
}

func TestRemindersGoOutOncePerDay(t *testing.T) {
	src := &fakeSource{boards: map[string][]map[string]interface{}{
		"BC": {{"destino": "33", "comboio": "1A", "tempoChegada1": float64(150)}},
	}}
	p, out := newTestService(t, src)
	subscribe(t, p, "https://push.test/a", func(s *Subscription) {
		s.Reminders = []Reminder{{ID: "r1", Station: "BC", Time: "08:15", Days: []int{1, 2, 3, 4, 5}}}
	})
	w := &worker{p: p, lastStatus: map[string]string{}}

	monday := time.Date(2026, 11, 16, 8, 15, 10, 0, lisbon)
	w.reminders(monday)
	w.reminders(monday.Add(15 * time.Second)) // the next tick, same minute
	if len(*out) != 1 {
		t.Fatalf("want one reminder, got %d", len(*out))
	}
	if body := (*out)[0].msg.Body; body != "Reboleira: 3 min" {
		t.Errorf("body = %q", body)
	}

	w.reminders(time.Date(2026, 11, 22, 8, 15, 0, 0, lisbon)) // a Sunday
	if len(*out) != 1 {
		t.Errorf("no reminder on a day it is not set for, got %d", len(*out))
	}
}

func TestFollowedTrainIsAnnouncedOnceNearArrival(t *testing.T) {
	src := &fakeSource{boards: map[string][]map[string]interface{}{
		"BC": {{"destino": "33", "comboio": "7A", "tempoChegada1": float64(300)}},
	}}
	p, out := newTestService(t, src)
	subscribe(t, p, "https://push.test/a", func(s *Subscription) {
		s.Trains = []TrainAlert{{ID: "t1", Station: "BC", Destination: "33", Train: "7A", CreatedAt: time.Now()}}
	})
	w := &worker{p: p, lastStatus: map[string]string{}}

	w.trains()
	if len(*out) != 0 {
		t.Fatalf("5 min away is too early, sent %v", *out)
	}

	src.boards["BC"][0]["tempoChegada1"] = float64(110)
	w.trains()
	if len(*out) != 1 || !strings.Contains((*out)[0].msg.Title, "Reboleira") {
		t.Fatalf("want the arrival message, got %v", *out)
	}

	w.trains()
	if len(*out) != 1 {
		t.Errorf("a train alert fires once, got %d messages", len(*out))
	}
	if sub, _ := p.store.get("https://push.test/a"); len(sub.Trains) != 0 {
		t.Errorf("the alert should be gone after firing, have %v", sub.Trains)
	}
}
