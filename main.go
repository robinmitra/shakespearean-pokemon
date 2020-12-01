package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type HttpError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *HttpError) Error() string {
	return e.Message
}

type PokemonFetcher interface {
	Get(name string) (*Pokemon, error)
}

type PokemonService struct {
	baseUrl string
}

func NewPokemonService() *PokemonService {
	return &PokemonService{"https://pokeapi.co/api/v2/pokemon-species"}
}

func (p *PokemonService) Get(name string) (*Pokemon, error) {
	res, err := http.Get(p.baseUrl + "/" + name)
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

type PokemonHandler struct {
	pokemonService PokemonFetcher
}

type HttpResponse struct{}

func (r *HttpResponse) Send(w http.ResponseWriter, data interface{}) {
	r.setHeaders(w)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		r.Error(w, &HttpError{"Failed to encode data", http.StatusInternalServerError})
		return
	}
}

func (r *HttpResponse) Error(w http.ResponseWriter, err *HttpError) {
	r.setHeaders(w)
	w.WriteHeader(err.Status)
	if err := json.NewEncoder(w).Encode(err); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (r HttpResponse) setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func (p *PokemonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := HttpResponse{}

	character := strings.TrimPrefix(r.URL.Path, "/pokemon/")

	if strings.Contains(character, "/") {
		response.Error(w, &HttpError{"Invalid request", http.StatusBadRequest})
		return
	}

	res, err := p.pokemonService.Get(character)
	if err != nil {
		response.Error(w, &HttpError{
			fmt.Sprintf("Failed to fetch Pokemon - %v",
				err.Error()), http.StatusNotFound,
		})
		return
	}
	if res == nil {
		response.Error(w, &HttpError{"Not found", http.StatusNotFound})
		return
	}
	response.Send(w, res)
}

func main() {
	h := PokemonHandler{pokemonService: NewPokemonService()}
	http.Handle("/pokemon/", &h)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
