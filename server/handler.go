package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"
	"unicode"
)

type DirEntry struct {
	Name         string
	LastModified time.Time
	IsFile       bool
	IsDirectory  bool
}

type Handler struct {
	Title string
	Root  string

	Template *template.Template
}

func (hand *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	reqPath := req.URL.Path
	log.Printf("%-35s %s", req.RemoteAddr, reqPath)

	if !strings.HasSuffix(reqPath, "/") {
		http.Redirect(
			writer,
			req,
			fmt.Sprintf("%s/", reqPath),
			http.StatusMovedPermanently,
		)
		return
	}

	name := path.Join(hand.Root, reqPath)
	fileInfo, err := os.Stat(name)
	switch {
	case os.IsNotExist(err):
		writer.WriteHeader(http.StatusNotFound)
		return
	case os.IsPermission(err):
		writer.WriteHeader(http.StatusForbidden)
		return
	case err != nil:
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	fileMode := fileInfo.Mode()
	switch {
	case fileMode.IsRegular():
		hand.serveFile(name, fileInfo, writer, req)
		return
	case fileMode.IsDir():
		hand.serveDirectory(name, writer, req)
		return
	default:
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (hand *Handler) serveFile(name string, fileInfo os.FileInfo, writer http.ResponseWriter, req *http.Request) {
	file, err := os.Open(name)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer file.Close()
	http.ServeContent(writer, req, fileInfo.Name(), fileInfo.ModTime(), file)
}

func (hand *Handler) serveDirectory(name string, writer http.ResponseWriter, req *http.Request) {

	rawEntries, err := os.ReadDir(name)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	entries := make([]DirEntry, len(rawEntries))
	for i, rawEntry := range rawEntries {
		entryMode := rawEntry.Type()
		entry := DirEntry{
			Name:         rawEntry.Name(),
			LastModified: time.UnixMilli(0),
		}

		entryInfo, err := rawEntry.Info()
		if err == nil {
			entry.LastModified = entryInfo.ModTime()
		}

		switch {
		case entryMode.IsRegular():
			entry.IsFile = true
		case entryMode.IsDir():
			entry.IsDirectory = true
		}

		entries[i] = entry
	}

	hand.sortEntires(
		req.URL.Query().Get("sort_by"),
		entries,
	)

	err = hand.Template.ExecuteTemplate(
		writer,
		"directory",
		map[string]any{
			"Title":   hand.Title,
			"Path":    req.URL.Path,
			"Entries": entries,
		},
	)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (handler *Handler) sortEntires(sortBy string, entries []DirEntry) {
	switch sortBy {
	case "type", "":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].IsFile && entries[j].IsDirectory
		})
	case "name":
		sort.Slice(entries, func(i, j int) bool {
			iRunes := []rune(entries[i].Name)
			jRunes := []rune(entries[j].Name)

			max := len(iRunes)
			if max > len(jRunes) {
				max = len(jRunes)
			}

			for idx := 0; idx < max; idx++ {
				ir := iRunes[idx]
				jr := jRunes[idx]

				lir := unicode.ToLower(ir)
				ljr := unicode.ToLower(jr)

				if lir != ljr {
					return lir < ljr
				}

				// the lowercase runes are the same, so compare the original
				if ir != jr {
					return ir < jr
				}
			}

			return false
		})
	case "last_modified":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].LastModified.Before(entries[j].LastModified)
		})
	}
}
