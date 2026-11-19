package services

import (
	"testing"

	"github.com/HelderMartins1/metro-lis-go/models"
)

var testLine = models.Line{ID: "test", Stations: []models.Station{
	{Code: "A", Name: "Alfa"}, {Code: "B", Name: "Bravo"}, {Code: "C", Name: "Charlie"}, {Code: "D", Name: "Delta"},
}}

var testDests = map[string]string{"1": "Alfa", "4": "Delta", "9": "Elsewhere"}

func entry(dest string, trains ...interface{}) map[string]interface{} {
	e := map[string]interface{}{"destino": dest}
	keys := [][2]string{{"comboio", "tempoChegada1"}, {"comboio2", "tempoChegada2"}, {"comboio3", "tempoChegada3"}}
	for i := 0; i+1 < len(trains); i += 2 {
		e[keys[i/2][0]], e[keys[i/2][1]] = trains[i], trains[i+1]
	}
	return e
}

func TestComputePlacesTrainBeforeItsNextStation(t *testing.T) {
	perStation := [][]map[string]interface{}{
		{entry("4", "T1", float64(400))},                              // Alfa: T1 is far away (it is on the other track)
		{entry("4", "T2", float64(30))},                               // Bravo
		{entry("4", "T2", float64(130))},                              // Charlie: B -> C takes 100 s
		{entry("1", "T3", float64(50)), entry("9", "X", float64(10))}, // Delta: T3 heads back; X is another line's
	}

	got, err := compute(testLine, testDests, perStation)
	if err != nil {
		t.Fatal(err)
	}

	byID := map[string]Train{}
	for _, tr := range got.Trains {
		byID[tr.Direction+tr.ID] = tr
	}

	if _, ok := byID["upT1"]; ok {
		t.Error("a train still to reach the first station should not be on the line")
	}
	if _, ok := byID["upX"]; ok {
		t.Error("a train to another line's destination should be ignored")
	}

	t2 := byID["upT2"]
	if t2.Next != 1 || t2.NextSeconds != 30 || t2.Destination != "Delta" {
		t.Errorf("T2 = %+v, want next Bravo in 30 s towards Delta", t2)
	}
	if t2.SegmentSeconds != defaultSegment {
		t.Errorf("A -> B was never observed, want the default segment, got %d", t2.SegmentSeconds)
	}

	// T3 is seen at the last station heading down: it is there, not on the line yet.
	if _, ok := byID["downT3"]; ok {
		t.Error("a train still to reach the last station on a down run should not be on the line")
	}
}

func TestSegmentTimesUseMedian(t *testing.T) {
	seen := map[string]map[string][]sighting{
		"up": {
			"a": {{0, 10}, {1, 100}},
			"b": {{0, 200}, {1, 300}},
			"c": {{0, 500}, {1, 700}},
		},
	}
	got := segmentTimes(2, seen)
	if got[0] != 100 {
		t.Errorf("segment 0 = %d, want the median 100", got[0])
	}
}

func TestToSeconds(t *testing.T) {
	cases := map[interface{}]int{float64(42): 42, "17": 17, "--": -1, nil: -1}
	for in, want := range cases {
		if got := toSeconds(in); got != want {
			t.Errorf("toSeconds(%v) = %d, want %d", in, got, want)
		}
	}
}
