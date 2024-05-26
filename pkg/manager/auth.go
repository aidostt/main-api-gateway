package authManager

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// TokenManager provides logic for JWT & Refresh tokens generation and parsing.
type TokenManager interface {
	NewAccessToken(string, time.Duration, []string, string, bool) (string, error)
	Parse(accessToken string) (string, []string, bool, error)
	NewRefreshToken() (string, error)
	HexToObjectID(string) (primitive.ObjectID, error)
	ParseActivationToken(string) (string, time.Time, error)
}

type Manager struct {
	signingKey string
}

type CustomClaims struct {
	UserID    string   `json:"user_id"`
	Roles     []string `json:"roles"`
	Activated bool     `json:"activated"`
	jwt.StandardClaims
}

func NewManager(signingKey string) (*Manager, error) {
	if signingKey == "" {
		return nil, errors.New("empty signing key")
	}

	return &Manager{signingKey: signingKey}, nil
}

func (m *Manager) NewAccessToken(userID string, ttl time.Duration, roles []string, issuer string, activated bool) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		Roles:     roles,
		Activated: activated,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.signingKey))
}

// Parse taking from the payload of JWT user id and returns it in string format. Token is still returned
// in both cases, if it is expired or not.
func (m *Manager) Parse(accessToken string) (string, []string, bool, error) {
	token, err := jwt.ParseWithClaims(accessToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.signingKey), nil
	})

	if err != nil {
		var validationError *jwt.ValidationError
		if errors.As(err, &validationError) && validationError.Errors&jwt.ValidationErrorExpired != 0 {
			err = errors.New("token is expired")
		} else {
			return "", nil, false, err
		}
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", nil, false, fmt.Errorf("error getting user claims from token")
	}

	return claims.UserID, claims.Roles, claims.Activated, err
}

func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func (m *Manager) HexToObjectID(hex string) (primitive.ObjectID, error) {
	objectId, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return objectId, nil
}

func (m *Manager) ParseActivationToken(token string) (string, time.Time, error) {
	decodedToken, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", time.Time{}, err
	}
	parts := strings.Split(string(decodedToken), ".")
	if len(parts) != 2 {
		return "", time.Time{}, fmt.Errorf("invalid token format")
	}

	data := parts[0]
	expectedSignature, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", time.Time{}, err
	}

	mac := hmac.New(sha256.New, []byte(m.signingKey))
	mac.Write([]byte(data))
	actualSignature := mac.Sum(nil)

	if !hmac.Equal(expectedSignature, actualSignature) {
		return "", time.Time{}, fmt.Errorf("invalid token signature")
	}

	dataParts := strings.Split(data, ":")
	if len(dataParts) != 2 {
		return "", time.Time{}, fmt.Errorf("invalid data format")
	}

	expiryUnix, err := strconv.ParseInt(dataParts[1], 10, 64)
	if err != nil {
		return "", time.Time{}, err
	}
	expiry := time.Unix(expiryUnix, 0)
	return dataParts[0], expiry, nil
}
