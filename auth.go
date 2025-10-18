package jwtauth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CtxUserIDKey 和 CtxUsernameKey 是用于 Gin Context 的键
const (
	CtxUserIDKey   = "userID"
	CtxUsernameKey = "username"
)

// Claims 定义了 JWT 载荷中包含的自定义数据
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthService 封装了 JWT 认证服务的所有配置和方法
type AuthService struct {
	signingKey []byte
	issuer     string
}

// NewAuthService 创建并初始化一个新的 AuthService 实例
// 参数:
//
//	key: 用于签名的密钥
//	issuer: Token 签发者 (可选，传入空字符串则使用默认值)
func NewAuthService(key []byte, issuer string) (*AuthService, error) {
	if len(key) == 0 {
		return nil, errors.New("signing key cannot be empty")
	}
	if issuer == "" {
		issuer = "GoAuthService"
	}
	return &AuthService{
		signingKey: key,
		issuer:     issuer,
	}, nil
}

// GenerateToken 用于创建一个新的 JWT 字符串
func (s *AuthService) GenerateToken(userID uint, username string, expiration time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiration)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.signingKey)
}

// ParseToken 用于验证和解析 JWT 字符串
func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.signingKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token is expired")
		}
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// JWTMiddleware 是 Gin 框架的认证中间件
// 当认证失败时，中间件会返回 401 错误并中断请求。
func (s *AuthService) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization format. Expected 'Bearer <token>'"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := s.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token validation failed", "detail": err.Error()})
			c.Abort()
			return
		}

		// 认证成功：将 Claims 数据设置到 Gin Context
		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxUsernameKey, claims.Username)

		c.Next()
	}
}
