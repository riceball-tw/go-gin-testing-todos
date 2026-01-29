package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go-gin-testing-todos/internal/controller"
	"go-gin-testing-todos/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// MongoDB Connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	db := client.Database("todo_db")

	// Dependencies
	todoService := service.NewTodoService(db)
	todoController := controller.NewTodoController(todoService)

	// Router
	r := gin.Default()
	
	// Create a new router group for /todos
    // this wraps the handlers to make them compatible with gin.HandlerFunc
    // or we can just pass the methods directly if they match the signature
	
    // Handlers
	r.POST("/todos", todoController.Create)
	r.GET("/todos", todoController.GetAll)
	r.GET("/todos/:id", todoController.GetByID)
	r.PUT("/todos/:id", todoController.Update)
	r.DELETE("/todos/:id", todoController.Delete)

    // Health check
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "pong",
        })
    })

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
