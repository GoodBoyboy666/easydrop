package database

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func newTestSQLiteDSN(t *testing.T) string {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	return fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
}

func TestNewDB_ConfigValidation(t *testing.T) {
	t.Parallel()

	_, err := NewDB(nil)
	if !errors.Is(err, ErrNilConfig) {
		t.Fatalf("期望错误 ErrNilConfig，实际为: %v", err)
	}

	_, err = NewDB(&Config{SQLitePath: newTestSQLiteDSN(t)})
	if !errors.Is(err, ErrEmptyDriver) {
		t.Fatalf("期望错误 ErrEmptyDriver，实际为: %v", err)
	}

	_, err = NewDB(&Config{Driver: DriverSQLite})
	if !errors.Is(err, ErrEmptySQLitePath) {
		t.Fatalf("期望错误 ErrEmptySQLitePath，实际为: %v", err)
	}

	_, err = NewDB(&Config{Driver: "oracle"})
	if !errors.Is(err, ErrUnsupportedDriver) {
		t.Fatalf("期望错误 ErrUnsupportedDriver，实际为: %v", err)
	}

	_, err = NewDB(&Config{Driver: DriverMySQL, Port: 3306, Username: "root", DatabaseName: "easydrop"})
	if !errors.Is(err, ErrEmptyHost) {
		t.Fatalf("期望错误 ErrEmptyHost，实际为: %v", err)
	}

	_, err = NewDB(&Config{Driver: DriverMySQL, Host: "127.0.0.1", Port: 0, Username: "root", DatabaseName: "easydrop"})
	if !errors.Is(err, ErrInvalidPort) {
		t.Fatalf("期望错误 ErrInvalidPort，实际为: %v", err)
	}

	_, err = NewDB(&Config{Driver: DriverMySQL, Host: "127.0.0.1", Port: 3306, DatabaseName: "easydrop"})
	if !errors.Is(err, ErrEmptyUsername) {
		t.Fatalf("期望错误 ErrEmptyUsername，实际为: %v", err)
	}

	_, err = NewDB(&Config{Driver: DriverMySQL, Host: "127.0.0.1", Port: 3306, Username: "root"})
	if !errors.Is(err, ErrEmptyDatabaseName) {
		t.Fatalf("期望错误 ErrEmptyDatabaseName，实际为: %v", err)
	}
}

func TestNewDB_SQLiteSuccess(t *testing.T) {
	t.Parallel()

	db, err := NewDB(&Config{
		Driver:          DriverSQLite,
		SQLitePath:      newTestSQLiteDSN(t),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("创建 sqlite 数据库连接失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取 sql.DB 实例失败: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping 数据库失败: %v", err)
	}

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 10 {
		t.Fatalf("最大连接数不符合预期，want=10，got=%d", stats.MaxOpenConnections)
	}
}

func TestBuildDialector_CaseInsensitiveDriver(t *testing.T) {
	t.Parallel()

	db, err := NewDB(&Config{Driver: "SQLITE", SQLitePath: newTestSQLiteDSN(t)})
	if err != nil {
		t.Fatalf("大小写驱动名应可识别，实际错误: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取 sql.DB 实例失败: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
}
