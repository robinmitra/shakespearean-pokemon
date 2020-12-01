package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockPokemonService struct {
	responses map[string]*Pokemon
}

func (p *mockPokemonService) Get(name string) (*Pokemon, error) {
	if pokemon, ok := p.responses[name]; ok {
		return pokemon, nil
	}
	return nil, nil
}

func TestCanHandlePokemonRequest(t *testing.T) {
	t.Run("when pokemon exists", func(t *testing.T) {
		charizard := Pokemon{"Charizard", "Blah blah blah"}
		mock := mockPokemonService{responses: map[string]*Pokemon{"charizard": &charizard}}
		handler := PokemonHandler{&mock}

		req := httptest.NewRequest(http.MethodGet, "/pokemon/charizard", nil)
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		var pokemon Pokemon
		if err := json.NewDecoder(res.Body).Decode(&pokemon); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(pokemon, charizard) {
			t.Errorf("Expected Pokemon to be Charizard")
		}

		assertContentType(t, res)
		assertStatusCode(t, res.Code, http.StatusOK)
	})

	t.Run("when pokemon does not exist", func(t *testing.T) {
		handler := PokemonHandler{&mockPokemonService{}}

		req := httptest.NewRequest(http.MethodGet, "/pokemon/batman", nil)
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		var httpErr HttpError
		if err := json.NewDecoder(res.Body).Decode(&httpErr); err != nil {
			t.Fatal(err)
		}

		assertContentType(t, res)
		assertStatusCode(t, res.Code, http.StatusNotFound)
	})

	t.Run("when path is invalid", func(t *testing.T) {
		handler := PokemonHandler{&mockPokemonService{}}

		req := httptest.NewRequest(http.MethodGet, "/pokemon/123/456", nil)
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		var httpErr HttpError
		if err := json.NewDecoder(res.Body).Decode(&httpErr); err != nil {
			t.Fatal(err)
		}

		assertContentType(t, res)
		assertStatusCode(t, res.Code, http.StatusBadRequest)
	})
}

func assertContentType(t *testing.T, res *httptest.ResponseRecorder) {
	t.Helper()
	expected := "application/json"
	found := res.Result().Header.Get("content-type")
	if found != expected {
		t.Errorf("Expected content type to be '%v', found '%v'", expected, found)
	}
}

func assertStatusCode(t *testing.T, actual, expected int) {
	t.Helper()
	if actual != expected {
		t.Errorf("Expected status to be %d, fount %d", expected, actual)
	}
}

func TestCanFetchPokemon(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		var res string
		switch r.URL.Path {
		case "/bulbasaur":
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			res = `
			  {
				"flavor_text_entries": [
				  {
					"flavor_text": "Lorem ipsum",
					"language": { "name": "lo" }
				  }
				]
			  }
			`

		case "/charizard":
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			res = `
			  {
				"flavor_text_entries": [
				  {
					"flavor_text": "Blah blah",
					"language": { "name": "en" }
				  },
				  {
					"flavor_text": "Lorem ipsum",
					"language": { "name": "lo" }
				  }
				]
			  }
			`

		default:
			writer.Header().Set("Content-Type", "text/plain")
			res = "Not Found"
		}
		if _, err := fmt.Fprint(writer, res); err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	service := NewPokemonService()
	service.baseUrl = ts.URL

	t.Run("when Pokemon exists", func(t *testing.T) {
		pokemon, err := service.Get("charizard")
		if err != nil {
			t.Fatal(err)
		}

		if pokemon.Name != "charizard" {
			t.Errorf("Expected Pokemon name to be 'charizard', found '%v'", pokemon.Name)
		}
		if pokemon.Description != "Blah blah" {
			t.Errorf("Expected Pokemon description to be 'Blah blah', found '%v'", pokemon.Description)
		}
	})

	t.Run("when Pokemon does not exist", func(t *testing.T) {
		pokemon, err := service.Get("superman")
		if err != nil {
			t.Fatal(err)
		}

		if pokemon != nil {
			t.Errorf("Expect Pokemon to be nil, found %v", pokemon)
		}
	})

	t.Run("when Pokemon exists but not in English", func(t *testing.T) {
		pokemon, err := service.Get("bulbasaur")
		if err != nil {
			t.Fatal(err)
		}

		if pokemon != nil {
			t.Errorf("Expect Pokemon to be nil, found %v", pokemon)
		}
	})
}
