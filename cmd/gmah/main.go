package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/BrunoTeixeira1996/gmah/internal/auth"
	"github.com/BrunoTeixeira1996/gmah/internal/queries"
	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func logic() error {
	var clientSecretFlag = flag.String("client-secret", "", "-client-secret='/path/client_secret.json'")
	var tokFileFlag = flag.String("token-file", "", "-token-fike='/path/token.json'")
	var dumpFlag = flag.String("dump", "", "-dump='/path/html/'")
	flag.Parse()

	if *clientSecretFlag == "" || *tokFileFlag == "" || *dumpFlag == "" {
		return fmt.Errorf("Did not provided client_secret.json or token.json or the html folder to dump html files")
	}

	ctx := context.Background()

	byteFile, err := os.ReadFile(*clientSecretFlag)
	if err != nil {
		return err
	}

	client, err := auth.NewClient(byteFile, *tokFileFlag)
	if err != nil {
		return err
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("Unable to retrieve Gmail client: %v", err)
	}

	mailsMetadata, err := queries.GetMessages(srv)
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
	if err := queries.CreateHTMLFile(emails, *dumpFlag); err != nil {
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
