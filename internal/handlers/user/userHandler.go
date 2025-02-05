package user

import (
	"Go-internship-Manifure/internal/auth"
	"encoding/json"
	"net/http"

	"Go-internship-Manifure/internal/kafka"
	"Go-internship-Manifure/internal/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	Users         map[string]model.User
	KafkaProducer kafka.ProducerInterface
}

// Создание нового обработчика пользователя.
func NewUserHandler(kafkaProducer kafka.ProducerInterface) *Handler {
	return &Handler{
		Users:         make(map[string]model.User),
		KafkaProducer: kafkaProducer,
	}
}

// Регистрация нового пользователя.
func (uh *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user model.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if user.Name == "" || user.Email == "" || user.Password == "" {
		http.Error(w, "User name, email and password are required", http.StatusBadRequest)

		return
	}

	user.ID = uuid.New().String()
	uh.Users[user.ID] = user

	message, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	// Генерирует JWT токен при регистрации нового пользователя
	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	err = uh.KafkaProducer.Produce(string(message))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]string{
		"user_id": user.ID,
		"token":   token,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}
}

// Получение пользователя по id.
func (uh *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	user, ok := uh.Users[id]

	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, "Failed to encode user data", http.StatusInternalServerError)

		return
	}
}

// Обновление пользователя.
func (uh *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var updatedUser model.User

	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	user, ok := uh.Users[id]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)

		return
	}

	if updatedUser.Name != "" {
		user.Name = updatedUser.Name
	}

	if updatedUser.Email != "" {
		user.Email = updatedUser.Email
	}

	if updatedUser.Password != "" {
		user.Password = updatedUser.Password
	}

	uh.Users[id] = user

	message, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = uh.KafkaProducer.Produce(string(message))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}
}
