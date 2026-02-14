// Package middleware 提供HTTP中间件
package middleware

import (
	"errors"
	"strings"
	"time"

	"go_wails_project_manager/logger"
	"go_wails_project_manager/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT 相关错误
var (
	ErrInvalidToken     = errors.New("无效的token")
	ErrExpiredToken     = errors.New("token已过期")
	ErrMissingToken     = errors.New("缺少token")
	ErrInvalidSignature = errors.New("签名验证失败")
)

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey     string        // 密钥
	ExpireTime    time.Duration // 过期时间
	RefreshTime   time.Duration // 刷新时间
	Issuer        string        // 签发者
	TokenLookup   string        // token查找位置 "header:Authorization" 或 "query:token"
	TokenHeadName string        // token前缀，如 "Bearer"
}

// DefaultJWTConfig 默认JWT配置
var DefaultJWTConfig = JWTConfig{
	SecretKey:     "your-secret-key-change-in-production",
	ExpireTime:    24 * time.Hour,
	RefreshTime:   7 * 24 * time.Hour,
	Issuer:        "go_wails_project_manager",
	TokenLookup:   "header:Authorization",
	TokenHeadName: "Bearer",
}

// Claims 自定义JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth JWT认证器
type JWTAuth struct {
	config JWTConfig
}

// NewJWTAuth 创建JWT认证器
func NewJWTAuth(config ...JWTConfig) *JWTAuth {
	cfg := DefaultJWTConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	return &JWTAuth{config: cfg}
}

// GenerateToken 生成JWT token
func (j *JWTAuth) GenerateToken(userID uint, username, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.ExpireTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// GenerateRefreshToken 生成刷新token
func (j *JWTAuth) GenerateRefreshToken(userID uint) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(j.config.RefreshTime)),
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   string(rune(userID)),
		Issuer:    j.config.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// ParseToken 解析JWT token
func (j *JWTAuth) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken 刷新token
func (j *JWTAuth) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil && !errors.Is(err, ErrExpiredToken) {
		return "", err
	}

	// 即使token过期，只要在刷新时间内就可以刷新
	if claims != nil {
		return j.GenerateToken(claims.UserID, claims.Username, claims.Role)
	}

	return "", ErrInvalidToken
}

// AuthMiddleware 认证中间件
func (j *JWTAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := j.extractToken(c)
		if tokenString == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			logger.Log.Warnf("JWT认证失败: %v", err)
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件（不强制要求登录）
func (j *JWTAuth) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := j.extractToken(c)
		if tokenString != "" {
			if claims, err := j.ParseToken(tokenString); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("role", claims.Role)
				c.Set("claims", claims)
			}
		}
		c.Next()
	}
}

// RoleMiddleware 角色验证中间件
func (j *JWTAuth) RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, r := range roles {
			if r == userRole {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}

// ExtractToken 从请求中提取token（公开方法）
func (j *JWTAuth) ExtractToken(c *gin.Context) string {
	return j.extractToken(c)
}

// extractToken 从请求中提取token
func (j *JWTAuth) extractToken(c *gin.Context) string {
	// 从Header提取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == j.config.TokenHeadName {
			return parts[1]
		}
	}

	// 从Query参数提取
	if token := c.Query("token"); token != "" {
		return token
	}

	// 从Cookie提取
	if token, err := c.Cookie("token"); err == nil {
		return token
	}

	return ""
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) uint {
	if id, exists := c.Get("user_id"); exists {
		return id.(uint)
	}
	return 0
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		return username.(string)
	}
	return ""
}

// GetUserRole 从上下文获取用户角色
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return ""
}

// ==================== 密码工具 ====================

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ==================== 响应结构 ====================

// TokenResponse token响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// GenerateTokenResponse 生成token响应
func (j *JWTAuth) GenerateTokenResponse(userID uint, username, role string) (*TokenResponse, error) {
	accessToken, err := j.GenerateToken(userID, username, role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := j.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.config.ExpireTime.Seconds()),
		TokenType:    j.config.TokenHeadName,
	}, nil
}

// ==================== 登录处理示例 ====================

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginHandler 登录处理器示例（需要根据实际用户模型调整）
func (j *JWTAuth) LoginHandler(validateUser func(username, password string) (uint, string, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "请求参数错误")
			return
		}

		userID, role, err := validateUser(req.Username, req.Password)
		if err != nil {
			response.Unauthorized(c, "用户名或密码错误")
			return
		}

		tokenResp, err := j.GenerateTokenResponse(userID, req.Username, role)
		if err != nil {
			response.InternalServerError(c, "生成token失败")
			return
		}

		response.Success(c, tokenResp)
	}
}

// RefreshHandler 刷新token处理器
func (j *JWTAuth) RefreshHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := j.extractToken(c)
		if tokenString == "" {
			response.Unauthorized(c, "缺少token")
			return
		}

		newToken, err := j.RefreshToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "刷新token失败")
			return
		}

		response.Success(c, gin.H{
			"access_token": newToken,
			"expires_in":   int64(j.config.ExpireTime.Seconds()),
			"token_type":   j.config.TokenHeadName,
		})
	}
}

// LogoutHandler 登出处理器
func (j *JWTAuth) LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 清除cookie
		c.SetCookie("token", "", -1, "/", "", false, true)
		response.SuccessWithMsg(c, "登出成功", nil)
	}
}
