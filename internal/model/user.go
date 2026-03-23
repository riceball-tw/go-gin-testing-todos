package model

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system with FIDO2 credentials
type User struct {
	ID          primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	Name        string                `bson:"name" json:"name"`
	DisplayName string                `bson:"display_name" json:"displayName"`
	Credentials []webauthn.Credential `bson:"credentials" json:"-"`
}

// WebAuthnID returns the user's ID
func (u *User) WebAuthnID() []byte {
	return []byte(u.ID.Hex())
}

// WebAuthnName returns the user's name
func (u *User) WebAuthnName() string {
	return u.Name
}

// WebAuthnDisplayName returns the user's display name
func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnIcon is not implemented
func (u *User) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials returns credentials owned by the user
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}
