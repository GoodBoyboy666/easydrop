package database

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
	DriverSQLite   = "sqlite"
)

var (
	ErrNilConfig         = errors.New("数据库配置不能为空")
	ErrEmptyDriver       = errors.New("数据库类型不能为空")
	ErrEmptyHost         = errors.New("数据库主机不能为空")
	ErrInvalidPort       = errors.New("数据库端口必须大于 0")
	ErrEmptyUsername     = errors.New("数据库用户名不能为空")
	ErrEmptyDatabaseName = errors.New("数据库名称不能为空")
	ErrEmptySQLitePath   = errors.New("sqlite 数据库路径不能为空")
	ErrUnsupportedDriver = errors.New("不支持的数据库类型")
)

var ProviderSet = wire.NewSet(NewDB)

type Config struct {
	Driver          string
	Host            string
	Port            int
	Username        string
	Password        string
	DatabaseName    string
	SSLMode         string
	Charset         string
	ParseTime       bool
	Loc             string
	TimeZone        string
	SQLitePath      string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func NewDB(cfg *Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		return nil, ErrEmptyDriver
	}

	dsn, err := buildDSN(driver, cfg)
	if err != nil {
		return nil, err
	}

	dialector, err := buildDialector(driver, dsn)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	return db, nil
}

func buildDialector(driver, dsn string) (gorm.Dialector, error) {
	switch driver {
	case DriverMySQL:
		return mysql.Open(dsn), nil
	case DriverPostgres:
		return postgres.Open(dsn), nil
	case DriverSQLite:
		return sqlite.Open(dsn), nil
	default:
		return nil, ErrUnsupportedDriver
	}
}

func buildDSN(driver string, cfg *Config) (string, error) {
	switch driver {
	case DriverMySQL:
		return buildMySQLDSN(cfg)
	case DriverPostgres:
		return buildPostgresDSN(cfg)
	case DriverSQLite:
		return buildSQLiteDSN(cfg)
	default:
		return "", ErrUnsupportedDriver
	}
}

func buildMySQLDSN(cfg *Config) (string, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return "", ErrEmptyHost
	}
	if cfg.Port <= 0 {
		return "", ErrInvalidPort
	}
	username := strings.TrimSpace(cfg.Username)
	if username == "" {
		return "", ErrEmptyUsername
	}
	databaseName := strings.TrimSpace(cfg.DatabaseName)
	if databaseName == "" {
		return "", ErrEmptyDatabaseName
	}

	charset := strings.TrimSpace(cfg.Charset)
	if charset == "" {
		charset = "utf8mb4"
	}
	loc := strings.TrimSpace(cfg.Loc)
	if loc == "" {
		loc = "Local"
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		username,
		cfg.Password,
		host,
		cfg.Port,
		databaseName,
		url.QueryEscape(charset),
		cfg.ParseTime,
		url.QueryEscape(loc),
	), nil
}

func buildPostgresDSN(cfg *Config) (string, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return "", ErrEmptyHost
	}
	if cfg.Port <= 0 {
		return "", ErrInvalidPort
	}
	username := strings.TrimSpace(cfg.Username)
	if username == "" {
		return "", ErrEmptyUsername
	}
	databaseName := strings.TrimSpace(cfg.DatabaseName)
	if databaseName == "" {
		return "", ErrEmptyDatabaseName
	}

	pgURL := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", host, cfg.Port),
		Path:   databaseName,
	}
	if cfg.Password == "" {
		pgURL.User = url.User(username)
	} else {
		pgURL.User = url.UserPassword(username, cfg.Password)
	}

	query := url.Values{}
	sslMode := strings.TrimSpace(cfg.SSLMode)
	if sslMode == "" {
		sslMode = "disable"
	}
	query.Set("sslmode", sslMode)
	if timeZone := strings.TrimSpace(cfg.TimeZone); timeZone != "" {
		query.Set("TimeZone", timeZone)
	}
	pgURL.RawQuery = query.Encode()

	return pgURL.String(), nil
}

func buildSQLiteDSN(cfg *Config) (string, error) {
	sqlitePath := strings.TrimSpace(cfg.SQLitePath)
	if sqlitePath == "" {
		return "", ErrEmptySQLitePath
	}
	return sqlitePath, nil
}

