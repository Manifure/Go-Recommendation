package product

import (
	"encoding/json"
	"log"
	"net/http"

	"Go-internship-Manifure/internal/kafka"
	"Go-internship-Manifure/internal/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	Products      map[string]model.Product
	KafkaProducer kafka.ProducerInterface
}

// Создание нового обработчика продуктов.
func NewProductHandler(kafkaProducer kafka.ProducerInterface) *Handler {
	return &Handler{
		Products:      make(map[string]model.Product),
		KafkaProducer: kafkaProducer,
	}
}

// Добавление нового продукта.
func (ph *Handler) AddProduct(w http.ResponseWriter, r *http.Request) {
	var product model.Product

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if product.Name == "" {
		http.Error(w, "Product name is required", http.StatusBadRequest)

		return
	}

	product.ID = uuid.New().String()
	ph.Products[product.ID] = product

	message, err := json.Marshal(product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = ph.KafkaProducer.Produce(string(message))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}
}

// Удаление продукта.
func (ph *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if _, ok := ph.Products[id]; !ok {
		http.Error(w, "Product not found", http.StatusNotFound)

		return
	}

	delete(ph.Products, id)
	log.Printf("Product %v deleted", id)
}

// Обновление продукта.
func (ph *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var UpdatedProduct model.Product

	if err := json.NewDecoder(r.Body).Decode(&UpdatedProduct); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	product, ok := ph.Products[id]
	if !ok {
		http.Error(w, "Product not found", http.StatusNotFound)

		return
	}

	if UpdatedProduct.Name != "" {
		product.Name = UpdatedProduct.Name
	}

	product.Price = UpdatedProduct.Price
	ph.Products[id] = product

	message, err := json.Marshal(product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = ph.KafkaProducer.Produce(string(message))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(UpdatedProduct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}
}
