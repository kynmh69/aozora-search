package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "database.sqlite")

	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	queries := []string{
		`create table if not exists authors(author_id TEXT, auther TEXT, PRIMARY KEY(auther_id))`,
		`create table if not exists contents(author_id TEXT, title_id TEXT, title TEXT, content TEXT PRIMARY KEY(autnor_id, title_id))`,
		`create virtual table if not exists contents_fts USING fts4(words)`,
	}
	for _, q := range queries {
		_, err = db.Exec(q)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
