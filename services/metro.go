package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/HelderMartins1/metro-lis-go/config"
	"github.com/HelderMartins1/metro-lis-go/models"
)

var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func get(cfg *config.Config, endpoint string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", cfg.URL+endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetLines(cfg *config.Config) (map[string]interface{}, error) {
	req, err := get(cfg, "estadoLinha/todos")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	code, _ := req["codigo"].(string)
	if code < "200" || code >= "300" && code != "404" {

		return nil, fmt.Errorf("error: %v", req["codigo"])
	}
	resp, ok := req["resposta"].(map[string]interface{})
	if !ok {
		msg, _ := req["resposta"].(string)
		return map[string]interface{}{
			"amarela":  msg,
			"azul":     msg,
			"vermelha": msg,
			"verde":    msg,
		}, nil
	}
	return map[string]interface{}{
		"amarela":  resp["amarela_curta"],
		"azul":     resp["azul_curta"],
		"vermelha": resp["vermelha_curta"],
		"verde":    resp["verde_curta"],
	}, nil
}

func GetDestinations(cfg *config.Config) (map[string]interface{}, error) {
	dest, err := get(cfg, "infoDestinos/todos")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	code, _ := dest["codigo"].(string)
	if code < "200" || code >= "300" {
		return nil, fmt.Errorf("error: %v", dest["codigo"])
	}

	resp, _ := dest["resposta"].([]interface{})

	toSet := func(m map[string]string) map[string]struct{} {
		s := make(map[string]struct{}, len(m))
		for _, name := range m {
			s[name] = struct{}{}
		}
		return s
	}
	lines := map[string]map[string]struct{}{
		"blue":   toSet(models.BlueLine),
		"red":    toSet(models.RedLine),
		"green":  toSet(models.GreenLine),
		"yellow": toSet(models.YellowLine),
	}

	mapped := map[string][]interface{}{
		"yellow": {},
		"blue":   {},
		"green":  {},
		"red":    {},
	}

	for _, d := range resp {
		destination, ok := d.(map[string]interface{})
		if !ok {
			continue
		}
		nome, _ := destination["nome_destino"].(string)
		for line, set := range lines {
			if _, found := set[nome]; found {
				mapped[line] = append(mapped[line], destination)
			}
		}
	}

	out := make(map[string]interface{}, len(mapped))
	for k, v := range mapped {
		out[k] = v
	}
	return out, nil
}

func GetStations(station string) map[string]string {
	switch station {
	case "yellow":
		return models.YellowLine
	case "blue":
		return models.BlueLine
	case "green":
		return models.GreenLine
	case "red":
		return models.RedLine
	default:
		return map[string]string{}
	}
}

func GetTrains(cfg *config.Config, station string) (interface{}, error) {
	req, err := get(cfg, "tempoEspera/Estacao/"+station)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	code, _ := req["codigo"].(string)
	if code < "200" || code >= "300" {
		return nil, fmt.Errorf("error: %v", req["codigo"])
	}
	return req["resposta"], nil
}
