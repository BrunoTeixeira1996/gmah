package queries

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mnako/letters"
	"google.golang.org/api/gmail/v1"
)

type MailMetadata struct {
	Sender  string
	From    string
	Subject string
	Snipet  string
	Body    string
	HTML    string
	Link    string
}

type EmailTemplate struct {
	From    string
	Subject string
	Snipet  string
	Link    string
}

type Serve struct {
	Date   string
	Emails []*EmailTemplate
}

// Method that fills Sender, From and Subject to MailMetadata struct
func (m *MailMetadata) getMailMetadata(msg *gmail.Message) {
	for _, v := range msg.Payload.Headers {
		switch strings.ToLower(v.Name) {
		case "sender":
			m.Sender = v.Value
		case "from":
			m.From = v.Value
		case "subject":
			m.Subject = v.Value
		}
	}
}

// Method that fills body and snipet of MailMetadata struct
func (m *MailMetadata) getMailBody(srv *gmail.Service, messageId string) error {
	gmailMessageResponse, err := srv.Users.Messages.Get("me", messageId).Format("RAW").Do()
	if err != nil {
		return err
	}

	if gmailMessageResponse != nil {
		decodedData, err := base64.URLEncoding.DecodeString(gmailMessageResponse.Raw)
		if err != nil {
			return err
		}
		m.Body = string(decodedData)
		m.Snipet = strings.Split(gmailMessageResponse.Snippet, "€")[0]
	}

	return nil
}

// Function that lists all available labels
func listLabels(srv *gmail.Service) error {
	labels, err := srv.Users.Labels.List("me").Do()
	if err != nil {
		return err
	}

	if len(labels.Labels) == 0 {
		return fmt.Errorf("No labels available")
	}

	for _, l := range labels.Labels {
		fmt.Printf("- %s\n", l.Name)
	}

	return nil
}

// Function that returns messages by ID
func getByID(srv *gmail.Service, msgs *gmail.ListMessagesResponse) ([]*gmail.Message, error) {
	var msgSlice []*gmail.Message
	for _, v := range msgs.Messages {
		msg, err := srv.Users.Messages.Get("me", v.Id).Do()
		if err != nil {
			return msgSlice, err
		}
		msgSlice = append(msgSlice, msg)
	}
	return msgSlice, nil
}

// Function that extract links using goquery
func extractLinks(md *MailMetadata, hrefLink string, hrefSlice *[]string) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(md.HTML))
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
func getLinkFromSource(source string, md *MailMetadata, hrefSlice *[]string) error {
	switch source {
	case "idealista":
		if err := extractLinks(md, "imovel", hrefSlice); err != nil {
			return err
		}

	case "supercasa":
		if err := extractLinks(md, "https://supercasa.pt/venda", hrefSlice); err != nil {
			return err
		}
	default:
		(*hrefSlice)[0] = "test"
	}

	return nil
}

// Func to mark message as read by removing UNREAD label
func markAsRead(srv *gmail.Service, msg *gmail.Message) error {
	req := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}

	// Request to remove the "UNREAD" label thus maring it as "READ"
	if _, err := srv.Users.Messages.Modify("me", msg.Id, req).Do(); err != nil {
		return err
	}
	return nil
}

// Func that writes a template to a HTML file
func writeTemplateToFile(outputPath string, outTemp bytes.Buffer) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	if _, err := w.WriteString(string(outTemp.Bytes())); err != nil {
		return err
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

func (e *EmailTemplate) BuildEmail(msg *MailMetadata) {
	e.From = msg.From
	e.Subject = msg.Subject
	e.Snipet = msg.Snipet
	e.Link = msg.Link
}

func CreateHTMLFile(emails []*EmailTemplate, htmlLocation string, isGokrazy bool) error {
	serve := &Serve{}
	// FIXME: This is breaking gokrazy conf
	tmpl := "serve_template.html"
	if isGokrazy {
		tmpl = "/perm/gmah/serve_template.html"
	}
	templ, err := template.New("serve_template.html").ParseFiles(tmpl)
	if err != nil {
		return err
	}

	var outTemp bytes.Buffer
	// This is the struct that is written in the html template
	serve.Date = time.Now().Format("2006-01-02")
	serve.Emails = emails
	if err := templ.Execute(&outTemp, serve); err != nil {
		return err
	}

	fileName := time.Now().Format("2006-01-02") + "_serve.html"
	outputPath := htmlLocation + fileName
	if err := writeTemplateToFile(outputPath, outTemp); err != nil {
		return err
	}

	return nil
}

// Function that gets unread messages in label Casas
func GetMessages(srv *gmail.Service, newMessages *int) ([]*MailMetadata, error) {
	var mailsMetadata []*MailMetadata

	// Get the messages metadata
	inbox, err := srv.Users.Messages.List("me").Q("Casas is:unread").Do()
	if err != nil {
		return nil, err
	}
	log.Println("Got new unread messages")

	msgs, err := getByID(srv, inbox)
	if err != nil {
		return nil, err
	}

	for _, msg := range msgs {
		var hrefSlice []string
		md := &MailMetadata{}
		md.getMailMetadata(msg)
		if err := md.getMailBody(srv, msg.Id); err != nil {
			return nil, err
		}
		*newMessages += 1

		reader := strings.NewReader(md.Body)
		email, err := letters.ParseEmail(reader)
		if err != nil {
			return nil, err
		}
		md.HTML = email.HTML

		if err := getLinkFromSource(strings.ToLower(strings.Split(md.From, " ")[0]), md, &hrefSlice); err != nil {
			return nil, err
		}
		md.Link = hrefSlice[0]
		mailsMetadata = append(mailsMetadata, md)

		// comment for debug
		if err := markAsRead(srv, msg); err != nil {
			return nil, err
		}
	}
	log.Printf("Got %d unread emails\n", *newMessages)

	return mailsMetadata, nil
}
