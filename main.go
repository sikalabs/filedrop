package main

import (
	"bytes"
	_ "embed"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/sikalabs/go-utils/pkg/mail"
)

//go:embed upload.html
var UPLOAD_HTML_FILE_CONTENT []byte

//go:embed done.html
var DONE_HTML_FILE_CONTENT []byte

var FILEDROP_SMTP_HOST string
var FILEDROP_SMTP_PORT int
var FILEDROP_EMAIL_FROM string
var FILEDROP_SMTP_USERNAME string
var FILEDROP_SMTP_PASSWORD string
var FILEDROP_EMAIL_TO string

func main() {
	var err error
	FILEDROP_SMTP_HOST = os.Getenv("FILEDROP_SMTP_HOST")
	if FILEDROP_SMTP_HOST == "" {
		log.Fatal("FILEDROP_SMTP_HOST is not set")
	}
	FILEDROP_SMTP_PORT_ENV := os.Getenv("FILEDROP_SMTP_PORT")
	if FILEDROP_SMTP_PORT_ENV == "" {
		log.Fatal("FILEDROP_SMTP_PORT is not set")
	}
	FILEDROP_SMTP_PORT, err = strconv.Atoi(FILEDROP_SMTP_PORT_ENV)
	if err != nil {
		log.Fatal("FILEDROP_SMTP_PORT is not a number")
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
		w.Write(UPLOAD_HTML_FILE_CONTENT)
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
		err = mail.SendEmailWithAttachment(
			FILEDROP_SMTP_HOST,
			FILEDROP_SMTP_PORT,
			FILEDROP_SMTP_USERNAME,
			FILEDROP_SMTP_PASSWORD,
			FILEDROP_EMAIL_FROM,
			FILEDROP_EMAIL_TO,
			"[filedrop] "+header.Filename,
			"", // No body
			header.Filename,
			fileData,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(DONE_HTML_FILE_CONTENT)
	}
}
