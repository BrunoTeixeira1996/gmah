package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/BrunoTeixeira1996/gmah/internal/email"
	"github.com/BrunoTeixeira1996/gmah/internal/handles"
	"github.com/BrunoTeixeira1996/gmah/internal/requests"
	"github.com/BrunoTeixeira1996/gmah/internal/serve"

	cp "github.com/otiai10/copy"
)

// Handles GET to lookup on demand
func lookUpHandle(args Args, newMessages int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "NOT GET!", http.StatusBadRequest)
			return
		}
		log.Println("Running lookup on demand")
		run(args.Debug, args.Email, args.Password, args.Dump, &newMessages, args.Gokrazy)
	}
}

func run(isDebug bool, emailFlag string, passwordFlag string, dump string, newMessages *int, isGokrazy bool) {
	var (
		emails []email.EmailTemplate
		err    error
	)

	log.Printf("Executing cronjob %s ...\n", time.Now().String())

	if emails, err = email.ReadEmails(isDebug, emailFlag, passwordFlag, newMessages); err != nil {
		log.Println("Error while performing the read emails inside the cronjob: ", err.Error())
	}

	log.Println("ReadEmails output err:", err)

	if err = serve.CreateHTMLFile(emails, dump, isGokrazy); err != nil {
		log.Println("Error while creating html file: ", err.Error())
	}

	log.Println("CreateHTMLFile output err:", err)

	newMessagesStr := strconv.Itoa(*newMessages)
	log.Printf("Got %s new messages\n", newMessagesStr)

	// Notifies telegram
	if err := requests.NotifyTelegramBot(newMessagesStr, isGokrazy); err != nil {
		log.Println("Error while notifying telegram bot: " + err.Error())
	}

	// Clean newMessages pointer
	*newMessages = 0

	log.Println("NotifyTelegramBot output err:", err)

	log.Printf("Finished cronjob %s\n", time.Now().String())
}

type Args struct {
	Email    string
	Password string
	Gokrazy  bool
	Dump     string
	Debug    bool
}

func gatherFlags() (Args, error) {
	var emailFlag = flag.String("email", "", "-email='youremail@mail.com'")
	var passwordFlag = flag.String("password", "", "-password='yourpassword'")
	var gokrazyFlag = flag.Bool("gokrazy", false, "use this if you are using gokrazy")
	var dumpFlag = flag.String("dump", "", "-dump='/path/html/'")
	var debugFlag = flag.Bool("debug", false, "use this to ignore cronjob")
	flag.Parse()

	if *emailFlag == "" || *passwordFlag == "" {
		return Args{}, fmt.Errorf("Please provide the email and password")
	}

	args := Args{
		Email:    *emailFlag,
		Password: *passwordFlag,
		Gokrazy:  *gokrazyFlag,
		Dump:     *dumpFlag,
		Debug:    *debugFlag,
	}

	if *gokrazyFlag {
		log.Println("OK lets do this on gokrazy then ...")
		// copy required folders to /pem
		errTemplate := cp.Copy("/etc/gmah/serve_template.html", "/perm/home/gmah/serve_template.html")
		errHtml := cp.Copy("/etc/gmah/html", "/perm/home/gmah/html")

		if errTemplate != nil || errHtml != nil {
			return Args{}, fmt.Errorf("Error while copying files and folders to gokrazy perm (template:%v),(html:%v)", errTemplate, errHtml)
		}
	}

	return args, nil
}

func main() {
	var (
		args Args
		err  error
	)
	if args, err = gatherFlags(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	var newMessages int

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(args.Dump))
	mux.Handle("/dump/", http.StripPrefix("/dump/", fs))
	mux.HandleFunc("/", handles.IndexHandle)
	mux.HandleFunc("/lookup", lookUpHandle(args, newMessages))
	go http.ListenAndServe(":9090", mux)

	// If its debug mode then run and ignore cronjob
	if args.Debug {
		run(args.Debug, args.Email, args.Password, args.Dump, &newMessages, args.Gokrazy)
		return
	}

	// Starts cronjon
	runCh := make(chan struct{})
	go func() {
		// Run forever, trigger a run at 18:00 every day.
		for {
			now := time.Now()
			runToday := now.Hour() < 18
			today := now.Day()
			log.Printf("now = %v, runToday = %v", now, runToday)
			for {
				if time.Now().Day() != today {
					// Day changed, re-evaluate whether to run today.
					break
				}

				nextHour := time.Now().Truncate(time.Hour).Add(1 * time.Hour)
				//log.Printf("today = %d, runToday = %v, next hour: %v", today, runToday, nextHour)
				time.Sleep(time.Until(nextHour))

				if time.Now().Hour() >= 18 && runToday {
					runToday = false
					runCh <- struct{}{}
				}
			}
		}
	}()

	for range runCh {
		run(args.Debug, args.Email, args.Password, args.Dump, &newMessages, args.Gokrazy)
	}
}
