package models

import (
	"errors"
	"time"
)

var (
	ErrPostNotFound = errors.New("post not found")
)

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

type PostSearchFilters struct {
	Title         *string
	Tags          []string
	Sources       []string
	PublishedFrom *time.Time
}

type CreatePostDTO struct {
	Title       string
	Content     string
	PublishDate time.Time
	Tags        []Tag
	Sources     []Source
	Media       []MediaID
}
