# jwtauth

`jwtauth` 是一个基于 `github.com/golang-jwt/jwt/v5` 的轻量级、可配置的 Go JWT 认证服务库，专门为 [Gin Web 框架](https://github.com/gin-gonic/gin) 设计了开箱即用的中间件。

## ✨ 特性

* **配置解耦**: 密钥和签发者在服务初始化时传入，而非硬编码。
* **Gin 中间件**: 提供一个 `JWTMiddleware`，轻松保护路由。
* **类型安全**: 使用 `Claims` 结构体处理用户数据。
* **错误处理**: 明确区分 Token 过期、无效签名等错误。

## 📦 安装

```bash
go get github.com/scott-x/jwtauth
```

## 🚀 使用示例 (Gin 框架)

### 步骤 1: 初始化 JWT 服务

在您的主程序中，初始化 `AuthService` 实例。

```go
package main

import (
    "log"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/scott-x/jwtauth" 
)

// 初始化全局 JWT 服务实例
var authService *jwtauth.AuthService

func init() {
    secretKey := []byte("a_secure_32_byte_string_for_signing")
    
    var err error
    authService, err = jwtauth.NewAuthService(secretKey, "MyAppIssuer")
    if err != nil {
        log.Fatalf("Failed to initialize JWT service: %v", err)
    }
}
```

### 步骤 2: 登录路由 (生成 Token)

使用 `authService.GenerateToken` 来签发 Token。

```go
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
    if req.Username != "admin" || req.Password != "password" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // 签发 Token (例如，用户 ID 101，有效期 12 小时)
    token, err := authService.GenerateToken(101, req.Username, 12 * time.Hour)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": token, "expires_in": "12h"})
}
```

### 步骤 3: 受保护路由 (应用中间件)

使用 `authService.JWTMiddleware()` 来保护您的路由。

```go
// ProtectedResourceHandler 访问受保护的资源
func ProtectedResourceHandler(c *gin.Context) {
    // 从 Gin Context 中获取用户 ID 和 Username
    // 中间件已将它们注入
    userID := c.MustGet(jwtauth.CtxUserIDKey).(uint)
    username := c.MustGet(jwtauth.CtxUsernameKey).(string)

    c.JSON(http.StatusOK, gin.H{
        "message": "Access granted",
        "user": gin.H{"id": userID, "name": username},
    })
}

func SetupRouter() *gin.Engine {
    router := gin.Default()

    router.POST("/login", LoginHandler)

    // 应用中间件到受保护的组
    protected := router.Group("/api/v1")
    protected.Use(authService.JWTMiddleware()) 
    {
        protected.GET("/user/info", ProtectedResourceHandler)
    }
    
    return router
}

// func main() { ... router.Run(":8080") }
```

## 🛠️ 方法参考

| 方法 | 签名 | 描述 |
| :--- | :--- | :--- |
| `NewAuthService` | `func(key []byte, issuer string) (*AuthService, error)` | 构造函数，初始化服务并设置密钥和签发者。 |
| `GenerateToken` | `func(userID uint, username string, expiration time.Duration) (string, error)` | 签发新的 JWT 字符串。 |
| `ParseToken` | `func(tokenString string) (*Claims, error)` | 解析和验证 Token，并返回 Claims。 |
| `JWTMiddleware` | `func() gin.HandlerFunc` | Gin 中间件，用于验证请求头中的 Bearer Token。 |

## 许可证

本项目采用 MIT 许可证。

