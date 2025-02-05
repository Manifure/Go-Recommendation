package product_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Go-internship-Manifure/internal/handlers/product"
	"Go-internship-Manifure/internal/kafka"
	"Go-internship-Manifure/internal/model"
	"github.com/gorilla/mux"
)

func TestAddProduct(t *testing.T) {
	mockProducer := &kafka.MockProducer{}
	handler := product.NewProductHandler(mockProducer)

	// Данные нового продукта
	productData := model.Product{Name: "Test Product", Price: 100.50}

	body, err := json.Marshal(productData)
	if err != nil {
		t.Fatalf("Failed to marshal product data: %s", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/products/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()
	handler.AddProduct(responseRecorder, req)

	if responseRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status code: got %v, want %v", responseRecorder.Code, http.StatusCreated)
	}

	var createdProduct model.Product

	err = json.NewDecoder(responseRecorder.Body).Decode(&createdProduct)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	// Проверка, что продукт добавлен в обработчик
	if _, exists := handler.Products[createdProduct.ID]; !exists {
		t.Fatalf("product not saved in handler: %+v", createdProduct)
	}

	// Проверка данных продукта
	if createdProduct.Name != productData.Name || createdProduct.Price != productData.Price {
		t.Fatalf("unexpected product data: got %+v, want %+v", createdProduct, productData)
	}
}

func TestDeleteProduct(t *testing.T) {
	mockProducer := &kafka.MockProducer{}
	handler := product.NewProductHandler(mockProducer)

	// Создание продукта для теста
	productID := "test-product-id"
	handler.Products[productID] = model.Product{ID: productID, Name: "Test Product", Price: 100.50, Views: 100}

	req := httptest.NewRequest(http.MethodDelete, "/products/"+productID, nil)
	responseRecorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/products/{id}", handler.DeleteProduct).Methods(http.MethodDelete)
	router.ServeHTTP(responseRecorder, req)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %v, want %v", responseRecorder.Code, http.StatusOK)
	}

	// Проверка, что продукт удален из обработчика
	if _, exists := handler.Products[productID]; exists {
		t.Fatalf("product was not deleted: %+v", handler.Products[productID])
	}
}

func TestUpdateProduct(t *testing.T) {
	mockProducer := &kafka.MockProducer{}
	handler := product.NewProductHandler(mockProducer)

	// Создание продукта для теста
	productID := "test-product-id"
	handler.Products[productID] = model.Product{ID: productID, Name: "Old Product", Price: 100.50, Views: 100}

	updatedProductData := model.Product{Name: "Updated Product", Price: 150.75, Views: 150}

	body, err := json.Marshal(updatedProductData)
	if err != nil {
		t.Fatalf("failed to marshal updated product data: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/products/"+productID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	responseRecorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/products/{id}", handler.UpdateProduct).Methods(http.MethodPut)
	router.ServeHTTP(responseRecorder, req)

	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %v, want %v", responseRecorder.Code, http.StatusOK)
	}

	// Проверка, что продукт обновлен
	updatedProduct := handler.Products[productID]
	if updatedProduct.Name != updatedProductData.Name || updatedProduct.Price != updatedProductData.Price {
		t.Fatalf("product not updated correctly: got %+v, want %+v", updatedProduct, updatedProductData)
	}
}
