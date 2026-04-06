package model

type Post struct {
	PostID  int64  `json:"post_id" gorm:"primaryKey"`
	Owner   int64  `json:"owner" gorm:"column:owner;index"`
	Title   string `json:"title" gorm:"column:title"`
	Content string `json:"content" gorm:"column:content"`
}

func (Post) TableName() string {
	return "posts"
}
