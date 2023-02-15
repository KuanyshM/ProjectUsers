package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

type SaltResponse struct {
	Salt string `json:"salt"`
}

func main() {
	r := chi.NewRouter()
	// добавляем middleware для логирования
	r.Use(middleware.Logger)

	r.Post("/generate-salt", func(w http.ResponseWriter, r *http.Request) {
		salt := GenerateSalt(12)
		response := SaltResponse{Salt: salt}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Ошибка при обработке ответа", http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func GenerateSalt(length int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())
	salt := make([]rune, length)
	for i := range salt {
		salt[i] = letters[rand.Intn(len(letters))]
	}
	return string(salt)
}
