package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/cache"
	"xingyunpan-v2/pkg/jwt"
	"xingyunpan-v2/pkg/response"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			if tokenString := strings.TrimSpace(c.Query("access_token")); tokenString != "" {
				authHeader = "Bearer " + tokenString
			} else if tokenString := strings.TrimSpace(c.Query("token")); tokenString != "" {
				authHeader = "Bearer " + tokenString
			}
		}
		if authHeader == "" {
			response.Unauthorized(c, "missing Authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid Authorization header")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			response.Unauthorized(c, "token cannot be empty")
			c.Abort()
			return
		}

		userID, username, oauthScopes, err := resolveBearerToken(c.Request.Context(), tokenString)
		if err != nil {
			response.Unauthorized(c, "token is invalid or expired")
			c.Abort()
			return
		}

		userStatus, err := getUserAuthStatus(c.Request.Context(), userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Unauthorized(c, "user not found")
			} else {
				response.Unauthorized(c, "failed to verify user status")
			}
			c.Abort()
			return
		}
		if !userStatus.Enabled {
			response.Forbidden(c, "account is disabled")
			c.Abort()
			return
		}

		setAuthenticatedUser(c, userStatus)
		if len(oauthScopes) > 0 {
			c.Set("oauth_scopes", oauthScopes)
			c.Set("oauth_username", username)
		}
		touchUserLastSeen(userStatus.ID)

		c.Next()
	}
}

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := strings.TrimSpace(parts[1])
			if tokenString != "" {
				claims, err := jwt.ParseToken(tokenString, config.Config.JWT.Secret)
				if err == nil {
					userStatus, statusErr := getUserAuthStatus(c.Request.Context(), claims.UserID)
					if statusErr == nil && userStatus.Enabled {
						setAuthenticatedUser(c, userStatus)
						touchUserLastSeen(userStatus.ID)
					}
				}
			}
		}

		c.Next()
	}
}

func RequireOAuthScope(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get("oauth_scopes")
		if !exists {
			c.Next()
			return
		}
		granted, _ := raw.([]string)
		grantedSet := make(map[string]bool, len(granted))
		for _, scope := range granted {
			grantedSet[strings.TrimSpace(scope)] = true
		}
		for _, scope := range scopes {
			scope = strings.TrimSpace(scope)
			if scope == "" || grantedSet[scope] {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "insufficient OAuth scope")
		c.Abort()
	}
}

func resolveBearerToken(ctx context.Context, tokenString string) (uint, string, []string, error) {
	claims, err := jwt.ParseToken(tokenString, config.Config.JWT.Secret)
	if err == nil {
		return claims.UserID, claims.Username, nil, nil
	}

	db := config.GetDB()
	if db == nil {
		return 0, "", nil, gorm.ErrInvalidDB
	}
	now := time.Now()
	var access model.OAuthAccessToken
	if err := db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hashBearerToken(tokenString), now).
		First(&access).Error; err != nil {
		return 0, "", nil, err
	}
	var app model.OAuthApp
	if err := db.WithContext(ctx).First(&app, access.AppID).Error; err != nil {
		return 0, "", nil, err
	}
	if !app.Enabled {
		return 0, "", nil, gorm.ErrRecordNotFound
	}
	var scopes []string
	_ = json.Unmarshal([]byte(access.ScopesJSON), &scopes)
	return access.UserID, "", scopes, nil
}

func hashBearerToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func setAuthenticatedUser(c *gin.Context, userStatus *cache.UserAuthStatusCacheEntry) {
	c.Set("user_id", userStatus.ID)
	c.Set("username", userStatus.Username)
	c.Set("user_role", userStatus.Role)
}

func getUserAuthStatus(ctx context.Context, userID uint) (*cache.UserAuthStatusCacheEntry, error) {
	if userID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	cacheService := cache.NewCacheService(config.GetRedis())
	var cached cache.UserAuthStatusCacheEntry
	if err := cacheService.GetUserAuthStatus(ctx, userID, &cached); err == nil && cached.ID != 0 {
		return &cached, nil
	}

	db := config.GetDB()
	if db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var user model.User
	if err := db.Select("id", "username", "role", "enabled").First(&user, userID).Error; err != nil {
		return nil, err
	}

	status := cache.UserAuthStatusCacheEntry{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		Enabled:  user.Enabled,
	}
	_ = cacheService.CacheUserAuthStatus(ctx, userID, status)
	return &status, nil
}

func touchUserLastSeen(userID uint) {
	if userID == 0 {
		return
	}
	db := config.GetDB()
	if db == nil {
		return
	}

	now := time.Now()
	go func(database *gorm.DB) {
		_ = database.Model(&model.User{}).
			Where("id = ?", userID).
			Update("last_seen_at", now).Error
	}(db)
}
