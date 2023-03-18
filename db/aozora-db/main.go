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
	fmt.Println("create db")

	b, err := os.ReadFile("../../contents/ababababa.txt")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("read file.")

	b, err = japanese.ShiftJIS.NewDecoder().Bytes(b)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("decode shift jis.")

	content := string(b)

	res, _ := db.Exec(
		`insert into contents(author_id, title_id, title, content) values (?, ?, ?, ?)`,
		"000879", "14", "あばばばば", content,
	)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("insert data.")

	docID, err := res.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("get docid", docID)

	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		log.Fatalln(err)
	}

	seg := t.Wakati(content)

	fmt.Println("done wakati", seg)

	_, err = db.Exec(
		`insert into contents_fts(docid, words) values (?, ?)`,
		docID,
		strings.Join(seg, " "),
	)
	if err != nil {
		log.Fatalln(err)
	}

	arg := "虫 AND ココア"

	query := `
	SELECT 
		a.author, 
		c.title
	FROM 
		contents c 
	INNER JOIN 
		authors a 
			on a.author_id = c.author_id 
	INNER JOIN 
		contents_fts f 
			ON c.rowid = f.docid 
			AND words MATCH ?
	`
	rows, err := db.Query(query, arg)

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("query done.", query)

	defer rows.Close()

	for rows.Next() {
		var author, title string
		err = rows.Scan(&author, &title)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(author, title)
	}

	fmt.Println("script done.")

}
