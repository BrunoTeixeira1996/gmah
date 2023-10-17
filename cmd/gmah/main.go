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
	"time"

	"github.com/BrunoTeixeira1996/gmah/internal/auth"
	"github.com/BrunoTeixeira1996/gmah/internal/handles"
	"github.com/BrunoTeixeira1996/gmah/internal/queries"
	"github.com/BrunoTeixeira1996/gmah/internal/requests"
	"github.com/go-co-op/gocron"
	"google.golang.org/api/gmail/v1"

	cp "github.com/otiai10/copy"
	"google.golang.org/api/option"
)

// Cronjob to check new email
// this executes once a day
func getNewEmailsCronJob(clientSecret string, tokFile string, dump string, newMessages *int, isGokrazy bool) {
	c := gocron.NewScheduler(time.UTC)
	// c.Cron("* * * * *") // every minute
	c.Cron("0 17 * * *").Do(func() {
		if err := readEmails(clientSecret, tokFile, dump, newMessages, isGokrazy); err != nil {
			log.Println("Error while performing the read emails inside the cronjob: " + err.Error())
		}
		newMessagesStr := strconv.Itoa(*newMessages)
		// Notifies telegram
		if err := requests.NotifyTelegramBot(newMessagesStr, isGokrazy); err != nil {
			log.Println("Error while notifying telegram bot: " + err.Error())
		}
	})

	c.StartAsync()
}

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
func readEmails(clientSecret string, tokFile string, dumpLocation string, newMessages *int, isGokrazy bool) error {
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
	if err := queries.CreateHTMLFile(emails, dumpLocation, isGokrazy); err != nil {
		return err
	}

	return nil
}

func logic() error {
	var gokrazyFlag = flag.Bool("gokrazy", false, "use this if you are using gokrazy")
	var clientSecretFlag = flag.String("client-secret", "", "-client-secret='/path/client_secret.json'")
	var tokFileFlag = flag.String("token-file", "", "-token-fike='/path/token.json'")
	var dumpFlag = flag.String("dump", "", "-dump='/path/html/'")
	flag.Parse()

	if *gokrazyFlag {
		log.Println("OK lets do this on gokrazy then ...")
		// copy required folders to /pem
		errTemplate := cp.Copy("/etc/gmah/serve_template.html", "/perm/home/gmah/serve_template.html")
		errHtml := cp.Copy("/etc/gmah/html", "/perm/home/gmah/html")

		if errTemplate != nil || errHtml != nil {
			return fmt.Errorf("Error while copying files and folders to gokrazy perm (template:%v),(html:%v)", errTemplate, errHtml)
		}
	}

	if *clientSecretFlag == "" || *tokFileFlag == "" || *dumpFlag == "" {
		return fmt.Errorf("Did not provided client_secret.json or token.json or the html folder to dump html files")
	}

	// Cronjob to check new emails per day
	var newMessages int
	getNewEmailsCronJob(*clientSecretFlag, *tokFileFlag, *dumpFlag, &newMessages, *gokrazyFlag)

	// Starts webserver
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
