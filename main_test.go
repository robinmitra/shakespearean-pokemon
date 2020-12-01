package main

import (
	"encoding/json"
	"github.com/robinmitra/shakespearean-pokemon/pokemon"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockPokemonService struct {
	responses map[string]*pokemon.Pokemon
}

func (p *mockPokemonService) Get(name string) (*pokemon.Pokemon, error) {
	if p, ok := p.responses[name]; ok {
		return p, nil
	}
	return nil, nil
}

func TestCanHandlePokemonRequest(t *testing.T) {
	t.Run("when pokemon exists", func(t *testing.T) {
		charizard := pokemon.Pokemon{Name: "Charizard", Description: "Blah blah blah"}
		mock := mockPokemonService{responses: map[string]*pokemon.Pokemon{"charizard": &charizard}}
		handler := PokemonHandler{&mock}

		req := httptest.NewRequest(http.MethodGet, "/pokemon/charizard", nil)
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		var p pokemon.Pokemon
		if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(p, charizard) {
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
