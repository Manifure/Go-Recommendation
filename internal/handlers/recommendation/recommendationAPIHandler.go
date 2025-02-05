package recommendation

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"Go-internship-Manifure/internal/model"
	"Go-internship-Manifure/internal/redis"
	"gorm.io/gorm"
)

type APIHandler struct {
	DB    *gorm.DB
	Cache redis.CacheInterface
}

// Инициализация нового API обработчика.
func NewRecommendationAPIHandler(db *gorm.DB, cache redis.CacheInterface) *APIHandler {
	return &APIHandler{
		DB:    db,
		Cache: cache,
	}
}

// Получение рекомендаций.
func (api *APIHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	limitParam := r.URL.Query().Get("limit")
	limit := 10 // default value

	// Парсим параметр limit
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
			limit = parsedLimit
		}
	}

	cacheKey := fmt.Sprintf("recommendations:limit:%d", limit) // Ключ для кэша
	// Проверка наличия данных в кэше
	cacheData, err := api.Cache.Get(cacheKey)
	if err != nil {
		log.Printf("Error accessing Redis: %v\n", err)
		http.Error(w, "Failed to access cache", http.StatusInternalServerError)

		return
	}

	if cacheData != "" {
		log.Printf("Cache found")
		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write([]byte(cacheData))
		if err != nil {
			log.Printf("Error writing response: %v\n", err)

			return
		}

		return
	}
	// Если данных нет в кэше, выполняем запрос к базе данных
	log.Println("Not found recommendations in cache")

	var product []model.Recommendations

	if err := api.DB.Order("popularity_score DESC").Limit(limit).Find(&product).Error; err != nil {
		http.Error(w, "Failed to fetch recommendations", http.StatusInternalServerError)

		return
	}
	// Кодируем данные в JSON
	recommendationsJSON, err := json.Marshal(product)
	if err != nil {
		http.Error(w, "Failed to fetch recommendations", http.StatusInternalServerError)

		return
	}

	// Сохраняем результат в кэше с TTL 1 минута
	if err := api.Cache.Set(cacheKey, string(recommendationsJSON), 1*time.Minute); err != nil {
		log.Printf("Failed to cache recommendations: %v", err)
	}

	// Возврат результата
	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(recommendationsJSON)
	if err != nil {
		log.Printf("Failed to writing data: %v", err)

		return
	}
}
