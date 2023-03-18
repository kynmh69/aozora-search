package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/text/encoding/japanese"
)

func main() {
	db, err := sql.Open("sqlite3", "database.sqlite")

	if err != nil {
		log.Fatalln("not open db. err")
	}

	defer db.Close()

	queries := []string{
		`create table if not exists authors(author_id TEXT, author TEXT, PRIMARY KEY(author_id))`,
		`create table if not exists contents(author_id TEXT, title_id TEXT, title TEXT, content TEXT, PRIMARY KEY(author_id, title_id))`,
		`create virtual table if not exists contents_fts USING fts4(words)`,
	}
	for _, q := range queries {
		_, err = db.Exec(q)
		if err != nil {
			log.Fatalln(err)
		}
	}

	b, err := os.ReadFile("../../contents/ababababa.txt")
	if err != nil {
		log.Fatalln(err)
	}

	b, err = japanese.ShiftJIS.NewDecoder().Bytes(b)
	if err != nil {
		log.Fatalln(err)
	}

	content := string(b)

	res, err := db.Exec(
		`insert into contents(author_id, title_id, title, content) values (?, ?, ?, ?)`,
		"000879", "14", "あばばばば", content,
	)
	if err != nil {
		log.Fatalln(err)
	}

	docID, err := res.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}

	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		log.Fatalln(err)
	}

	seg := t.Wakati(content)

	_, err = db.Exec(
		`insert into contents_fts(docid, words) values (?, ?)`,
		docID,
		strings.Join(seg, " "),
	)
	if err != nil {
		log.Fatalln(err)
	}

	query := "虫 AND ココア"
	rows, err := db.Query(`
	select 
		a.author, 
		c.title
	from
		contents c 
	inner join 
		authors a
			on a.author_id = c.author_id 
	inner join 
		contents_fts f 
			on c.rowid = f.docid 
			and words MATCH ?
	`,
		query,
	)

	if err != nil {
		log.Fatalln(err)
	}

	defer rows.Close()

	for rows.Next() {
		var author, title string
		err := rows.Scan(&author, &title)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(author, title)
	}

}
