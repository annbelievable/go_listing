package database

import (
	"database/sql"
	"time"

	"github.com/annbelievable/go_listing/models"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func InsertPage(db *sql.DB, email, password string) error {
	_, err := db.Exec("INSERT INTO admin_user(email, password, dateupdated, datecreated) VALUES($1, $2, $3, $4);", email, password, time.Now(), time.Now())

	return err
}

func SelectPage(db *sql.DB, email string) (models.AdminUser, error) {
	row := db.QueryRow("SELECT id, email, password AS count FROM admin_user WHERE email = $1;", email)
	var admin models.AdminUser
	err := row.Scan(&admin.Id, &admin.Email, &admin.Password)

	if err != nil && err != sql.ErrNoRows {
		return admin, err
	}

	return admin, nil
}

func UpdatePage(db *sql.DB, email string) (string, error) {
	row := db.QueryRow("SELECT password FROM admin_user WHERE email = $1;", email)

	var hpwd string
	err := row.Scan(&hpwd)

	if err != nil && err != sql.ErrNoRows {
		return hpwd, err
	}

	return hpwd, nil
}

func DeletePage(db *sql.DB, email string) bool {
	row := db.QueryRow("SELECT count(*) AS count FROM admin_user WHERE email = $1;", email)
	var count int
	err := row.Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return false
	}

	return count > 0
}
