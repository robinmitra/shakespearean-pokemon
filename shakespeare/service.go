package shakespeare

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type Service struct {
	BaseUrl string
}

func NewService() *Service {
	return &Service{BaseUrl: "https://api.funtranslations.com/translate/shakespeare.json"}
}

func (s *Service) Translate(text string) (string, error) {
	req, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return "", err
	}
	res, err := http.Post(s.BaseUrl, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return "", err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	type ResJson struct {
		Contents struct {
			Text       string
			Translated string
		}
	}
	var resJson ResJson
	if err := json.NewDecoder(res.Body).Decode(&resJson); err != nil {
		return "", err
	}

	return resJson.Contents.Translated, nil
}
