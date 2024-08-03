package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
)

//go:embed template.html
var fs embed.FS

func main() {
	template, err := template.ParseFS(fs, "template.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %s", err)
	}

	title := os.Getenv("STATICLY_TITLE")
	if title == "" {
		title = "Staticly"
	}

	root := os.Getenv("STATICLY_ROOT")
	if root == "" {
		root = "data"
	}

	address := os.Getenv("STATICLY_ADDRESS")
	if address == "" {
		address = ":3000"
	}

	err = os.MkdirAll(root, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create root directory: %s", err)
	}

	handler := &Handler{
		Title: title,
		Root:  root,

		Template: template,
	}

	log.Printf("Listening on address %s", address)
	err = http.ListenAndServe(address, handler)
	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}
}
