package integration

import (
	"context"
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

func TestDeleteTodoIntegration(t *testing.T) {
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
	router.DELETE("/todos/:id", ctrl.Delete)
	router.GET("/todos/:id", ctrl.GetByID)

	// Setup: Create todo via direct DB insert
	todo := &model.Todo{
		Title:     "To be deleted",
		Completed: false,
		CreatedAt: time.Now(),
	}
	collection := db.Collection("todos")
	res, err := collection.InsertOne(ctx, todo)
	assert.NoError(t, err)
	id := res.InsertedID.(primitive.ObjectID)

	// Test: Delete via API
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/todos/"+id.Hex(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify: Check deletion via API
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/todos/"+id.Hex(), nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusInternalServerError, w2.Code)
}
