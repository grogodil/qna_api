package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"qna_api/internal/model"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB создает тестовую БД в памяти
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&model.Question{}, &model.Answer{})
	assert.NoError(t, err)
	return db
}

// TestListQuestions - тест получения списка всех вопросов
func TestListQuestions(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	db.Create(&model.Question{Text: "Вопрос 1?"})
	db.Create(&model.Question{Text: "Вопрос 2?"})

	req := httptest.NewRequest("GET", "/questions/", nil)
	w := httptest.NewRecorder()

	h.ListQuestions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var questions []model.Question
	err := json.Unmarshal(w.Body.Bytes(), &questions)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(questions))
}

// TestListQuestionsEmpty - тест получения пустого списка
func TestListQuestionsEmpty(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	req := httptest.NewRequest("GET", "/questions/", nil)
	w := httptest.NewRecorder()

	h.ListQuestions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var questions []model.Question
	err := json.Unmarshal(w.Body.Bytes(), &questions)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(questions))
}

// TestCreateQuestion - тест создания нового вопроса
func TestCreateQuestion(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	body := []byte(`{"text":"Мой первый вопрос?"}`)
	req := httptest.NewRequest("POST", "/questions/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.CreateQuestion(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var question model.Question
	err := json.Unmarshal(w.Body.Bytes(), &question)
	assert.NoError(t, err)
	assert.Equal(t, "Мой первый вопрос?", question.Text)
	assert.NotZero(t, question.ID)
}

// TestCreateQuestionInvalidJSON - тест создания с невалидным JSON
func TestCreateQuestionInvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest("POST", "/questions/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.CreateQuestion(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetQuestion - тест получения вопроса со всеми ответами
func TestGetQuestion(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	question := model.Question{Text: "Тестовый вопрос?"}
	db.Create(&question)

	db.Create(&model.Answer{
		QuestionID: question.ID,
		UserID:     "user-1",
		Text:       "Ответ 1",
	})
	db.Create(&model.Answer{
		QuestionID: question.ID,
		UserID:     "user-2",
		Text:       "Ответ 2",
	})

	req := httptest.NewRequest("GET", "/questions/1", nil)
	w := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	h.GetQuestion(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result model.Question
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "Тестовый вопрос?", result.Text)
	assert.Equal(t, 2, len(result.Answers))
}

// TestGetQuestionNotFound - тест получения несуществующего вопроса
func TestGetQuestionNotFound(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	req := httptest.NewRequest("GET", "/questions/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()

	h.GetQuestion(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestDeleteQuestion - тест удаления вопроса
func TestDeleteQuestion(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	question := model.Question{Text: "Вопрос для удаления?"}
	db.Create(&question)

	req := httptest.NewRequest("DELETE", "/questions/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.DeleteQuestion(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	var q model.Question
	result := db.First(&q, 1)
	assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
}

// TestCreateAnswer - тест создания ответа
func TestCreateAnswer(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	question := model.Question{Text: "Вопрос?"}
	db.Create(&question)

	body := []byte(`{"user_id":"550e8400-e29b-41d4-a716-446655440000","text":"Мой ответ"}`)
	req := httptest.NewRequest("POST", "/questions/1/answers/", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.CreateAnswer(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var answer model.Answer
	err := json.Unmarshal(w.Body.Bytes(), &answer)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), answer.QuestionID)
	assert.Equal(t, "Мой ответ", answer.Text)
}

// TestCreateAnswerToNonExistentQuestion - тест валидации: нельзя создать ответ к несуществующему вопросу
func TestCreateAnswerToNonExistentQuestion(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	body := []byte(`{"user_id":"550e8400-e29b-41d4-a716-446655440000","text":"Ответ"}`)
	req := httptest.NewRequest("POST", "/questions/999/answers/", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()

	h.CreateAnswer(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestMultipleAnswersFromSameUser - тест валидации: один пользователь может оставить несколько ответов на один вопрос
func TestMultipleAnswersFromSameUser(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	question := model.Question{Text: "Вопрос?"}
	db.Create(&question)

	userID := "550e8400-e29b-41d4-a716-446655440000"

	body1 := []byte(`{"user_id":"` + userID + `","text":"Первый ответ"}`)
	req1 := httptest.NewRequest("POST", "/questions/1/answers/", bytes.NewReader(body1))
	req1 = mux.SetURLVars(req1, map[string]string{"id": "1"})
	w1 := httptest.NewRecorder()
	h.CreateAnswer(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	body2 := []byte(`{"user_id":"` + userID + `","text":"Второй ответ"}`)
	req2 := httptest.NewRequest("POST", "/questions/1/answers/", bytes.NewReader(body2))
	req2 = mux.SetURLVars(req2, map[string]string{"id": "1"})
	w2 := httptest.NewRecorder()
	h.CreateAnswer(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)

	var q model.Question
	db.Preload("Answers").First(&q, 1)
	assert.Equal(t, 2, len(q.Answers))
}

// TestGetAnswer - тест получения конкретного ответа
func TestGetAnswer(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	question := model.Question{Text: "Вопрос?"}
	db.Create(&question)

	answer := model.Answer{
		QuestionID: question.ID,
		UserID:     "user-1",
		Text:       "Мой ответ",
	}
	db.Create(&answer)

	req := httptest.NewRequest("GET", "/answers/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.GetAnswer(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result model.Answer
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "Мой ответ", result.Text)
}

// TestGetAnswerNotFound - тест получения несуществующего ответа
func TestGetAnswerNotFound(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	req := httptest.NewRequest("GET", "/answers/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	w := httptest.NewRecorder()

	h.GetAnswer(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestDeleteAnswer - тест удаления ответа
func TestDeleteAnswer(t *testing.T) {
	db := setupTestDB(t)
	h := New(db)

	question := model.Question{Text: "Вопрос?"}
	db.Create(&question)

	answer := model.Answer{
		QuestionID: question.ID,
		UserID:     "user-1",
		Text:       "Ответ",
	}
	db.Create(&answer)

	req := httptest.NewRequest("DELETE", "/answers/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.DeleteAnswer(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	var a model.Answer
	result := db.First(&a, 1)
	assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
}
