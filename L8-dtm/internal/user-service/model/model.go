package model

type User struct {
	UserID   int64  `json:"user_id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"column:username;uniqueIndex"`
	Email    string `json:"email" gorm:"column:email;uniqueIndex"`
	Password string `json:"password" gorm:"column:password"`
}

func (User) TableName() string {
	return "users"
}
