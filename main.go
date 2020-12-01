package main

import (
	"encoding/json"
	"fmt"
	"github.com/robinmitra/shakespearean-pokemon/pokemon"
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
	Get(name string) (*pokemon.Pokemon, error)
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
	h := PokemonHandler{pokemonService: pokemon.NewService()}
	http.Handle("/pokemon/", &h)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
