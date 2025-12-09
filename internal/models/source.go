package models

import "errors"

type Source string

var (
	ErrSourceNotFound = errors.New("source not found")
)

var (
	SourceTG      Source = "Телеграмм"
	SourceWebsite Source = "Вебсайт"
)
