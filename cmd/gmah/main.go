package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/BrunoTeixeira1996/gmah/internal/email"
	"github.com/BrunoTeixeira1996/gmah/internal/handles"
	"github.com/BrunoTeixeira1996/gmah/internal/requests"
	"github.com/BrunoTeixeira1996/gmah/internal/serve"
	"github.com/go-co-op/gocron"

	cp "github.com/otiai10/copy"
)

// Cronjob to check new email
// this executes once a day
func getNewEmailsCronJob(emailFlag string, passwordFlag string, dump string, newMessages *int, isGokrazy bool) {
	c := gocron.NewScheduler(time.UTC)
	// c.Cron("* * * * *") // every minute
	c.Cron("0 17 * * *").Do(func() {
		var (
			emails      []email.EmailTemplate
			newMessages int
			err         error
		)

		log.Printf("Executing cronjob %s ...\n", time.Now().String())

		if emails, err = email.ReadEmails(emailFlag, passwordFlag, &newMessages); err != nil {
			log.Println("Error while performing the read emails inside the cronjob: ", err.Error())
		}

		log.Println("ReadEmails output err:", err)

		if err = serve.CreateHTMLFile(emails, dump, isGokrazy); err != nil {
			log.Println("Error while creating html file: ", err.Error())
		}

		log.Println("CreateHTMLFile output err:", err)

		newMessagesStr := strconv.Itoa(newMessages)
		log.Printf("Got %s new messages\n", newMessagesStr)

		// Notifies telegram
		if err := requests.NotifyTelegramBot(newMessagesStr, isGokrazy); err != nil {
			log.Println("Error while notifying telegram bot: " + err.Error())
		}

		log.Println("NotifyTelegramBot output err:", err)

		log.Printf("Finished cronjob %s\n", time.Now().String())
	})

	c.StartAsync()
}

func getNewEmails(emailFlag string, passwordFlag string, dump string, newMessages *int, isGokrazy bool) {
	if _, err := email.ReadEmails(emailFlag, passwordFlag, newMessages); err != nil {
		log.Println("Error while performing the read emails inside the cronjob: ", err.Error())
	}
}

func handleExit(exit chan bool) {
	ch := make(chan os.Signal, 5)
	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("Closing web server")
	exit <- true
}

func startServer(dumpFlag string) error {
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

func logic() error {
	var emailFlag = flag.String("email", "", "-email='youremail@mail.com'")
	var passwordFlag = flag.String("password", "", "-password='yourpassword'")
	var gokrazyFlag = flag.Bool("gokrazy", false, "use this if you are using gokrazy")
	var dumpFlag = flag.String("dump", "", "-dump='/path/html/'")
	var isDebugFlag = flag.Bool("debug", false, "use this if in debug mode")
	flag.Parse()

	if *emailFlag == "" || *passwordFlag == "" {
		return fmt.Errorf("Please provide the email and password")
	}

	if *gokrazyFlag {
		log.Println("OK lets do this on gokrazy then ...")
		// copy required folders to /pem
		errTemplate := cp.Copy("/etc/gmah/serve_template.html", "/perm/home/gmah/serve_template.html")
		errHtml := cp.Copy("/etc/gmah/html", "/perm/home/gmah/html")

		if errTemplate != nil || errHtml != nil {
			return fmt.Errorf("Error while copying files and folders to gokrazy perm (template:%v),(html:%v)", errTemplate, errHtml)
		}
	}

	// Cronjob to check new emails per day
	var newMessages int
	if *isDebugFlag {
		getNewEmails(*emailFlag, *passwordFlag, *dumpFlag, &newMessages, *gokrazyFlag)
	} else {
		getNewEmailsCronJob(*emailFlag, *passwordFlag, *dumpFlag, &newMessages, *gokrazyFlag)
	}

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
