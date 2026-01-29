package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-gin-testing-todos/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockTodoService is a mock implementation of service.TodoService
type MockTodoService struct {
	mock.Mock
}

func (m *MockTodoService) Create(todo *model.Todo) error {
	args := m.Called(todo)
	return args.Error(0)
}

func (m *MockTodoService) GetAll() ([]model.Todo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Todo), args.Error(1)
}

func (m *MockTodoService) GetByID(id string) (*model.Todo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Todo), args.Error(1)
}

func (m *MockTodoService) Update(id string, todo *model.Todo) error {
	args := m.Called(id, todo)
	return args.Error(0)
}

func (m *MockTodoService) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

func TestCreateTodo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.POST("/todos", controller.Create)

		todo := model.Todo{Title: "Test Todo", Completed: false}
		mockService.On("Create", mock.AnythingOfType("*model.Todo")).Return(nil)

		jsonValue, _ := json.Marshal(todo)
		req, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("BindingError", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.POST("/todos", controller.Create)

		req, _ := http.NewRequest("POST", "/todos", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.POST("/todos", controller.Create)

		todo := model.Todo{Title: "Test Todo", Completed: false}
		mockService.On("Create", mock.AnythingOfType("*model.Todo")).Return(errors.New("db error"))

		jsonValue, _ := json.Marshal(todo)
		req, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestGetAllTodos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.GET("/todos", controller.GetAll)

		todos := []model.Todo{
			{Title: "Todo 1", Completed: false},
			{Title: "Todo 2", Completed: true},
		}
		mockService.On("GetAll").Return(todos, nil)

		req, _ := http.NewRequest("GET", "/todos", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var responseTodos []model.Todo
		json.Unmarshal(w.Body.Bytes(), &responseTodos)
		assert.Len(t, responseTodos, 2)
		
		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.GET("/todos", controller.GetAll)

		mockService.On("GetAll").Return(nil, errors.New("db error"))

		req, _ := http.NewRequest("GET", "/todos", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestGetTodoByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.GET("/todos/:id", controller.GetByID)

		id := primitive.NewObjectID()
		todo := &model.Todo{ID: id, Title: "Test Todo"}
		mockService.On("GetByID", id.Hex()).Return(todo, nil)

		req, _ := http.NewRequest("GET", "/todos/"+id.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.GET("/todos/:id", controller.GetByID)

		id := primitive.NewObjectID()
		mockService.On("GetByID", id.Hex()).Return(nil, errors.New("not found"))

		req, _ := http.NewRequest("GET", "/todos/"+id.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUpdateTodo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.PUT("/todos/:id", controller.Update)

		id := primitive.NewObjectID()
		todo := model.Todo{Title: "Updated Todo", Completed: true}
		mockService.On("Update", id.Hex(), mock.AnythingOfType("*model.Todo")).Return(nil)

		jsonValue, _ := json.Marshal(todo)
		req, _ := http.NewRequest("PUT", "/todos/"+id.Hex(), bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

    t.Run("BadRequest", func(t *testing.T) {
        mockService := new(MockTodoService)
        controller := NewTodoController(mockService)
        router := setupRouter()
        router.PUT("/todos/:id", controller.Update)

        req, _ := http.NewRequest("PUT", "/todos/123", bytes.NewBufferString("invalid json"))
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })

	t.Run("ServiceError", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.PUT("/todos/:id", controller.Update)

		id := primitive.NewObjectID()
		todo := model.Todo{Title: "Updated Todo", Completed: true}
		mockService.On("Update", id.Hex(), mock.AnythingOfType("*model.Todo")).Return(errors.New("update failed"))

		jsonValue, _ := json.Marshal(todo)
		req, _ := http.NewRequest("PUT", "/todos/"+id.Hex(), bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestDeleteTodo(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.DELETE("/todos/:id", controller.Delete)

		id := primitive.NewObjectID()
		mockService.On("Delete", id.Hex()).Return(nil)

		req, _ := http.NewRequest("DELETE", "/todos/"+id.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService := new(MockTodoService)
		controller := NewTodoController(mockService)
		router := setupRouter()
		router.DELETE("/todos/:id", controller.Delete)

		id := primitive.NewObjectID()
		mockService.On("Delete", id.Hex()).Return(errors.New("delete failed"))

		req, _ := http.NewRequest("DELETE", "/todos/"+id.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}
