package model

type Recommendations struct {
	ID              string  `gorm:"primary_key"`
	Name            string  `gorm:"not null"`
	Price           float32 `gorm:"default:0"`
	PopularityScore int     `gorm:"not null"`
}
