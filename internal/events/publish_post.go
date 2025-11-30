package events

import "time"

type PublishedPost struct {
	EventID   string            `json:"event_id"`
	Data      PublishedPostData `json:"data"`
	CreatedAt time.Time         `json:"created_at"`
}

type PublishedPostData struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Content     string               `json:"content"`
	PublishDate time.Time            `json:"publish_date"`
	Tags        []string             `json:"tags"`
	Sources     []string             `json:"sources"`
	Media       []PublishedPostMedia `json:"media"`
}

type PublishedPostMedia struct {
	ID       string `json:"id"`
	Filetype string `json:"filetype"`
	URI      string `json:"uri"`
}
