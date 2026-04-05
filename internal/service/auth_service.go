package service

import (
	"context"
	"errors"
	"fmt"

	"go-gin-testing-todos/internal/model"

	"github.com/go-webauthn/webauthn/webauthn"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	collection *mongo.Collection
	WebAuthn   *webauthn.WebAuthn
}

func NewAuthService(db *mongo.Database) (*AuthService, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: "Go Gin Todos",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	}

	wa, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn : %v", err)
	}

	return &AuthService{
		collection: db.Collection("users"),
		WebAuthn:   wa,
	}, nil
}

func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := s.collection.FindOne(ctx, bson.M{"name": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // User not found
		}
		return nil, err
	}
	// Initializing credentials slice if nil to prevent issues
	if user.Credentials == nil {
		user.Credentials = []webauthn.Credential{}
	}
	return &user, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var user model.User
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	if user.Credentials == nil {
		user.Credentials = []webauthn.Credential{}
	}
	return &user, nil
}

func (s *AuthService) CreateUser(ctx context.Context, username, displayName string) (*model.User, error) {
	user := &model.User{
		ID:          primitive.NewObjectID(),
		Name:        username,
		DisplayName: displayName,
		Credentials: []webauthn.Credential{},
	}

	_, err := s.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) UpdateUser(ctx context.Context, user *model.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{
		"credentials": user.Credentials,
	}}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// GetOrCreateUser looks up a user, if not found creates a new one
func (s *AuthService) GetOrCreateUser(ctx context.Context, username, displayName string) (*model.User, error) {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return s.CreateUser(ctx, username, displayName)
	}
	return user, nil
}
