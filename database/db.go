package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func ConnectDatabase() *sql.DB {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "127.0.0.1", 5432, "postgres", "password", "go_my_diary")

	db, err := sql.Open("pgx", psqlconn)

	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(1 * time.Second)
	db.SetConnMaxLifetime(30 * time.Second)

	if err := db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}

	return db
}

func CreateAdminTable(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS admin_user (id SERIAL PRIMARY KEY NOT NULL, email CHAR(255) NOT NULL, password CHAR(255) NOT NULL, dateupdated TIMESTAMP NOT NULL, datecreated TIMESTAMP NOT NULL);")

	if err != nil {
		return err
	}

	return nil
}
