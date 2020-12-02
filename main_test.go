package main

import (
	"encoding/json"
	"errors"
	"github.com/robinmitra/shakespearean-pokemon/pokemon"
	"net/http"
	"net/http/httptest"
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

type mockShakespeareService struct {
	responses map[string]string
}

func (s *mockShakespeareService) Translate(text string) (string, error) {
	if t, ok := s.responses[text]; ok {
		return t, nil
	}
	return "", errors.New("some error")
}

func TestCanHandlePokemonRequest(t *testing.T) {
	t.Run("when pokemon exists", func(t *testing.T) {
		charizard := pokemon.Pokemon{Name: "Charizard", Description: "Blah"}
		mPokemonService := mockPokemonService{responses: map[string]*pokemon.Pokemon{"charizard": &charizard}}

		translation := "Blah cough"
		mShakespeareService := mockShakespeareService{responses: map[string]string{"Blah": translation}}

		handler := PokemonHandler{&mPokemonService, &mShakespeareService}

		req := httptest.NewRequest(http.MethodGet, "/pokemon/charizard", nil)
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		var pr PokemonResponse

		if err := json.NewDecoder(res.Body).Decode(&pr); err != nil {
			t.Fatal(err)
		}

		if pr.Name != charizard.Name {
			t.Errorf("Expected Pokemon name to be to be '%s', found '%s'", charizard.Name, pr.Name)
		}

		if pr.Description != translation {
			t.Errorf("Expected translation to be '%s', found '%s'", translation, pr.Description)
		}

		assertContentType(t, res)
		assertStatusCode(t, res.Code, http.StatusOK)
	})

	t.Run("when pokemon does not exist", func(t *testing.T) {
		handler := PokemonHandler{&mockPokemonService{}, &mockShakespeareService{}}

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
		handler := PokemonHandler{&mockPokemonService{}, &mockShakespeareService{}}

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

	t.Run("when translation fails", func(t *testing.T) {
		charizard := pokemon.Pokemon{Name: "Charizard", Description: "Blah"}
		mPokemonService := mockPokemonService{responses: map[string]*pokemon.Pokemon{"charizard": &charizard}}

		handler := PokemonHandler{&mPokemonService, &mockShakespeareService{}}

		req := httptest.NewRequest(http.MethodGet, "/pokemon/charizard", nil)
		res := httptest.NewRecorder()

		handler.ServeHTTP(res, req)

		var httpErr HttpError
		if err := json.NewDecoder(res.Body).Decode(&httpErr); err != nil {
			t.Fatal(err)
		}

		assertContentType(t, res)
		assertStatusCode(t, res.Code, http.StatusInternalServerError)
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
