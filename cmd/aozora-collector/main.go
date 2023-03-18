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
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/mattn/go-sqlite3"
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

const dataSourceName = "database.sql"

func setupDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		log.Fatalln(err)
	}

	queries := []string{
		"CREATE TABLE IF NOT EXISTS authors(author_id TEXT, author TEXT, PRIMARY KEY(author_id))",
		"CREATE TABLE IF NOT EXISTS contents(author_id TEXT, title TEXT, title_id TEXT, content TEXT, PRIMARY KEY(author_id, title_id))",
		"CREATE VIRTUAL TABLE IF NOT EXISTS contents_fts USING fts4(words)",
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			log.Fatalln("create table err", err)
		}
	}
	return db, nil
}

func addEntry(db *sql.DB, entry *Entry, content string) error {
	_, err := db.Exec("INSERT INTO authors(author_id, author) values (?, ?)", entry.AuthorID, entry.Author)
	if err != nil {
		return err
	}

	res, err := db.Exec("REPLACE INTO contents(author_id, title_id, title, content) values (?, ?, ?, ?)",
		entry.AuthorID, entry.TitleID, entry.Title, content)
	if err != nil {
		return err
	}

	docID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())

	if err != nil {
		return err
	}

	seg := t.Wakati(content)

	_, err = db.Exec("REPLACE INTO contents_fts(docid, words) values (?, ?)", docID, strings.Join(seg, " "))

	if err != nil {
		return err
	}

	return nil

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

	db, err := setupDB(dataSourceName)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	entries, err := findEntries(listURL)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("found entries.", len(entries))

	for _, entry := range entries {
		log.Printf("Adding %+v\n", entry)
		content, err := extractText(entry.ZipURL)

		if err != nil {
			log.Println(err)
			continue
		}

		err = addEntry(db, &entry, content)

		if err != nil {
			log.Println(err)
			continue
		}
	}
}
