package email

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type EmailTemplate struct {
	From    string
	Subject string
	Snippet string
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
func ExtractLinks(html string, hrefLink string, hrefSlice *[]string) error {
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
func GetLinkFromSource(source string, html string, hrefSlice *[]string) error {
	switch source {
	case "idealista":
		if err := ExtractLinks(html, "https://www.idealista.pt/imovel", hrefSlice); err != nil {
			return err
		}

	case "SUPERCASA":
		if err := ExtractLinks(html, "https://supercasa.pt/venda", hrefSlice); err != nil {
			return err
		}

	case "Casa Sapo":
		if err := ExtractLinks(html, "https://casa.sapo.pt/detalhes", hrefSlice); err != nil {
			return err
		}
	case "Imovirtual":
		if err := ExtractLinks(html, "anuncio", hrefSlice); err != nil {
			return err
		}
	case "CasaYes":
		if err := ExtractLinks(html, "1818X.trk.elasticemail.com", hrefSlice); err != nil {
			return err
		}
	}
	return nil
}

// Function that extracts the snippet from the HTML itself
// it cuts from the startCut until the finalCut and grab the content of a tag inside that cut
func ExtractSnippet(html string, startCut string, finalCut string, tag string, source string) (string, error) {
	var cleanedString string

	sc := strings.Split(html, startCut)
	//This is giving warnings and i am still getting the house
	// so I decided to send error as nil to proceed
	if len(sc) < 2 {
		return "", nil
	}

	fc := strings.Split(sc[1], finalCut)
	if len(fc) < 1 {
		return "", fmt.Errorf("Error finalCut string not found in HTML")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fc[0]))
	if err != nil {
		return "", err
	}
	output := doc.Find(tag).Text()

	switch source {
	case "idealista":
		cleanedString = strings.TrimSpace(regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(strings.Split(output, "€")[0], "")) + "€"
	case "SUPERCASA":
		cleanedString = strings.TrimSpace(regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(output, ""))
	case "Casa Sapo":
		paraRegex := regexp.MustCompile(`Para: (Venda|Arrendar)`)
		precoRegex := regexp.MustCompile(`Preço: (.*?€)`)
		estadoRegex := regexp.MustCompile(`Estado: (\w+\s*\w*)`)

		paraMatches := paraRegex.FindStringSubmatch(output)
		precoMatches := precoRegex.FindStringSubmatch(output)
		estadoMatches := estadoRegex.FindStringSubmatch(output)
		cleanedString = fmt.Sprintf("Para: %s - Preço: %s - Estado: %s", paraMatches[1], precoMatches[1], estadoMatches[1])
	case "Imovirtual":
		cleanedString = strings.TrimSpace(regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(output, ""))
	case "CasaYes":
		cleanedString = strings.TrimSpace(regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(output, ""))
	}

	if cleanedString == "" {
		return "", fmt.Errorf("Error while cleaning the string, looks like its empty")
	}

	return cleanedString, nil
}

// Function that returns a small snippet about the email
func GetSnippetFromSource(source string, html string, snippet *string) error {
	var err error
	switch source {
	case "idealista":
		*snippet, err = ExtractSnippet(html, "<!-- preheader - description mail -->", "<!-- header -->", "span", source)
		if err != nil {
			return err
		}
	case "SUPERCASA":
		*snippet, err = ExtractSnippet(html, "<!-- Pre-header -->", "<!-- End region Pre-header -->", "div", source)
		if err != nil {
			return err
		}
	case "Casa Sapo":
		*snippet, err = ExtractSnippet(html, "font-size: 13px; color: #777777; font-family: Arial, Helvetica, sans-serif; padding: 2px 0;", "text-align: center; margin: 0 0 30px 0; font-family: Arial, Helvetica, sans-serif", "span", source)
		if err != nil {
			return err
		}
	case "Imovirtual":
		*snippet, err = ExtractSnippet(html, `<td style="padding-bottom: 8px;">`, "</td>", "h2", source)
		if err != nil {
			return err
		}
	case "CasaYes":
		*snippet, err = ExtractSnippet(html, "p style=color:#111317;line-height:27px;margin:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:260px", "p style=color:#576075;line-height:27px;margin:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:260px", "b", source)
	}

	return nil
}

// NormalizeSnippet collapses multiple spaces into one
func NormalizeSnippet(snippet string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(snippet, " ")
}

// Function to process the email body and extract links and snippets
func ProcessEmailBody(from string, body string) (EmailTemplate, error) {
	var email EmailTemplate
	var hrefSlice []string
	var snippet string

	// Extract links from the body
	if err := GetLinkFromSource(from, body, &hrefSlice); err != nil {
		log.Println(fmt.Errorf("Error while getLinkFromSource: %v", err))
	} else if len(hrefSlice) > 0 {
		// CasaYes for some reason uses the second link
		// to be the valid link for the house
		if from == "CasaYes" {
			email.Link = hrefSlice[1]
		} else {
			email.Link = hrefSlice[0]
		}
	}

	// Extract snippet from the body
	if err := GetSnippetFromSource(from, body, &snippet); err != nil {
		log.Printf("Error while getting Snippet in %s : %v\n", from, err)
	} else {
		email.Snippet = NormalizeSnippet(snippet)
	}

	return email, nil
}

// Function that generates the final slice to place inside the HTML template
func buildEmail(messages chan *imap.Message, section *imap.BodySectionName, newMessages *int) ([]EmailTemplate, error) {
	var emails []EmailTemplate

	for message := range messages {
		*(newMessages) += 1

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

		header := mr.Header

		var email EmailTemplate
		if from, err := header.AddressList("From"); err == nil {
			email.From = from[0].Name
		}
		if subject, err := header.Subject(); err == nil {
			email.Subject = subject
		}

		// Workaround for unwanted emails
		if email.Subject != "Novos anúncios hoje" && email.Subject != "Imóveis da mediadora Loben" {
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

				// Process the email body
				processedEmail, err := ProcessEmailBody(email.From, string(b))
				if err != nil {
					log.Printf("Error processing email body: %v", err)
					continue
				}
				email.Link = processedEmail.Link
				email.Snippet = processedEmail.Snippet
			}
			emails = append(emails, email)
		}
	}

	return emails, nil
}

// Main function that performs all the necessary logic to read and build emails
func ReadEmails(isDebug bool, email string, password string, newMessages *int) ([]EmailTemplate, error) {
	c, err := initClient()
	if err != nil {
		return []EmailTemplate{}, err
	}
	defer c.Close()

	if err := loginClient(c, email, password); err != nil {
		return []EmailTemplate{}, err
	}

	var mbox *imap.MailboxStatus

	if isDebug {
		mbox, err = c.Select("teste", false)
	} else {
		mbox, err = c.Select("Casas", false)
	}

	if err != nil {
		return []EmailTemplate{}, err
	}

	if mbox.Messages == 0 {
		log.Println("No messages in Casas so skipping ...")
		return []EmailTemplate{}, fmt.Errorf("No messages in Casas so skipping ...")
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
	var emails []EmailTemplate
	//isErr := false
	go func() {
		if err = c.Fetch(seqset, items, messages); err != nil {
			log.Println("Fetch err:", err)
		}
	}()

	emails, err = buildEmail(messages, section, newMessages)
	if err != nil {
		return []EmailTemplate{}, err
	}

	return emails, err
}
