package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
)

var FILEDROP_SMTP_HOST string
var FILEDROP_SMTP_PORT string
var FILEDROP_EMAIL_FROM string
var FILEDROP_SMTP_USERNAME string
var FILEDROP_SMTP_PASSWORD string
var FILEDROP_EMAIL_TO string

func main() {
	FILEDROP_SMTP_HOST = os.Getenv("FILEDROP_SMTP_HOST")
	if FILEDROP_SMTP_HOST == "" {
		log.Fatal("FILEDROP_SMTP_HOST is not set")
	}
	FILEDROP_SMTP_PORT = os.Getenv("FILEDROP_SMTP_PORT")
	if FILEDROP_SMTP_PORT == "" {
		log.Fatal("FILEDROP_SMTP_PORT is not set")
	}
	FILEDROP_EMAIL_FROM = os.Getenv("FILEDROP_EMAIL_FROM")
	if FILEDROP_EMAIL_FROM == "" {
		log.Fatal("FILEDROP_EMAIL_FROM is not set")
	}
	FILEDROP_SMTP_USERNAME = os.Getenv("FILEDROP_SMTP_USERNAME")
	if FILEDROP_SMTP_USERNAME == "" {
		log.Fatal("FILEDROP_SMTP_USERNAME is not set")
	}
	FILEDROP_SMTP_PASSWORD = os.Getenv("FILEDROP_SMTP_PASSWORD")
	if FILEDROP_SMTP_PASSWORD == "" {
		log.Fatal("FILEDROP_SMTP_PASSWORD is not set")
	}
	FILEDROP_EMAIL_TO = os.Getenv("FILEDROP_EMAIL_TO")
	if FILEDROP_EMAIL_TO == "" {
		log.Fatal("FILEDROP_EMAIL_TO is not set")
	}

	http.HandleFunc("/", uploadHandler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte(`
            <html>
            <body>
            <form enctype="multipart/form-data" action="/" method="post">
                <input type="file" name="file" />
                <input type="submit" value="Upload" />
            </form>
            </body>
            </html>
        `))
	} else if r.Method == "POST" {
		// Parse the multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read file content into a byte array
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fileData := buf.Bytes()

		// Send email with the file
		if err := sendEmailWithAttachment(fileData, header.Filename); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "File uploaded and sent successfully")
	}
}

func sendEmailWithAttachment(fileData []byte, filename string) error {
	// Set up authentication information.
	auth := smtp.PlainAuth("", FILEDROP_SMTP_USERNAME, FILEDROP_SMTP_PASSWORD, FILEDROP_SMTP_HOST)

	// Create a new email
	body := new(bytes.Buffer)

	writer := multipart.NewWriter(body)

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              []string{"application/octet-stream"},
		"Content-Transfer-Encoding": []string{"base64"},
		"Content-Disposition":       []string{fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filename))},
	})
	if err != nil {
		panic(err)
	}

	// Write the encoded file to the part
	part.Write([]byte(base64.StdEncoding.EncodeToString(fileData)))

	subject := "[filedrop] " + filename

	// Send the email
	message := []byte("To: " + FILEDROP_EMAIL_TO + "\r\n" +
		"From: " + FILEDROP_EMAIL_FROM + "\r\n" +
		"Subject: " + subject + "\r\n")
	message = append(message, []byte(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n\n", writer.Boundary()))...)
	message = append(message, body.Bytes()...)

	return smtp.SendMail(FILEDROP_SMTP_HOST+":"+FILEDROP_SMTP_PORT, auth, FILEDROP_EMAIL_FROM, []string{FILEDROP_EMAIL_TO}, message)
}
