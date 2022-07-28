package database

import (
	"database/sql"
	"time"

	"github.com/annbelievable/go_listing/models"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func InsertPage(db *sql.DB, page models.Page) error {
	_, err := db.Exec("INSERT INTO page(title, url, teaser, content, dateupdated, datecreated) VALUES($1, $2, $3, $4, $5, $6);", page.Title, page.Url, page.Teaser, page.Content, time.Now(), time.Now())

	return err
}

func GetPageById(db *sql.DB, id uint64) (models.Page, error) {
	row := db.QueryRow("SELECT id, title, url, teaser, content FROM page WHERE id = $1;", id)
	var page models.Page
	err := row.Scan(&page.Id, &page.Title, &page.Url, &page.Teaser, &page.Content)

	if err != nil {
		return page, err
	}

	return page, nil
}

func GetPageByUrl(db *sql.DB, url string) (models.Page, error) {
	row := db.QueryRow("SELECT title, url, teaser, content FROM page WHERE url = $1;", url)
	var page models.Page
	err := row.Scan(&page.Id, &page.Title, &page.Url, &page.Teaser, &page.Content)

	if err != nil {
		return page, err
	}

	return page, nil
}

func GetPages(db *sql.DB) ([]models.Page, error) {
	rows, err := db.Query("SELECT id, title, url, teaser, content FROM page ORDER BY url ASC;")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []models.Page
	for rows.Next() {
		var page models.Page
		if err := rows.Scan(&page.Id, &page.Title, &page.Url, &page.Teaser, &page.Content); err != nil {
			return pages, err
		}
		pages = append(pages, page)
	}

	return pages, nil
}

func UpdatePage(db *sql.DB, page models.Page) error {
	_, err := db.Exec("UPDATE page SET title = $1, url = $2, teaser = $3, content = $4 WHERE id = $5;", page.Title, page.Url, page.Teaser, page.Content, page.Id)
	return err
}

func DeletePage(db *sql.DB, id uint64) {
	db.Exec("DELETE FROM page WHERE id = $1;", id)
}
