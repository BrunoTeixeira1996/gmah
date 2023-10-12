package queries

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mnako/letters"
	"google.golang.org/api/gmail/v1"
)

type MailMetadata struct {
	// Sender is the entity that originally created and sent the message
	Sender string
	// From is the entity that sent the message to you (e.g. googlegroups). Most
	// of the time this information is only relevant to mailing lists.
	From string
	// Subject is the email subject
	Subject string
	// Snipet is a snippet of the email body
	Snipet string
	// Body is the email body
	Body string
	// HTML is the HTML code of the email
	HTML string
	// Link is the link of the property
	Link string
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

// TODO: The mark as read works, but now I need to make a better print statement and then
// use that in an html template and every day I create an html file and send the file
// from the telegram bot to me https://github.com/BrunoTeixeira1996/gbackup/blob/master/internal/email.go#L39

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

// Function that gets unread messages in label Casas
func GetMessages(srv *gmail.Service) error {
	// Get the messages metadata
	inbox, err := srv.Users.Messages.List("me").Q("Casas is:unread").Do()
	if err != nil {
		return err
	}

	msgs, err := getByID(srv, inbox)
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		var hrefSlice []string
		md := &MailMetadata{}
		md.getMailMetadata(msg)
		if err := md.getMailBody(srv, msg.Id); err != nil {
			return err
		}

		reader := strings.NewReader(md.Body)
		email, err := letters.ParseEmail(reader)
		if err != nil {
			return err
		}
		md.HTML = email.HTML

		if err := getLinkFromSource(strings.ToLower(strings.Split(md.From, " ")[0]), md, &hrefSlice); err != nil {
			return err
		}
		md.Link = hrefSlice[0]
		//fmt.Printf("------\nFrom:%s\nSubject:%s\nSnipet:%s\nLink:%s\n------\n", md.From, md.Subject, md.Snipet, md.Link)
		if err := markAsRead(srv, msg); err != nil {
			return err
		}

		fmt.Println("marked as read")

	}

	return nil
}
