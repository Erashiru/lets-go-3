package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	"Snippetbox/internal/models/sqlite"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	snippets *sqlite.Storage
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "./storage/storage.db", "MySQL data source name")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := New(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &sqlite.Storage{DB: db},
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	// storage, err := sqlite.New("./storage/storage.db")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// expires := time.Now().AddDate(0, 0, 7)
	// id, err := storage.InsertSnippets("adil", "lox", expires)
	// if err != nil {
	// 	errorLog.Fatal(err)
	// }

	// fmt.Println(id)

	infoLog.Printf("Stating server on http://localhost%s/", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func New(storagePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS snippets (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		title VARCHAR(100) NOT NULL,
		content TEXT NOT NULL,
		created DATETIME NOT NULL,
		expires DATETIME NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_snippets_created ON snippets(created);
	`)

	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return db, nil
}
