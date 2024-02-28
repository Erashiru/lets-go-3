package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"Snippetbox/internal/models/sqlite"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       *sqlite.Storage
	users          *sqlite.UserModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
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

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &sqlite.Storage{DB: db},
		users:          &sqlite.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Stating server on http://localhost%s/", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
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

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			token CHAR(43) PRIMARY KEY,
			data BLOB NOT NULL,
			expiry TIMESTAMP NOT NULL
		)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions (expiry)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			hashed_password CHAR(60) NOT NULL,
			created DATETIME NOT NULL
		)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS users_uc_email ON users (email)
	`)
	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	return db, nil
}
