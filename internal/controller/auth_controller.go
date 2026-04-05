package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go-gin-testing-todos/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ac *AuthController) BeginRegistration(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := ac.authService.GetOrCreateUser(ctx, username, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get or create user"})
		return
	}

	options, sessionData, err := ac.authService.WebAuthn.BeginRegistration(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin registration"})
		return
	}

	// Save session data
	session := sessions.Default(c)
	sessionDataJSON, _ := json.Marshal(sessionData)
	session.Set("webauthn_session", sessionDataJSON)
	session.Set("webauthn_username", username)
	session.Save()

	c.JSON(http.StatusOK, options)
}

func (ac *AuthController) FinishRegistration(c *gin.Context) {
	session := sessions.Default(c)
	sessionDataStr := session.Get("webauthn_session")
	username := session.Get("webauthn_username")

	if sessionDataStr == nil || username == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session expired or invalid"})
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionDataStr.([]byte), &sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid session data format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := ac.authService.GetUserByUsername(ctx, username.(string))
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	credential, err := ac.authService.WebAuthn.FinishRegistration(user, sessionData, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to finish registration: " + err.Error()})
		return
	}

	if credential != nil {
		user.Credentials = append(user.Credentials, *credential)
		err = ac.authService.UpdateUser(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear credentials"})
			return
		}
	}

	// Login the user implicitly after registration
	session.Set("logged_in_user_id", user.ID.Hex())
	session.Delete("webauthn_session")
	session.Delete("webauthn_username")
	session.Save()

	c.JSON(http.StatusOK, gin.H{"status": "registration successful!"})
}

func (ac *AuthController) BeginLogin(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := ac.authService.GetUserByUsername(ctx, username)
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	options, sessionData, err := ac.authService.WebAuthn.BeginLogin(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin login"})
		return
	}

	session := sessions.Default(c)
	sessionDataJSON, _ := json.Marshal(sessionData)
	session.Set("webauthn_session", sessionDataJSON)
	session.Set("webauthn_username", username)
	session.Save()

	c.JSON(http.StatusOK, options)
}

func (ac *AuthController) FinishLogin(c *gin.Context) {
	session := sessions.Default(c)
	sessionDataStr := session.Get("webauthn_session")
	username := session.Get("webauthn_username")

	if sessionDataStr == nil || username == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session expired or invalid"})
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionDataStr.([]byte), &sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid session data format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := ac.authService.GetUserByUsername(ctx, username.(string))
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	credential, err := ac.authService.WebAuthn.FinishLogin(user, sessionData, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to finish login: " + err.Error()})
		return
	}

	// Update the credential's sign count if needed (optional, depends on implementation)
    // For simplicity, we skip updating the credential sign count here
	_ = credential 

	// Login successful
	session.Set("logged_in_user_id", user.ID.Hex())
	session.Delete("webauthn_session")
	session.Delete("webauthn_username")
	session.Save()

	c.JSON(http.StatusOK, gin.H{"status": "login successful!"})
}

// CurrentUser returns the logged-in user
func (ac *AuthController) CurrentUser(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("logged_in_user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

func (ac *AuthController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}
