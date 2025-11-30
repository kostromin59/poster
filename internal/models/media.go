package models

type MediaID ID[Media]

type Media struct {
	ID       MediaID
	Filetype string
	URI      string
}
