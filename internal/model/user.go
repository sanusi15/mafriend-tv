package model

import "time"

type User struct {
	GoogleID  string    `gorm:"primaryKey;type:varchar(100)" json:"google_id"`
	Name      string    `gorm:"type:varchar(150);not null" json:"name"`
	Email     string    `gorm:"type:varchar(150);not null" json:"email"`
	Picture   string    `gorm:"type:text" json:"picture"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Menentukan nama tabel secara eksplisit agar sesuai keinginan lu
func (User) TableName() string {
	return "tbl_users"
}