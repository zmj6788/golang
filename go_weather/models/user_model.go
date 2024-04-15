package models

type User struct {
	Username string `gorm:"unique_index:username(100);not null"`
	Password string `gorm:"not null"`
}
