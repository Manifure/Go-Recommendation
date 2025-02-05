package model

type ProductUpdateAnalytics struct {
	ID string `json:"id" validate:"required,uuid"` // Поле ID обязательно и должно быть UUID
}

type UserUpdateAnalytics struct {
	ID string `json:"id" validate:"required,uuid"` // Поле ID обязательно и должно быть UUID
}
