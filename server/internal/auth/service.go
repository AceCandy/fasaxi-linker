package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrInvalidToken       = errors.New("无效的认证令牌")
	ErrTokenExpired       = errors.New("认证令牌已过期")
)

// Service 认证服务
type Service struct {
	store     *Store
	jwtSecret []byte
}

// NewService 创建新的认证服务
func NewService(pool *pgxpool.Pool) *Service {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "linker-default-jwt-secret-change-in-production"
	}

	return &Service{
		store:     NewStore(pool),
		jwtSecret: []byte(secret),
	}
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	// 查询用户
	user, err := s.store.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 生成 JWT
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:    token,
		Username: user.Username,
	}, nil
}

// ValidateToken 验证 JWT 并返回 Claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetCurrentUser 获取当前用户信息
func (s *Service) GetCurrentUser(ctx context.Context, userID int) (*User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`
	var user User
	err := s.store.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// EnsureDefaultUser 确保默认用户存在
func (s *Service) EnsureDefaultUser(ctx context.Context) error {
	return s.store.EnsureDefaultUser(ctx)
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	// 获取用户信息
	user, err := s.GetCurrentUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("旧密码错误")
	}

	// 生成新密码的哈希
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err = s.store.pool.Exec(ctx, query, string(newPasswordHash), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// generateToken 生成 JWT
func (s *Service) generateToken(user *User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "linker",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
