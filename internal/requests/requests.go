package requests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Notifies telegram bot with link for the current day
func NotifyTelegramBot(newMessages string, isGokrazy bool, err error) error {
	var link string
	url := "http://192.168.30.90:8000/gmah"

	if isGokrazy {
		link = "http://192.168.30.12:9090/dump/" + time.Now().Format("2006-01-02") + "_serve.html"
	} else {
		link = "http://localhost:9090/dump/" + time.Now().Format("2006-01-02") + "_serve.html"
	}

	newDay := struct {
		Lookup string `json:"lookup"`
		Date   string `json:"date"`
		Link   string `json:"link"`
		Count  string `json:"count"`
		Error  error
	}{
		Lookup: "false",
		Date:   time.Now().Format("2006-01-02"),
		Link:   link,
		Count:  newMessages,
		Error:  err,
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

// Notifies telegram bot about a specific date lookup
// send the error in lookUp anonymous struct if any
func NotifyTelegramBotAboutSpecificLookup(wantFileName string, date string, err error) error {
	url := "http://192.168.30.90:8000/gmah"
	link := "http://192.168.30.12:9090/dump/" + wantFileName

	lookUp := struct {
		Lookup string `json:"lookup"`
		Date   string `json:"date"`
		Link   string `json:"link"`
		Count  string `json:"count"`
		Error  error
	}{
		Lookup: "true",
		Date:   date,
		Link:   link,
		Count:  "",
		Error:  err,
	}

	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(&lookUp)

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
