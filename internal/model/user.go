package model

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Cart     []struct {
		ProductID string `json:"product_id"`
	} `json:"cart"`
}
