package main

import (
	"Snippetbox/internal/storage/sqlite"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	storage, err := sqlite.New("./storage/storage.db")
	if err != nil {
		fmt.Println(err)
	}

	_ = storage

	infoLog.Printf("Stating server on http://localhost%s/", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
