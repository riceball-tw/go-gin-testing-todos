package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-gin-testing-todos/internal/controller"
	"go-gin-testing-todos/internal/model"
	"go-gin-testing-todos/internal/service"
	"go-gin-testing-todos/tests/integration/testhelper"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestEditTodoIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	container, db, err := testhelper.SetupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	svc := service.NewTodoService(db)
	ctrl := controller.NewTodoController(svc)
	router := gin.Default()
	router.PUT("/todos/:id", ctrl.Update)
	router.GET("/todos/:id", ctrl.GetByID)

	// Setup: Create todo via direct DB insert
	todo := &model.Todo{
		Title:     "Original title",
		Completed: false,
		CreatedAt: time.Now(),
	}
	collection := db.Collection("todos")
	res, err := collection.InsertOne(ctx, todo)
	assert.NoError(t, err)
	id := res.InsertedID.(primitive.ObjectID)

	// Test: Update via API
	updateData := map[string]interface{}{
		"title":     "Updated title",
		"completed": true,
	}
	jsonData, _ := json.Marshal(updateData)
	
	updateRes := httptest.NewRecorder()
	updateReq, _ := http.NewRequest("PUT", "/todos/"+id.Hex(), bytes.NewBuffer(jsonData))
	updateReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(updateRes, updateReq)

	assert.Equal(t, http.StatusOK, updateRes.Code)

	// Verify: Check update via API
	getRes := httptest.NewRecorder()
	getReq, _ := http.NewRequest("GET", "/todos/"+id.Hex(), nil)
	router.ServeHTTP(getRes, getReq)

	assert.Equal(t, http.StatusOK, getRes.Code)
	
	var updatedTodo model.Todo
	err = json.Unmarshal(getRes.Body.Bytes(), &updatedTodo)
	assert.NoError(t, err)
	assert.Equal(t, "Updated title", updatedTodo.Title)
	assert.True(t, updatedTodo.Completed)
}
