package analytics_test

import (
	"encoding/json"
	"log"
	"testing"

	"Go-internship-Manifure/internal/handlers/analytics"
	"Go-internship-Manifure/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Автоматически мигрировать таблицы
	err = db.AutoMigrate(&model.ProductStatistics{}, &model.UserStatistics{})
	require.NoError(t, err)

	return db
}

func TestHandleProductUpdate(t *testing.T) {
	db := setupTestDB(t)
	handler := analytics.NewAnalyticsHandler(db)

	// Генерируем валидный UUID
	productID := uuid.New().String()

	productUpdate := map[string]string{"id": productID}
	message, err := json.Marshal(productUpdate)
	require.NoError(t, err)

	err = handler.HandleProductUpdate(message)
	require.NoError(t, err, "handleProductUpdate should not return an error")

	// Проверка, что статистика была обновлена
	var stats model.ProductStatistics
	err = db.Where("product_id = ?", productID).First(&stats).Error
	require.NoError(t, err)
	require.Equal(t, 1, stats.UpdateCount)
	log.Printf("%v", stats)
}

func TestHandleUserUpdate(t *testing.T) {
	db := setupTestDB(t)
	handler := analytics.NewAnalyticsHandler(db)

	// Генерируем валидный UUID
	userID := uuid.New().String()

	userUpdate := map[string]string{"id": userID}
	message, err := json.Marshal(userUpdate)
	require.NoError(t, err)

	err = handler.HandleUserUpdate(message)
	require.NoError(t, err, "handleUserUpdate should not return an error")

	// Проверка, что статистика была обновлена
	var stats model.UserStatistics
	err = db.Where("user_id = ?", userID).First(&stats).Error
	require.NoError(t, err)
	require.Equal(t, 1, stats.ActivityCount)
	log.Printf("%v", stats)
}
