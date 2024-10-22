package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

// 定义 secret key 用于签发和验证 JWT
var jwtSecret = []byte("AFaGfgddjtyrjty46$xds")

// User 模拟一个用户结构，包含用户 ID 和权限
type User struct {
	ID       int
	Username string
	Password string
	Role     string // 角色：user 或 admin
}

// 模拟的用户数据库
var users = []User{
	{ID: 1, Username: "admin", Password: "password123", Role: "admin"},
	{ID: 2, Username: "user", Password: "password456", Role: "user"},
}

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWTToken
func GenerateJWT(userID int, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // JWT 24小时后过期
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用 secret key 签名并生成 JWT 字符串
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// JWT 验证函数
func validateJWT(tokenString string) (*Claims, error) {
	// 解析并验证 token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证 token 的签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	// 检查 token 是否有效
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		// 允许使用 "Bearer " 前缀的 JWT
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims, err := validateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 将用户 ID 和角色存储到上下文中，便于后续使用
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
