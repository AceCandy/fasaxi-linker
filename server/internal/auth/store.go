package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Store 用户数据存储
type Store struct {
	pool *pgxpool.Pool
}

// NewStore 创建新的用户存储
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// GetByUsername 根据用户名查询用户
func (s *Store) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at 
		FROM users 
		WHERE username = $1
	`
	var user User
	err := s.pool.QueryRow(ctx, query, username).Scan(
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

// Create 创建新用户
func (s *Store) Create(ctx context.Context, username, password string) (*User, error) {
	// 使用 bcrypt 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (username, password_hash) 
		VALUES ($1, $2) 
		RETURNING id, username, password_hash, created_at, updated_at
	`
	var user User
	err = s.pool.QueryRow(ctx, query, username, string(hashedPassword)).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

// EnsureDefaultUser 确保默认管理员用户存在
func (s *Store) EnsureDefaultUser(ctx context.Context) error {
	// 检查 admin 用户是否存在
	_, err := s.GetByUsername(ctx, "admin")
	if err == nil {
		// 用户已存在
		fmt.Println("✅ Default admin user already exists")
		return nil
	}

	// 创建默认 admin 用户
	_, err = s.Create(ctx, "admin", "admin123")
	if err != nil {
		return fmt.Errorf("failed to create default admin user: %w", err)
	}

	fmt.Println("✅ Default admin user created (username: admin, password: admin123)")
	return nil
}
