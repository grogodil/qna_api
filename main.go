package main

import (
	"log"
	"net/http"
	"os"
	"qna_api/internal/handler"

	"github.com/gorilla/mux"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL") // из докер-compose
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database:", err)
	}
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("failed to set dialect:", err)
	}
	if err := goose.Up(sqlDB, "./migrations"); err != nil {
		log.Fatal("failed to apply migrations:", err)
	}
	log.Println("Migrations applied successfully")

	r := mux.NewRouter()
	h := handler.New(db)

	r.HandleFunc("/questions/", h.ListQuestions).Methods("GET")
	r.HandleFunc("/questions/", h.CreateQuestion).Methods("POST")
	r.HandleFunc("/questions/{id:[0-9]+}", h.GetQuestion).Methods("GET")
	r.HandleFunc("/questions/{id:[0-9]+}", h.DeleteQuestion).Methods("DELETE")
	r.HandleFunc("/questions/{id:[0-9]+}/answers/", h.CreateAnswer).Methods("POST")
	r.HandleFunc("/answers/{id:[0-9]+}", h.GetAnswer).Methods("GET")
	r.HandleFunc("/answers/{id:[0-9]+}", h.DeleteAnswer).Methods("DELETE")

	log.Println("API running on :8080")
	http.ListenAndServe(":8080", r)
}
