package pokemon

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCanFetchPokemon(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res string
		switch r.URL.Path {
		case "/bulbasaur":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
			w.Header().Set("Content-Type", "text/plain")
			res = "Not Found"
		}
		if _, err := fmt.Fprint(w, res); err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	service := NewService()
	service.BaseUrl = ts.URL

	t.Run("when Pokemon exists", func(t *testing.T) {
		p, err := service.Get("charizard")
		if err != nil {
			t.Fatal(err)
		}

		if p.Name != "charizard" {
			t.Errorf("Expected Pokemon name to be 'charizard', found '%v'", p.Name)
		}
		if p.Description != "Blah blah" {
			t.Errorf("Expected Pokemon description to be 'Blah blah', found '%v'", p.Description)
		}
	})

	t.Run("when Pokemon does not exist", func(t *testing.T) {
		p, err := service.Get("superman")
		if err != nil {
			t.Fatal(err)
		}

		if p != nil {
			t.Errorf("Expect Pokemon to be nil, found %v", p)
		}
	})

	t.Run("when Pokemon exists but not in English", func(t *testing.T) {
		p, err := service.Get("bulbasaur")
		if err != nil {
			t.Fatal(err)
		}

		if p != nil {
			t.Errorf("Expect Pokemon to be nil, found %v", p)
		}
	})
}
