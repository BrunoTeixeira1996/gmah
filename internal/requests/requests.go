package requests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func NotifyTelegramBot(newMessages string, isGokrazy bool) error {
	var link string
	url := "http://192.168.30.171:8000/gmah"

	if isGokrazy {
		link = "http://192.168.30.12:9090/dump/" + time.Now().Format("2006-01-02") + "_serve.html"
	} else {
		link = "http://localhost:9090/dump/" + time.Now().Format("2006-01-02") + "_serve.html"
	}

	newDay := struct {
		Date  string `json:"date"`
		Link  string `json:"link"`
		Count string `json:"count"`
	}{
		Date:  time.Now().Format("2006-01-02"),
		Link:  link,
		Count: newMessages,
	}

	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(&newDay)

	r, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Println("Got 500 status while notifying telegram bot")
	}

	return nil
}
