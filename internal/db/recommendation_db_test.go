package db_test

import (
	mock_db "Go-internship-Manifure/internal/db/mocks"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRecommendationDatabaseClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_db.NewMockDatabaseRecommendationInterface(ctrl)

	mockDB.EXPECT().CloseRecommendationDB().Return(nil)

	err := mockDB.CloseRecommendationDB()
	assert.NoError(t, err)
}

func TestRecommendationDatabaseMigrationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_db.NewMockDatabaseRecommendationInterface(ctrl)

	mockDB.EXPECT().MigrateRecommendationModels().Return(errors.New("migration error"))

	err := mockDB.MigrateRecommendationModels()
	assert.EqualError(t, err, "migration error")
}
