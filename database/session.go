package database

import (
	"database/sql"
	"time"

	"github.com/annbelievable/go_listing/models"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func InsertAdminSession(db *sql.DB, session_id string, admin_user uint64, expiry_date time.Time) error {
	_, err := db.Exec("INSERT INTO admin_user_session(session_id, admin_user, expiry_date, datecreated) VALUES($1, $2, $3, $4);", session_id, admin_user, expiry_date, time.Now())

	return err
}

func SelectAdminSession(db *sql.DB, sessionId string) (models.AdminUserSession, error) {
	row := db.QueryRow("SELECT session_id, admin_user, expiry_date FROM admin_user_session WHERE session_id = $1;", sessionId)
	var session models.AdminUserSession
	err := row.Scan(&session.SessionId, &session.AdminUser, &session.ExpiryDate)

	if err != nil {
		return session, err
	}

	return session, nil
}

func AdminSessionExist(db *sql.DB, session_id string) bool {
	row := db.QueryRow("SELECT count(*) FROM admin_user_session WHERE session_id = $1;", session_id)

	var count int
	err := row.Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

func DeleteAdminSession(db *sql.DB, session_id string) {
	db.Exec("DELETE FROM admin_user_session WHERE session_id = $1;", session_id)
}

func DeleteAdminSessionByAdminId(db *sql.DB, adminId uint64) {
	db.Exec("DELETE FROM admin_user_session WHERE admin_user = $1;", adminId)
}
