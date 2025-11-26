package handler

import (
	"encoding/json"
	"net/http"
	"qna_api/internal/model"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Handler { return &Handler{db: db} }

// Вопросы

func (h *Handler) ListQuestions(w http.ResponseWriter, r *http.Request) {
	var questions []model.Question
	h.db.Preload("Answers").Find(&questions)
	json.NewEncoder(w).Encode(questions)
}

func (h *Handler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	var q model.Question
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	h.db.Create(&q)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(q)
}

func (h *Handler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var q model.Question
	if err := h.db.Preload("Answers").First(&q, id).Error; err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(q)
}

func (h *Handler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	h.db.Delete(&model.Question{}, id)
	w.WriteHeader(http.StatusNoContent)
}

// Ответы

func (h *Handler) CreateAnswer(w http.ResponseWriter, r *http.Request) {
	qid, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Bad Question ID", http.StatusBadRequest)
		return
	}
	var ans model.Answer
	if err := json.NewDecoder(r.Body).Decode(&ans); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	var q model.Question
	if err := h.db.First(&q, qid).Error; err != nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}
	ans.QuestionID = uint(qid)
	h.db.Create(&ans)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ans)
}

func (h *Handler) GetAnswer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var ans model.Answer
	if err := h.db.First(&ans, id).Error; err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(ans)
}

func (h *Handler) DeleteAnswer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	h.db.Delete(&model.Answer{}, id)
	w.WriteHeader(http.StatusNoContent)
}
