package sqlite

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `
		INSERT INTO users (name, email, hashed_password, created)
		VALUES(?, ?, ?, ?)
	`

	_, err = m.DB.Exec(stmt, name, email, string(hashedpassword))
	if err != nil {
		var sqliteError sqlite3.Error
		if errors.As(err, sqliteError) {
			if sqliteError.ExtendedCode == sqlite3.ErrConstraintUnique {
				if strings.Contains(sqliteError.Error(), "users_uc_email") {
					return ErrDuplicateEmail
				}
			}
		}
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	return 0, nil
}

func (m *UserModel) Exitsts(id int) (bool, error) {
	return false, nil
}
