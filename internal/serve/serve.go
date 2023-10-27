package serve

import (
	"bufio"
	"bytes"
	"os"
	"text/template"
	"time"

	"github.com/BrunoTeixeira1996/gmah/internal/email"
)

type Serve struct {
	Date   string
	Emails []email.EmailTemplate
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

func CreateHTMLFile(emails []email.EmailTemplate, htmlLocation string, isGokrazy bool) error {
	serve := &Serve{}
	// FIXME: This is breaking gokrazy conf
	tmpl := "serve_template.html"
	if isGokrazy {
		tmpl = "/perm/home/gmah/serve_template.html"
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
