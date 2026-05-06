package service

import (
	"context"
	"log/slog"
	"time"

	"go-gin-testing-todos/internal/logger"
	"go-gin-testing-todos/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TodoService interface {
	Create(todo *model.Todo) error
	GetAll() ([]model.Todo, error)
	GetByID(id string) (*model.Todo, error)
	Update(id string, todo *model.Todo) error
	Delete(id string) error
}

type todoService struct {
	collection *mongo.Collection
}

func NewTodoService(db *mongo.Database) TodoService {
	return &todoService{
		collection: db.Collection("todos"),
	}
}

func (s *todoService) Create(todo *model.Todo) error {
	todo.CreatedAt = time.Now()

	logger.Log.Debug("Creating todo", slog.String("collection", "todos"), slog.String("operation", "InsertOne"), slog.Any("document", todo))

	res, err := s.collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}
	todo.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (s *todoService) GetAll() ([]model.Todo, error) {
	logger.Log.Debug("Get All Todos", slog.String("collection", "todos"), slog.String("operation", "Find"), slog.Any("filter", bson.M{}))

	cursor, err := s.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var todos []model.Todo
	if err = cursor.All(context.Background(), &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func (s *todoService) GetByID(id string) (*model.Todo, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("executing database query", slog.String("collection", "todos"), slog.String("operation", "FindOne"), slog.Any("filter", bson.M{"_id": oid}))

	var todo model.Todo
	err = s.collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&todo)
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

func (s *todoService) Update(id string, todo *model.Todo) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": bson.M{
			"title":     todo.Title,
			"completed": todo.Completed,
		},
	}

	logger.Log.Debug("Updating Todo", slog.String("collection", "todos"), slog.String("operation", "UpdateOne"), slog.Any("filter", bson.M{"_id": oid}), slog.Any("update", update))

	_, err = s.collection.UpdateOne(context.Background(), bson.M{"_id": oid}, update)
	return err
}

func (s *todoService) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	logger.Log.Debug("Deleting Todo", slog.String("collection", "todos"), slog.String("operation", "DeleteOne"), slog.Any("filter", bson.M{"_id": oid}))

	_, err = s.collection.DeleteOne(context.Background(), bson.M{"_id": oid})
	return err
}
