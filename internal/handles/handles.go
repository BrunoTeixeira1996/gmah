package handles

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/BrunoTeixeira1996/gmah/internal/requests"
)

// Handles POST to lookup for a specific date
func LookUpSpecificHandle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != "POST" {
		http.Error(w, "NOT POST!", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	temp := struct {
		Date string
	}{}

	if err := decoder.Decode(&temp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error while unmarshal json response:", err)
		return
	}
	tempD := strings.Split(temp.Date, "/")
	wantFileName := fmt.Sprintf("%s-%s-%s_serve.html", tempD[2], tempD[1], tempD[0])

	// Check if html and link already exist on /perm/home/gmah/html path
	path := "/perm/home/gmah/html"
	files, err := os.ReadDir(path)
	if err != nil {
		log.Println("Error while reading the perm data of gokrazy:", err)
		return
	}
	for _, file := range files {
		if wantFileName == file.Name() {
			// Notify bot to send link
			requests.NotifyTelegramBotAboutSpecificLookup(wantFileName, temp.Date, nil)
			return
		}
	}
	err = fmt.Errorf("Could not find any find from that date")
	log.Println(err)
	requests.NotifyTelegramBotAboutSpecificLookup("", temp.Date, err)
}

// Handles "/"
func IndexHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage")
}
