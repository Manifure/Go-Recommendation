package user_test

import (
	"bytes"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Go-internship-Manifure/internal/handlers/user"
	"Go-internship-Manifure/internal/kafka"
	"Go-internship-Manifure/internal/model"
	"github.com/gorilla/mux"
)

var jwtKey = []byte("your_secret_key")

func generateTestJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

func TestRegisterUser(t *testing.T) {
	mockProducer := &kafka.MockProducer{}
	handler := user.NewUserHandler(mockProducer)

	// Создание тестового HTTP-запроса
	userData := model.User{ID: "123", Name: "John Doe", Email: "john@example.com", Password: "password"}

	body, err := json.Marshal(userData)
	if err != nil {
		t.Error(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Создание ResponseRecorder
	responseRecorder := httptest.NewRecorder()

	// Вызов обработчика
	handler.RegisterUser(responseRecorder, req)

	// Проверка результата
	if status := responseRecorder.Code; status != http.StatusCreated {
		t.Errorf("unexpected status code: got %v, want %v", status, http.StatusCreated)
	}

	// Проверка Kafka Producer
	if len(mockProducer.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(mockProducer.Messages))
	}
}

func TestGetUser(t *testing.T) {
	// Создаю мок-продюсер и обработчик
	mockProducer := &kafka.MockProducer{}
	handler := user.NewUserHandler(mockProducer)

	// Создаю нового пользователя
	userData := model.User{
		ID: "123", Name: "John Doe", Email: "john@example.com", Password: "password", Cart: []struct {
			ProductID string `json:"product_id"`
		}{
			{ProductID: "test-product-id-1"},
		},
	}

	body, err := json.Marshal(userData)
	if err != nil {
		t.Error(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.RegisterUser(responseRecorder, req)

	if responseRecorder.Code != http.StatusCreated {
		t.Fatalf("failed to create user: got %v, want %v", responseRecorder.Code, http.StatusCreated)
	}

	// Получаю ID созданного пользователя из ответа
	var createdUser map[string]string

	err = json.NewDecoder(responseRecorder.Body).Decode(&createdUser)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	userID := createdUser["user_id"]
	if userID == "" {
		t.Fatalf("user ID not returned in response")
	}

	// Генерация токена для созданного пользователя
	token, err := generateTestJWT(userID)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	// Убедиться, что пользователь сохранен
	if _, exists := handler.Users[userID]; !exists {
		t.Fatalf("user not saved in handler: %+v", userData)
	}

	// Выполняется запрос на получение пользователя
	getUserReq := httptest.NewRequest(http.MethodGet, "/users/"+userID, nil)
	getUserReq.Header.Set("Authorization", token) // Добавляем токен в заголовок
	getUserRR := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/users/{id}", handler.GetUser).Methods(http.MethodGet)
	router.ServeHTTP(getUserRR, getUserReq)

	// Проверяется статус ответа
	if getUserRR.Code != http.StatusOK {
		t.Errorf("unexpected status code: got %v, want %v", getUserRR.Code, http.StatusOK)
	}

	// Проверка, что пользователь возвращен корректно
	var fetchedUser model.User

	err = json.NewDecoder(getUserRR.Body).Decode(&fetchedUser)
	if err != nil {
		t.Errorf("failed to decode response body: %v", err)
	}

	if fetchedUser.ID != userID || fetchedUser.Name != userData.Name || fetchedUser.Email != userData.Email {
		t.Errorf("unexpected user data: got %+v, want %+v", fetchedUser, userData)
	}
}
