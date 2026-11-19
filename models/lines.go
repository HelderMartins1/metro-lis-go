package models

import (
	_ "embed"
	"encoding/json"
)

type Station struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Line keeps its stations in network order, from one terminus to the other.
type Line struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Stations []Station `json:"stations"`
}

func (l Line) StationMap() map[string]string {
	m := make(map[string]string, len(l.Stations))
	for _, s := range l.Stations {
		m[s.Code] = s.Name
	}
	return m
}

// network.json is the one copy of the network: the frontend imports the same
// file, so the app can draw lines and stations with no connection at all.
//
//go:embed network.json
var network []byte

var Lines = mustParse(network)

func mustParse(data []byte) []Line {
	var lines []Line
	if err := json.Unmarshal(data, &lines); err != nil {
		panic("models/network.json: " + err.Error())
	}
	return lines
}

func LineByID(id string) (Line, bool) {
	for _, l := range Lines {
		if l.ID == id {
			return l, true
		}
	}
	return Line{}, false
}
