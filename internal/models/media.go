package models

import "errors"

var (
	ErrMediaNotFound = errors.New("media not found")
)

type MediaID ID[Media]

type Media struct {
	ID       MediaID
	Filetype string
	URI      string
}
