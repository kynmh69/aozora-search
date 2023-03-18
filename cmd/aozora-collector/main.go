package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
)

type Entry struct {
	AuthorID string
	Author   string
	TitleID  string
	Title    string
	InfoURL  string
	ZipURL   string
}

func setupDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		log.Fatalln(err)
	}

	queries := []string{
		`create table`,
	}
}

func findEntries(siteURL string) ([]Entry, error) {
	entries := []Entry{}
	doc, err := goquery.NewDocument(siteURL)
	if err != nil {
		return nil, err
	}
	pat := regexp.MustCompile(`.*/cards/([0-9]+)/card([0-9]+).html$`)
	doc.Find("ol li a").Each(func(_ int, elem *goquery.Selection) {
		attrOr := elem.AttrOr("href", "")
		title := elem.Text()
		token := pat.FindStringSubmatch(attrOr)
		if len(token) != 3 {
			return
		}
		fmt.Println(title, attrOr)
		pageURL := fmt.Sprintf("https://www.aozora.gr.jp/cards/%s/card%s.html", token[1], token[2])
		fmt.Println(pageURL)
		auther, zipURL := findAutherAndZIP(pageURL)

		if zipURL != "" {
			entries = append(entries, Entry{
				AuthorID: token[1],
				Author:   auther,
				TitleID:  token[2],
				Title:    title,
				InfoURL:  siteURL,
				ZipURL:   zipURL,
			})
		}
	})

	return entries, nil
}

func findAutherAndZIP(pageURL string) (string, string) {
	doc, err := goquery.NewDocument(pageURL)
	if err != nil {
		return "", ""
	}
	auther := doc.Find("table[summary=作家データ] tr:nth-child(1) td:nth-child(2)").Text()
	zipURL := ""
	doc.Find("table.download a").Each(func(_ int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if strings.HasSuffix(href, ".zip") {
			zipURL = href
		}
	})

	if zipURL == "" {
		return auther, zipURL
	}

	if strings.HasPrefix(zipURL, "http://") || strings.HasPrefix(zipURL, "https://") {
		return auther, zipURL
	}
	u, err := url.Parse(pageURL)
	if err != nil {
		return auther, ""
	}
	u.Path = path.Join(path.Dir(u.Path), zipURL)
	return auther, u.String()
}

func extractText(zipURL string) (string, error) {
	resp, err := http.Get(zipURL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", err
	}

	for _, file := range r.File {
		if path.Ext(file.Name) == ".txt" {
			f, err := file.Open()
			if err != nil {
				return "", err
			}
			b, err := ioutil.ReadAll(f)
			f.Close()

			if err != nil {
				return "", err
			}

			b, err = japanese.ShiftJIS.NewDecoder().Bytes(b)
			if err != nil {
				return "", err
			}

			return string(b), nil
		}
	}
	return "", errors.New("contents not found")
}

func main() {
	listURL := "https://www.aozora.gr.jp/index_pages/person879.html"
	entries, err := findEntries(listURL)
	if err != nil {
		log.Fatalln(err)
	}
	for _, entry := range entries {
		fmt.Println(entry.Title, entry.ZipURL)
	}
}
