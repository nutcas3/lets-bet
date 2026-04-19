package auth

import (
	"fmt"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	secretKey    []byte
	issuer       string
	expiryHours  int
	refreshHours int
}

// Claims represents JWT claims
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	CountryCode string    `json:"country_code"`
	IsVerified  bool      `json:"is_verified"`
	jwt.RegisteredClaims
}

func NewJWTService(secretKey, issuer string, expiryHours, refreshHours int) *JWTService {
	return &JWTService{
		secretKey:    []byte(secretKey),
		issuer:       issuer,
		expiryHours:  expiryHours,
		refreshHours: refreshHours,
	}
}

// GenerateAccessToken generates a new JWT access token
func (j *JWTService) GenerateAccessToken(user *domain.User) (string, error) {
	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		CountryCode: user.CountryCode,
		IsVerified:  user.IsVerified,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken generates a new JWT refresh token
func (j *JWTService) GenerateRefreshToken(user *domain.User) (string, error) {
	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		CountryCode: user.CountryCode,
		IsVerified:  user.IsVerified,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.refreshHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates a JWT token and returns claims
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshAccessToken generates a new access token from a refresh token
func (j *JWTService) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Create a minimal user object for token generation
	user := &domain.User{
		ID:          claims.UserID,
		Email:       claims.Email,
		CountryCode: claims.CountryCode,
		IsVerified:  claims.IsVerified,
	}

	return j.GenerateAccessToken(user)
}

// ExtractUserID extracts user ID from token without full validation
func (j *JWTService) ExtractUserID(tokenString string) (uuid.UUID, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

// IsTokenExpired checks if a token is expired
func (j *JWTService) IsTokenExpired(tokenString string) bool {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return true
	}
	return claims.ExpiresAt.Time.Before(time.Now())
}

// GetTokenRemainingTime returns remaining time until token expires
func (j *JWTService) GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 0 {
		return 0, fmt.Errorf("token expired")
	}

	return remaining, nil
}
