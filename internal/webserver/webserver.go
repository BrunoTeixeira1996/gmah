package webserver

import (
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/BrunoTeixeira1996/gmah/internal/handles"
)

// Handles the exit signal
func handleExit(exit chan bool) {
	ch := make(chan os.Signal, 5)
	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("Closing web server")
	exit <- true
}

// Starts the web server
func StartServer(currentPath string, dumpPath string) error {
	// Handle exit
	exit := make(chan bool)
	go handleExit(exit)

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(dumpPath))
	mux.Handle("/dump/", http.StripPrefix("/dump/", fs))

	mux.HandleFunc("/", handles.IndexHandle)

	// HTTP Server
	go func() {
		if err := http.ListenAndServe(":9090", mux); err != nil && err != http.ErrServerClosed {
			panic("Error trying to start http server: " + err.Error())
		}
	}()

	log.Println("Serving at :9090")
	<-exit

	return nil
}
