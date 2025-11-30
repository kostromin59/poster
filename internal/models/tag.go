package models

import "errors"

var (
	ErrTagNotFound = errors.New("tag not found")
)

type Tag string
