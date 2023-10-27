package email

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type EmailTemplate struct {
	From    string
	Subject string
	Snipet  string
	Link    string
}

func initClient() (*client.Client, error) {
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func loginClient(c *client.Client, email string, password string) error {
	if err := c.Login(email, password); err != nil {
		return err
	}
	return nil
}

// Function that extract links using goquery
func extractLinks(html string, hrefLink string, hrefSlice *[]string) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return err
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if strings.Contains(href, hrefLink) {
			*hrefSlice = append(*hrefSlice, href)
		}
	})

	return nil
}

// Function that returns link depending on the source (supercasa, idealista, etc)
func getLinkFromSource(source string, html string, hrefSlice *[]string) error {
	switch source {
	case "idealista":
		if err := extractLinks(html, "imovel", hrefSlice); err != nil {
			return err
		}

	case "SUPERCASA":
		if err := extractLinks(html, "https://supercasa.pt/venda", hrefSlice); err != nil {
			return err
		}
	}
	return nil
}

// Function that generates the final slice to place inside the HTML template
// FIXME : Get Snipet
func buildEmail(messages chan *imap.Message, section *imap.BodySectionName) ([]EmailTemplate, error) {
	var emails []EmailTemplate

	for message := range messages {
		var hrefSlice []string

		if message == nil {
			return []EmailTemplate{}, fmt.Errorf("Server didn't returned message")
		}
		r := message.GetBody(section)
		if r == nil {
			return []EmailTemplate{}, fmt.Errorf("Server didn't returned message body")
		}

		// Create a new mail reader
		mr, err := mail.CreateReader(r)
		if err != nil {
			return []EmailTemplate{}, err
		}

		var email EmailTemplate
		header := mr.Header

		if from, err := header.AddressList("From"); err == nil {
			email.From = from[0].Name
		}
		if subject, err := header.Subject(); err == nil {
			email.Subject = subject
		}

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			b, err := io.ReadAll(p.Body)
			if err != nil {
				return []EmailTemplate{}, err
			}

			if err := getLinkFromSource(email.From, string(b), &hrefSlice); err != nil {
				return []EmailTemplate{}, err
			}

			email.Link = hrefSlice[0]
		}

		emails = append(emails, email)
	}

	return emails, nil
}

// Main function that performs all the necessary logic to read and build emails
func ReadEmails(email string, password string) ([]EmailTemplate, error) {
	c, err := initClient()
	if err != nil {
		return []EmailTemplate{}, err
	}
	defer c.Close()

	if err := loginClient(c, email, password); err != nil {
		return []EmailTemplate{}, err
	}

	mbox, err := c.Select("Casas", false)
	if err != nil {
		return []EmailTemplate{}, err
	}

	if mbox.Messages == 0 {
		log.Println("No messages in Casas so skipping ...")
		return []EmailTemplate{}, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(mbox.Messages)
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}
	uids, err := c.Search(criteria)
	if err != nil {
		return []EmailTemplate{}, err
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}
	messages := make(chan *imap.Message, 1)

	// Fetch all messages unread that are inside Casas label
	go func() {
		if err = c.Fetch(seqset, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	emails, err := buildEmail(messages, section)
	if err != nil {
		return []EmailTemplate{}, err
	}

	return emails, err
}
