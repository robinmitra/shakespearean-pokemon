package shakespeare

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCanFetchTranslation(t *testing.T) {
	res := `
	  {
		"contents": {
			"text": "Lorem ipsum",
			"translated": "Lorem ipsum dolor"
	    }
	  }
	`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, res); err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	service := NewService()
	service.BaseUrl = ts.URL

	translated, err := service.Translate("Lorem ipsum")
	if err != nil {
		t.Fatal(err)
	}

	if translated != "Lorem ipsum dolor" {
		t.Errorf("Expected translation to be '%s', found '%s'", "Lorem ipsum dolor", translated)
	}
}
