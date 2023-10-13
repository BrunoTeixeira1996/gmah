package requests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func NotifyTelegramBot(newMessages string) error {
	url := "http://192.168.30.171:8000/gmah"

	newDay := struct {
		Date  string `json:"date"`
		Link  string `json:"link"`
		Count string `json:"count"`
	}{
		Date: time.Now().Format("2006-01-02"),
		// FIXME: change from localhost to brun0-pi when in prod
		//		Link:  "http://brun0-pi:9090/gmah/dump/" + time.Now().Format("2006-01-02") + "_serve.html",
		Link:  "http://localhost:9090/dump/" + time.Now().Format("2006-01-02") + "_serve.html",
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
