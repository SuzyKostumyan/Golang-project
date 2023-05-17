package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Promotion struct {
	ID             string `json:"id"`
	Price          string `json:"price"`
	ExpirationDate string `json:"expiration_date"`
}

type PromotionStorage struct {
	promotions map[string]Promotion
	mu         sync.RWMutex
}

func NewPromotionStorage() *PromotionStorage {
	return &PromotionStorage{
		promotions: make(map[string]Promotion),
		mu:         sync.RWMutex{},
	}
}

func (s *PromotionStorage) AddPromotion(promotion Promotion) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.promotions[promotion.ID] = promotion
}

func (s *PromotionStorage) GetPromotion(id string) (Promotion, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	promotion, found := s.promotions[id]
	return promotion, found
}

func main() {
	storage := NewPromotionStorage()
	loadPromotions(storage, "promotions.csv")

	http.HandleFunc("/promotions/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/promotions/")

		if len(id) == 0 {
			http.NotFound(w, r)
			return
		}

		promotion, found := storage.GetPromotion(id)
		if !found {
			http.NotFound(w, r)
			return
		}

		jsonBytes, err := json.Marshal(promotion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	log.Fatal(http.ListenAndServe(":1321", nil))
}

func loadPromotions(storage *PromotionStorage, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			continue
		}

		promotion := Promotion{
			ID:             record[0],
			Price:          record[1],
			ExpirationDate: record[2],
		}

		storage.AddPromotion(promotion)
	}
}
