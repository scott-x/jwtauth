package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scott-x/jwtauth"
)

// 初始化全局 JWT 服务实例
var (
	authService *jwtauth.AuthService
	expires     time.Duration = time.Hour * 12
)

func init() {
	secretKey := []byte("a_secure_32_byte_string_for_signing")

	var err error
	authService, err = jwtauth.NewAuthService(secretKey, "MyAppIssuer")
	if err != nil {
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}
}

func main() {
	r := SetupRouter()

	log.Println("Server listening on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func SetupRouter() *gin.Engine {
	router := gin.Default()
	
	//public group
	public := router.Group("/api/v1/public")
	{
		public.POST("/login", LoginHandler) //login
		public.GET("/ping", PingTest)
	}

	//protected group
	protected := router.Group("/api/v1/protected")
	protected.Use(authService.JWTMiddleware())
	{
		protected.GET("/user/info", ProtectedResourceHandler)
	}

	return router
}

func PingTest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// ProtectedResourceHandler 访问受保护的资源
func ProtectedResourceHandler(c *gin.Context) {
	// 从 Gin Context 中获取用户 ID 和 Username
	// 中间件已将它们注入
	userID := c.MustGet(jwtauth.CtxUserIDKey).(uint)
	username := c.MustGet(jwtauth.CtxUsernameKey).(string)

	c.JSON(http.StatusOK, gin.H{
		"message": "Access granted",
		"user":    gin.H{"id": userID, "name": username},
	})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler 模拟登录并签发 Token
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 实际应用中进行数据库验证...
	if req.Username == "admin" && req.Password == "123456" {
		// 签发 Token (例如，用户 ID 101，有效期 12 小时)
		token, err := authService.GenerateToken(101, req.Username, expires)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token, "expires_in": "12h"})

	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
}
