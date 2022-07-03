package main

import "time"

type Page struct {
	Id      string
	Url     string
	Title   string
	Teaser  string
	Content string
	Message string
	Errors  map[string]string
}

type AdminUser struct {
	Id    uint64
	Email string
}

type AdminUserSession struct {
	SessionId  string
	AdminUser  int
	ExpiryDate time.Time
}
