package main

type PageData struct {
	Id      string
	Url     string
	Title   string
	Teaser  string
	Content string
	Message string
	Errors  map[string]string
}
