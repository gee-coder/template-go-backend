package model

// ContactSubmission describes a message submitted from the website.
type ContactSubmission struct {
	BaseModel
	Name    string `gorm:"size:64;not null" json:"name"`
	Email   string `gorm:"size:128;not null" json:"email"`
	Phone   string `gorm:"size:32" json:"phone"`
	Company string `gorm:"size:128" json:"company"`
	Message string `gorm:"type:text;not null" json:"message"`
	Source  string `gorm:"size:64;default:website" json:"source"`
}

