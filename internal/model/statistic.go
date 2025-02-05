package model

type ProductStatistics struct {
	ProductID   string `gorm:"primaryKey"`
	UpdateCount int    `gorm:"default:0"`
}

type UserStatistics struct {
	UserID        string `gorm:"primaryKey"`
	ActivityCount int    `gorm:"default:0"`
}
