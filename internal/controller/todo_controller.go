package controller

import (
	"net/http"

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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Create(&todo); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, todo)
}

func (c *TodoController) GetAll(ctx *gin.Context) {
	todos, err := c.service.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if todos == nil {
		todos = []model.Todo{}
	}
	ctx.JSON(http.StatusOK, todos)
}

func (c *TodoController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	todo, err := c.service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, todo)
}

func (c *TodoController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var todo model.Todo
	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Update(id, &todo); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func (c *TodoController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.service.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
