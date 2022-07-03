package database

import (
	"database/sql"
	"time"

	"github.com/annbelievable/go_listing/models"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func InsertAdminSession(db *sql.DB, session_id string, admin_user int64, expiry_date time.Time) error {
	_, err := db.Exec("INSERT INTO admin_user_session(session_id, admin_user, expiry_date, datecreated) VALUES($1, $2, $3, $4);", session_id, admin_user, expiry_date, time.Now())

	return err
}

func SelectAdminSession(db *sql.DB, sessionId string) models.AdminUserSession {
	row := db.QueryRow("SELECT session_id, admin_user, expiry_date FROM admin_user_session WHERE session_id = $1;", sessionId)
	var session models.AdminUserSession
	err := row.Scan(&session.SessionId, &session.AdminUser, &session.ExpiryDate)

	if err != nil || err != sql.ErrNoRows {
		return session
	}

	return session
}

func AdminSessionExist(db *sql.DB, session_id string) bool {
	row := db.QueryRow("SELECT count(*) FROM admin_user_session WHERE session_id = $1;", session_id)

	var count int
	err := row.Scan(&count)

	if err != nil || err != sql.ErrNoRows {
		return false
	}

	return count > 0
}