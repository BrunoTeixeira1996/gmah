package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/BrunoTeixeira1996/gmah/internal/auth"
	"github.com/BrunoTeixeira1996/gmah/internal/handles"
	"github.com/BrunoTeixeira1996/gmah/internal/queries"
	"github.com/BrunoTeixeira1996/gmah/internal/requests"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func handleExit(exit chan bool) {
	ch := make(chan os.Signal, 5)
	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("Closing web server")
	exit <- true
}

func startServer(dumpFlag string) error {
	// HTTP Server
	// Handle exit
	exit := make(chan bool)
	go handleExit(exit)

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(dumpFlag))
	mux.Handle("/dump/", http.StripPrefix("/dump/", fs))
	mux.HandleFunc("/", handles.IndexHandle)

	go func() {
		if err := http.ListenAndServe(":9090", mux); err != nil && err != http.ErrServerClosed {
			panic("Error trying to start http server: " + err.Error())
		}
	}()

	log.Println("Serving at :9090")
	<-exit

	return nil
}

// Func that performs all operations related to email (read and mark as read)
func readEmails(clientSecret string, tokFile string, dumpLocation string, newMessages *int) error {
	ctx := context.Background()

	byteFile, err := os.ReadFile(clientSecret)
	if err != nil {
		return err
	}

	client, err := auth.NewClient(byteFile, tokFile)
	if err != nil {
		return err
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("Unable to retrieve Gmail client: %v", err)
	}

	mailsMetadata, err := queries.GetMessages(srv, newMessages)
	if err != nil {
		return err
	}

	// Gets all emails and assigns to a new struct
	// in order to create a new template in the future
	emails := make([]*queries.EmailTemplate, 0)
	for _, metadata := range mailsMetadata {
		email := &queries.EmailTemplate{}
		email.BuildEmail(metadata)
		emails = append(emails, email)
	}

	// Creates an HTML file from the emails slice
	if err := queries.CreateHTMLFile(emails, dumpLocation); err != nil {
		return err
	}

	return nil
}

func logic() error {
	var clientSecretFlag = flag.String("client-secret", "", "-client-secret='/path/client_secret.json'")
	var tokFileFlag = flag.String("token-file", "", "-token-fike='/path/token.json'")
	var dumpFlag = flag.String("dump", "", "-dump='/path/html/'")
	flag.Parse()

	if *clientSecretFlag == "" || *tokFileFlag == "" || *dumpFlag == "" {
		return fmt.Errorf("Did not provided client_secret.json or token.json or the html folder to dump html files")
	}

	// TODO: schedule this once per day
	var newMessages int
	if err := readEmails(*clientSecretFlag, *tokFileFlag, *dumpFlag, &newMessages); err != nil {
		return err
	}

	newMessagesStr := strconv.Itoa(newMessages)
	if err := requests.NotifyTelegramBot(newMessagesStr); err != nil {
		return err
	}

	if err := startServer(*dumpFlag); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := logic(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
