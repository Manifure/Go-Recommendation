package recommendation_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Go-internship-Manifure/internal/handlers/recommendation"
	"Go-internship-Manifure/internal/model"
	"Go-internship-Manifure/internal/redis"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Миграция схемы
	err = db.AutoMigrate(&model.Recommendations{})
	require.NoError(t, err)

	return db
}

func TestHandleProductMessage(t *testing.T) {
	db := setupTestDB(t)
	handler := recommendation.NewRecommendationHandler(db)

	product := model.Recommendations{
		ID:              "test-product-id",
		Name:            "Test Product",
		Price:           100.0,
		PopularityScore: 0,
	}

	// Отправка сообщения о продукте
	message, err := json.Marshal(product)
	require.NoError(t, err)

	err = handler.HandleProductMessage(message)
	require.NoError(t, err, "handleProductMessage should not return an error")

	// Проверка, что продукт был добавлен с правильными данными
	var savedProduct model.Recommendations
	err = db.First(&savedProduct, "id = ?", product.ID).Error
	require.NoError(t, err)
	require.Equal(t, product.ID, savedProduct.ID)
	require.Equal(t, 1, savedProduct.PopularityScore) // Должно быть увеличено до 1
}

func TestHandleUserMessage(t *testing.T) {
	db := setupTestDB(t)
	handler := recommendation.NewRecommendationHandler(db)

	// Создаем тестового пользователя с корзиной
	user := model.User{
		ID:       "test-user-id",
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "test",
		Cart: []struct {
			ProductID string `json:"product_id"`
		}{
			{ProductID: "test-product-id-1"},
			{ProductID: "test-product-id-2"},
		},
	}

	// Отправка сообщения о пользователе
	message, err := json.Marshal(user)
	require.NoError(t, err)

	err = handler.HandleUserMessage(message)
	require.NoError(t, err, "handleUserMessage should not return an error")

	// Проверка, что продукты из корзины были добавлены с правильными данными
	var product1, product2 model.Recommendations
	err = db.First(&product1, "id = ?", "test-product-id-1").Error
	require.NoError(t, err)
	require.Equal(t, 1, product1.PopularityScore)

	err = db.First(&product2, "id = ?", "test-product-id-2").Error
	require.NoError(t, err)
	require.Equal(t, 1, product2.PopularityScore)
}

func setupTestAPI(t *testing.T) (*gorm.DB, *redis.CacheMock, *recommendation.APIHandler) {
	t.Helper()

	// Инициализация SQLite в памяти
	db := setupTestDB(t)

	// Инициализация mock для Redis
	cache := redis.NewCacheMock()

	// Инициализация API handler
	apiHandler := recommendation.NewRecommendationAPIHandler(db, cache)

	return db, cache, apiHandler
}

func TestGetRecommendations_NoCache(t *testing.T) {
	db, _, apiHandler := setupTestAPI(t)

	// Добавляем тестовые данные в базу данных
	testProducts := []model.Recommendations{
		{ID: "product1", Name: "Product 1", Price: 10.0, PopularityScore: 5},
		{ID: "product2", Name: "Product 2", Price: 15.0, PopularityScore: 10},
		{ID: "product3", Name: "Product 3", Price: 20.0, PopularityScore: 7},
	}
	require.NoError(t, db.Create(&testProducts).Error)

	// Создаем HTTP-запрос
	req, err := http.NewRequest(http.MethodGet, "/recommendations?limit=2", nil)
	require.NoError(t, err)

	// Создаем HTTP-response-рекордер
	rec := httptest.NewRecorder()

	// Вызываем handler
	apiHandler.GetRecommendations(rec, req)

	// Проверяем ответ
	require.Equal(t, http.StatusOK, rec.Code)

	// Раскодируем JSON-ответ
	var recommendations []model.Recommendations
	err = json.Unmarshal(rec.Body.Bytes(), &recommendations)
	require.NoError(t, err)

	// Проверяем содержимое ответа
	require.Len(t, recommendations, 2)
	require.Equal(t, "Product 2", recommendations[0].Name) // Проверяем сортировку по PopularityScore
	require.Equal(t, "Product 3", recommendations[1].Name)
}

func TestGetRecommendations_WithCache(t *testing.T) {
	_, cache, apiHandler := setupTestAPI(t)

	// Добавляем данные в кэш
	cachedData := `[{"id":"product1","name":"Cached Product 1","price":10,"popularity_score":5}]`
	require.NoError(t, cache.Set("recommendations:limit:1", cachedData, 10*time.Minute))

	// Создаем HTTP-запрос
	req, err := http.NewRequest(http.MethodGet, "/recommendations?limit=1", nil)
	require.NoError(t, err)

	// Создаем HTTP-response-рекордер
	rec := httptest.NewRecorder()

	// Вызываем handler
	apiHandler.GetRecommendations(rec, req)

	// Проверяем ответ
	require.Equal(t, http.StatusOK, rec.Code)

	// Раскодируем JSON-ответ
	var recommendations []model.Recommendations
	err = json.Unmarshal(rec.Body.Bytes(), &recommendations)
	require.NoError(t, err)

	// Проверяем содержимое ответа
	require.Len(t, recommendations, 1)
	require.Equal(t, "Cached Product 1", recommendations[0].Name) // Проверяем, что данные взяты из кэша
}
