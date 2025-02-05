package model

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float32 `json:"price"`
	Views int64   `json:"views"`
}
