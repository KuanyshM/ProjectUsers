package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"regexp"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	Email string `json:"email"`
}

type UserSalt struct {
	Salt string `json:"salt"`
}

func main() {
	// создаем роутер Go-chi
	r := chi.NewRouter()

	// добавляем middleware для логирования
	r.Use(middleware.Logger)

	// создаем MongoDB клиент
	clientOptions := options.Client().ApplyURI("mongodb://root:example@mongo:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("testdb").Collection("users")

	r.Post("/create-user", func(w http.ResponseWriter, r *http.Request) {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}

		//user := r.Context().Value("user").(*User)

		if !isValidEmail(user.Email) {
			http.Error(w, "Invalid email format", http.StatusBadRequest)
			return
		}
		if isEmailExist(user.Email, collection) {
			http.Error(w, "Email already exists", http.StatusBadRequest)
			return
		}

		// обращаемся к сервису 1
		salt, err := getUserSaltFromService1("http://service1:8080/generate-salt")
		if err != nil {
			http.Error(w, "Failed to get user salt", http.StatusInternalServerError)
			return
		}

		hash := md5.Sum([]byte(user.Password + salt))

		err = saveUserToMongo(User{Email: user.Email, Password: hex.EncodeToString(hash[:])}, salt, collection)
		if err != nil {
			http.Error(w, "Failed to save user to database", http.StatusInternalServerError)
			return
		}

		response := UserResponse{
			Email: user.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	r.Get("/get-user/{email}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("asd")
		email := chi.URLParam(r, "email")
		fmt.Println(email)
		user, err := getUserByEmail(email, collection)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		response := UserResponse{
			Email: user.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	})

	r.Get("/get-test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("asd")
		type ResponseTest struct {
			Name string
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ResponseTest{Name: "hi from server"})

	})

	// запускаем сервер на порту 8090
	fmt.Println("Сервер запущен на http://localhost:8090")
	http.ListenAndServe(":8090", r)
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "user", &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	r := regexp.MustCompile(pattern)
	if r.MatchString(email) {
		return true
	} else {
		return false
	}
}

func isEmailExist(email string, collection *mongo.Collection) bool {
	filter := bson.M{"email": email}
	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	return count > 0
}

func getUserSaltFromService1(url string) (string, error) {
	requestBody, err := json.Marshal(map[string]string{})
	if err != nil {
		return "", fmt.Errorf("ошибка при создании тела запроса: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	var userSqlt UserSalt
	err = json.NewDecoder(resp.Body).Decode(&userSqlt)
	if err != nil {
		return "", fmt.Errorf("ошибка при обработке ответа: %v", err)
	}

	return userSqlt.Salt, nil
}

func saveUserToMongo(user User, salt string, collection *mongo.Collection) error {
	userToSave := bson.M{
		"email":    user.Email,
		"salt":     salt,
		"password": user.Password,
	}
	_, err := collection.InsertOne(context.Background(), userToSave)
	return err
}
func getUserByEmail(email string, collection *mongo.Collection) (User, error) {
	var user User
	err := collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return user, err
	}
	return user, nil
}
