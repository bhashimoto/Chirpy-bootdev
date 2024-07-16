package main

type Chirp struct {
	ID int `json:"id"`
	Body string `json:"body"`
	AuthorID int `json:"author_id"`
}

