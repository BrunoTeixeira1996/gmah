package main

import (
	"context"
	"fmt"
	"os"

	"github.com/BrunoTeixeira1996/gmah/internal/auth"
	"github.com/BrunoTeixeira1996/gmah/internal/queries"
	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func logic() error {
	ctx := context.Background()
	// FIXME: add flag to client_secret.json path
	byteFile, err := os.ReadFile("/home/brun0/Sync/gmail_tokens/client_secret.json")
	if err != nil {
		return err
	}

	client, err := auth.NewClient(byteFile)
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
	if err := queries.CreateHTMLFile(emails); err != nil {
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