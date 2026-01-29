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

// Helper to setup controller with mock service
func setupController() (*MockTodoService, *TodoController, *gin.Engine) {
	mockService := new(MockTodoService)
	controller := NewTodoController(mockService)
	router := setupRouter()
	return mockService, controller, router
}

// Helper to make HTTP request and return response
func makeRequest(router *gin.Engine, method, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonValue, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonValue)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	
	req, _ := http.NewRequest(method, url, reqBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCreateTodo(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func(*MockTodoService)
		expectedStatus int
	}{
		{
			name: "Success",
			body: model.Todo{Title: "Test Todo", Completed: false},
			mockSetup: func(m *MockTodoService) {
				m.On("Create", mock.AnythingOfType("*model.Todo")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "BindingError",
			body:           "invalid json",
			mockSetup:      func(m *MockTodoService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ServiceError",
			body: model.Todo{Title: "Test Todo", Completed: false},
			mockSetup: func(m *MockTodoService) {
				m.On("Create", mock.AnythingOfType("*model.Todo")).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN
			mockService, controller, router := setupController()
			router.POST("/todos", controller.Create)
			tt.mockSetup(mockService)

			// WHEN
			var w *httptest.ResponseRecorder
			if tt.name == "BindingError" {
				req, _ := http.NewRequest("POST", "/todos", bytes.NewBufferString("invalid json"))
				w = httptest.NewRecorder()
				router.ServeHTTP(w, req)
			} else {
				w = makeRequest(router, "POST", "/todos", tt.body)
			}

			// THEN
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetAllTodos(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockTodoService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "Success",
			mockSetup: func(m *MockTodoService) {
				todos := []model.Todo{
					{Title: "Todo 1", Completed: false},
					{Title: "Todo 2", Completed: true},
				}
				m.On("GetAll").Return(todos, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "ServiceError",
			mockSetup: func(m *MockTodoService) {
				m.On("GetAll").Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN
			mockService, controller, router := setupController()
			router.GET("/todos", controller.GetAll)
			tt.mockSetup(mockService)

			// WHEN
			w := makeRequest(router, "GET", "/todos", nil)

			// THEN
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedCount > 0 {
				var responseTodos []model.Todo
				json.Unmarshal(w.Body.Bytes(), &responseTodos)
				assert.Len(t, responseTodos, tt.expectedCount)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetTodoByID(t *testing.T) {
	id := primitive.NewObjectID()
	
	tests := []struct {
		name           string
		mockSetup      func(*MockTodoService)
		expectedStatus int
	}{
		{
			name: "Success",
			mockSetup: func(m *MockTodoService) {
				todo := &model.Todo{ID: id, Title: "Test Todo"}
				m.On("GetByID", id.Hex()).Return(todo, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ServiceError",
			mockSetup: func(m *MockTodoService) {
				m.On("GetByID", id.Hex()).Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN
			mockService, controller, router := setupController()
			router.GET("/todos/:id", controller.GetByID)
			tt.mockSetup(mockService)

			// WHEN
			w := makeRequest(router, "GET", "/todos/"+id.Hex(), nil)

			// THEN
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateTodo(t *testing.T) {
	id := primitive.NewObjectID()
	
	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func(*MockTodoService)
		expectedStatus int
	}{
		{
			name: "Success",
			body: model.Todo{Title: "Updated Todo", Completed: true},
			mockSetup: func(m *MockTodoService) {
				m.On("Update", id.Hex(), mock.AnythingOfType("*model.Todo")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "BadRequest",
			body:           "invalid json",
			mockSetup:      func(m *MockTodoService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ServiceError",
			body: model.Todo{Title: "Updated Todo", Completed: true},
			mockSetup: func(m *MockTodoService) {
				m.On("Update", id.Hex(), mock.AnythingOfType("*model.Todo")).Return(errors.New("update failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN
			mockService, controller, router := setupController()
			router.PUT("/todos/:id", controller.Update)
			tt.mockSetup(mockService)

			// WHEN
			var w *httptest.ResponseRecorder
			if tt.name == "BadRequest" {
				req, _ := http.NewRequest("PUT", "/todos/123", bytes.NewBufferString("invalid json"))
				w = httptest.NewRecorder()
				router.ServeHTTP(w, req)
			} else {
				w = makeRequest(router, "PUT", "/todos/"+id.Hex(), tt.body)
			}

			// THEN
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteTodo(t *testing.T) {
	id := primitive.NewObjectID()
	
	tests := []struct {
		name           string
		mockSetup      func(*MockTodoService)
		expectedStatus int
	}{
		{
			name: "Success",
			mockSetup: func(m *MockTodoService) {
				m.On("Delete", id.Hex()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ServiceError",
			mockSetup: func(m *MockTodoService) {
				m.On("Delete", id.Hex()).Return(errors.New("delete failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN
			mockService, controller, router := setupController()
			router.DELETE("/todos/:id", controller.Delete)
			tt.mockSetup(mockService)

			// WHEN
			w := makeRequest(router, "DELETE", "/todos/"+id.Hex(), nil)

			// THEN
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
