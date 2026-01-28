package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func LoadConfigFromEnv() (*Config, error) {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	if host == "" || port == "" || user == "" || password == "" || dbName == "" {
		return nil, fmt.Errorf("missing required database configuration: POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB must all be set")
	}

	return &Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
	}, nil
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

func (c *Config) ConnectionStringWithoutDB() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		c.User, c.Password, c.Host, c.Port)
}

func InitDB(cfg *Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ensureDatabaseExists(ctx, cfg); err != nil {
		return fmt.Errorf("failed to ensure database exists: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(ctx); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	fmt.Println("✅ Database initialized successfully")
	return nil
}

func ensureDatabaseExists(ctx context.Context, cfg *Config) error {
	tempPool, err := pgxpool.New(ctx, cfg.ConnectionStringWithoutDB())
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer tempPool.Close()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
	err = tempPool.QueryRow(ctx, query, cfg.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		fmt.Printf("📦 Creating database '%s'...\n", cfg.DBName)
		_, err = tempPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		fmt.Printf("✅ Database '%s' created successfully\n", cfg.DBName)
	} else {
		fmt.Printf("✅ Database '%s' already exists\n", cfg.DBName)
	}

	return nil
}

func createTables(ctx context.Context) error {
	tasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		type VARCHAR(50) NOT NULL,
		paths_mapping JSONB NOT NULL DEFAULT '[]'::jsonb,
		include_patterns JSONB NOT NULL DEFAULT '[]'::jsonb,
		exclude_patterns JSONB NOT NULL DEFAULT '[]'::jsonb,
		save_mode INTEGER NOT NULL DEFAULT 0,
		open_cache BOOLEAN NOT NULL DEFAULT true,
		mkdir_if_single BOOLEAN NOT NULL DEFAULT false,
		delete_dir BOOLEAN NOT NULL DEFAULT false,
		keep_dir_struct BOOLEAN NOT NULL DEFAULT true,
		schedule_type VARCHAR(50) DEFAULT '',
		schedule_value VARCHAR(255) DEFAULT '',
		reverse BOOLEAN DEFAULT false,
		config VARCHAR(255) DEFAULT '',
		config_id INTEGER,
		is_watching BOOLEAN DEFAULT false,
		watch_error TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	configsTable := `
	CREATE TABLE IF NOT EXISTS configs (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		detail JSONB NOT NULL DEFAULT '{}'::jsonb,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	cacheFilesTable := `
	CREATE TABLE IF NOT EXISTS cache_files (
		id SERIAL PRIMARY KEY,
		task_id INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(task_id, file_path)
	);
	`

	if _, err := pool.Exec(ctx, tasksTable); err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	tasksComments := `
	COMMENT ON TABLE tasks IS '任务表';
	COMMENT ON COLUMN tasks.name IS '任务名称';
	COMMENT ON COLUMN tasks.type IS '任务类型（main/prune 等）';
	COMMENT ON COLUMN tasks.paths_mapping IS '源路径与目标路径映射';
	COMMENT ON COLUMN tasks.include_patterns IS '包含的匹配模式';
	COMMENT ON COLUMN tasks.exclude_patterns IS '排除的匹配模式';
	COMMENT ON COLUMN tasks.save_mode IS '保存模式';
	COMMENT ON COLUMN tasks.open_cache IS '是否开启缓存';
	COMMENT ON COLUMN tasks.mkdir_if_single IS '是否单文件创建目录';
	COMMENT ON COLUMN tasks.delete_dir IS '是否删除目标目录';
	COMMENT ON COLUMN tasks.keep_dir_struct IS '是否保持目录结构';
	COMMENT ON COLUMN tasks.schedule_type IS '定时任务类型';
	COMMENT ON COLUMN tasks.schedule_value IS '定时任务取值';
	COMMENT ON COLUMN tasks.reverse IS '是否反向执行（目标到源）';
	COMMENT ON COLUMN tasks.config IS '绑定的配置名称';
	COMMENT ON COLUMN tasks.config_id IS '绑定的配置ID';
	COMMENT ON COLUMN tasks.is_watching IS '是否监听中';
	COMMENT ON COLUMN tasks.watch_error IS '监听错误信息';
	COMMENT ON COLUMN tasks.created_at IS '创建时间';
	COMMENT ON COLUMN tasks.updated_at IS '更新时间';
	`
	if _, err := pool.Exec(ctx, tasksComments); err != nil {
		return fmt.Errorf("failed to add comments for tasks table: %w", err)
	}

	if _, err := pool.Exec(ctx, configsTable); err != nil {
		return fmt.Errorf("failed to create configs table: %w", err)
	}

	configsComments := `
	COMMENT ON TABLE configs IS '配置表';
	COMMENT ON COLUMN configs.id IS '配置ID';
	COMMENT ON COLUMN configs.name IS '配置名称';
	COMMENT ON COLUMN configs.detail IS '配置详情（JSON）';
	COMMENT ON COLUMN configs.created_at IS '创建时间';
	COMMENT ON COLUMN configs.updated_at IS '更新时间';
	`
	if _, err := pool.Exec(ctx, configsComments); err != nil {
		return fmt.Errorf("failed to add comments for configs table: %w", err)
	}

	if _, err := pool.Exec(ctx, cacheFilesTable); err != nil {
		return fmt.Errorf("failed to create cache_files table: %w", err)
	}

	// Create index for ORDER BY created_at DESC queries (pagination)
	cacheFilesIndexes := `
	CREATE INDEX IF NOT EXISTS idx_cache_files_task_created ON cache_files(task_id, created_at DESC);
	`
	if _, err := pool.Exec(ctx, cacheFilesIndexes); err != nil {
		return fmt.Errorf("failed to create indexes for cache_files table: %w", err)
	}

	cacheFilesComments := `
	COMMENT ON TABLE cache_files IS '缓存文件表';
	COMMENT ON COLUMN cache_files.id IS '主键';
	COMMENT ON COLUMN cache_files.task_id IS '关联任务ID';
	COMMENT ON COLUMN cache_files.file_path IS '缓存文件路径';
	COMMENT ON COLUMN cache_files.created_at IS '创建时间';
	`
	if _, err := pool.Exec(ctx, cacheFilesComments); err != nil {
		return fmt.Errorf("failed to add comments for cache_files table: %w", err)
	}

	// Note: task_logs table has been removed - logs are now stored as files in ./logs/task_{id}/

	fmt.Println("✅ Tables created/verified successfully")
	return nil
}

func GetPool() *pgxpool.Pool {
	return pool
}

func Close() {
	if pool != nil {
		pool.Close()
	}
}
