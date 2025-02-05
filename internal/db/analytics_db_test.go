package db_test

import (
	mock_db "Go-internship-Manifure/internal/db/mocks"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAnalyticsDatabaseClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_db.NewMockDatabaseAnalyticsInterface(ctrl)

	mockDB.EXPECT().CloseAnalyticsDB().Return(nil)

	err := mockDB.CloseAnalyticsDB()
	assert.NoError(t, err)
}

func TestAnalyticsDatabaseMigrationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_db.NewMockDatabaseAnalyticsInterface(ctrl)

	mockDB.EXPECT().MigrateAnalyticsModels().Return(errors.New("migration error"))

	err := mockDB.MigrateAnalyticsModels()
	assert.EqualError(t, err, "migration error")
}
