package model

type EmailSend struct {
	To      []string `validate:"required" json:"to"`
	Subject string   `validate:"required" json:"subject"`
	Message string   `validate:"required" json:"message"`
}
