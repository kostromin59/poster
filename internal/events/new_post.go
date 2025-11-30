package events

import "time"

type NewPost struct {
	EventID   string      `json:"event_id"`
	Data      NewPostData `json:"data"`
	CreatedAt time.Time   `json:"created_at"`
}

type NewPostData struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Content     string         `json:"content"`
	PublishDate time.Time      `json:"publish_date"`
	Tags        []string       `json:"tags"`
	Sources     []string       `json:"sources"`
	Media       []NewPostMedia `json:"media"`
}

type NewPostMedia struct {
	ID       string `json:"id"`
	Filetype string `json:"filetype"`
	URI      string `json:"uri"`
}
