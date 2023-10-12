package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Starts new oauth2.Config from a JSON file
func configFromJSON(byteFile []byte) (*oauth2.Config, error) {
	config, err := google.ConfigFromJSON(byteFile, gmail.MailGoogleComScope)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// Retrieves a token from a local file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Request a token from the web, then returns the retrieved token
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	var authCode string

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Open this link and type the "+"authorization code: \n%v\n", authURL)

	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web: %v", err)
	}

	return tok, nil
}

// Saves a token to a file path
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

	return nil
}

// Retrieve a token, saves the token, then returns the generated client
func NewClient(byteFile []byte) (*http.Client, error) {
	config, err := configFromJSON(byteFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	tokFile := "/home/brun0/Desktop/personal/gmah/token.json"
	tok, err := tokenFromFile(tokFile)

	if err != nil {
		if tok, err = getTokenFromWeb(config); err != nil {
			return nil, err
		}
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
}
