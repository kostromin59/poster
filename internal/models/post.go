package models

import "time"

type PostID ID[Post]

type Post struct {
	ID          PostID
	Title       string
	Content     string
	PublishDate time.Time
	Tags        []Tag
	Sources     []Source
	Media       []Media
}

type CreatePostDTO struct {
	Title       string
	Content     string
	PublishDate time.Time
	Tags        []Tag
	Sources     []Source
	Media       []MediaID
}
