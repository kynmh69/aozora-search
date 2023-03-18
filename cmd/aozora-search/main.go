package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const dataSourceName = "database.sql"

func main() {
	usage := "aaaaaa"
	var dsn string
	flag.StringVar(&dsn, "d", dataSourceName, "database")
	flag.Usage = func() {
		fmt.Println(usage)
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	switch flag.Arg(0) {
	case "authors":
		err = showAuthers(db)
	case "titles":
		if flag.NArg() != 2 {
			flag.Usage()
			os.Exit(2)
		}
		err = showTitles(db, flag.Arg(1))
	case "content":
		if flag.NArg() != 3 {
			flag.Usage()
			os.Exit(2)
		}
		err = showContent(db, flag.Arg(1), flag.Arg(2))
	case "query":
		if flag.NArg() != 2 {
			flag.Usage()
			os.Exit(2)
		}
		err = queryContent(db, flag.Arg(1))
	}

	if err != nil {
		log.Fatal(err)
	}
}

func queryContent(db *sql.DB, s string) error {
	query := `
	SELECT
		a.author,
		c.title_id,
		c.title
	FROM
		contents c
	INNER JOIN authors a
		ON a.author_id = c.author_id
	INNER JOIN contents_fts f
		ON c.rowid = f.docid
		AND words MATCH ?
`
	rows, err := db.Query(query, s)
	if err != nil {
		return err
	}

	columns, _ := rows.Columns()
	log.Println(columns)

	for rows.Next() {
		var author, title_id, title string
		rows.Scan(&author, &title_id, &title)
		log.Println(author, title_id, title)
	}
	return nil
}

func showContent(db *sql.DB, s1, s2 string) error {
	query := "select (author_id, title_id, title, content) from contents where author_id = ? and where title_id = ?"
	rows, err := db.Query(query, s1, s2)
	if err != nil {
		return err
	}
	for rows.Next() {
		var author_id, title_id, title, content string
		rows.Scan(&author_id, &title_id, &title, &content)
		log.Println(author_id, title_id, title, content)
	}
	return nil
}

func showTitles(db *sql.DB, s string) error {
	// query := "select (author_id, title, title_id, content) from contents"
	query := "select * from contents where author_id = ?"
	log.Printf("%s, %T", s, s)
	rows, err := db.Query(query, s)
	if err != nil {
		return err
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	log.Println(columns)

	for rows.Next() {
		var author_id, title_id, title, content string
		rows.Scan(&author_id, &title, &title_id, &content)
		log.Println(title_id, title)
	}
	return nil
}

func showAuthers(db *sql.DB) error {
	query := "select author_id, author from authors"
	rows, err := db.Query(query)

	if err != nil {
		return err
	}

	for rows.Next() {
		var author_id, author string
		rows.Scan(&author_id, &author)
		log.Println(author_id, author)
	}
	return nil
}
