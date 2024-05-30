package models

import (
	"time"
)

type SentEmail struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	LastModifiedAt time.Time `json:"last_modified_at"`
	SenderEmail    string    `json:"sender_email"`
	ReceiverEmail  string    `json:"receiver_email"`
	TemplateName   string    `json:"template_name"`
}
