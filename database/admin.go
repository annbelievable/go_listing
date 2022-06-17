package database

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func InsertAdmin(db *sql.DB, email, password string) error {
	_, err := db.Exec("INSERT INTO admin_user(email, password, dateupdated, datecreated) VALUES($1, $2, $3, $4);", email, password, time.Now(), time.Now())

	return err
}

func SelectAdminHpwd(db *sql.DB, email string) (string, error) {
	row := db.QueryRow("SELECT password FROM admin_user WHERE email = $1;", email)

	var hpwd string
	err := row.Scan(&hpwd)

	if err != nil && err != sql.ErrNoRows {
		return hpwd, err
	}

	return hpwd, nil
}

// to be implemented
// func SelectAdminData(db *sql.DB, email string) (int, error) {
// }

func AdminEmailExist(db *sql.DB, email string) (int, error) {
	row := db.QueryRow("SELECT count(*) AS count FROM admin_user WHERE email = $1;", email)
	var count int
	err := row.Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return count, err
	}

	return count, nil
}
