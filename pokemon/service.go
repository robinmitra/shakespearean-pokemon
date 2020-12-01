package pokemon

import (
	"encoding/json"
	"net/http"
)

type Service struct {
	BaseUrl string
}

func NewService() *Service {
	return &Service{"https://pokeapi.co/api/v2/pokemon-species"}
}

func (p *Service) Get(name string) (*Pokemon, error) {
	res, err := http.Get(p.BaseUrl + "/" + name)
	if err != nil {
		return nil, err
	}
	if res.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		return nil, nil
	}
	type ResJson struct {
		Entries []struct {
			Text     string `json:"flavor_text"`
			Language struct {
				Name string `json:"name"`
			} `json:"language"`
		} `json:"flavor_text_entries"`
	}
	var resJson ResJson
	if err := json.NewDecoder(res.Body).Decode(&resJson); err != nil {
		return nil, err
	}
	var pokemon Pokemon
	for _, entry := range resJson.Entries {
		if entry.Language.Name == "en" {
			pokemon = Pokemon{
				Name:        name,
				Description: entry.Text,
			}
			break
		}
	}
	if pokemon == (Pokemon{}) {
		return nil, nil
	}

	return &pokemon, nil
}

type Pokemon struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
