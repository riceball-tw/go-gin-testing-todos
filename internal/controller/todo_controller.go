package controller

import (
	"log/slog"
	"net/http"

	"go-gin-testing-todos/internal/logger"
	"go-gin-testing-todos/internal/model"
	"go-gin-testing-todos/internal/service"

	"github.com/gin-gonic/gin"
)

type TodoController struct {
	service service.TodoService
}

func NewTodoController(s service.TodoService) *TodoController {
	return &TodoController{service: s}
}

func (c *TodoController) Create(ctx *gin.Context) {
	var todo model.Todo
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.AddContext(ctx, slog.String("todo_title", todo.Title))

	if err := c.service.Create(&todo); err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.AddContext(ctx, slog.String("todo_id", todo.ID.Hex()))
	ctx.JSON(http.StatusCreated, todo)
}

func (c *TodoController) GetAll(ctx *gin.Context) {
	todos, err := c.service.GetAll()
	if err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if todos == nil {
		todos = []model.Todo{}
	}

	logger.AddContext(ctx, slog.Int("todos_count", len(todos)))
	ctx.JSON(http.StatusOK, todos)
}

func (c *TodoController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	logger.AddContext(ctx, slog.String("todo_id", id))

	todo, err := c.service.GetByID(id)
	if err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, todo)
}

func (c *TodoController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	logger.AddContext(ctx, slog.String("todo_id", id))

	var todo model.Todo
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Update(id, &todo); err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func (c *TodoController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	logger.AddContext(ctx, slog.String("todo_id", id))

	if err := c.service.Delete(id); err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
