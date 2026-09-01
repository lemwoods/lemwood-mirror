package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"lemwood_mirror/internal/config"
)

func TestPostgresBuiltInMigrationIntegration(t *testing.T) {
	if testing.Short() || getenv("LEMWOOD_MIGRATION_INTEGRATION") != "1" {
		t.Skip("set LEMWOOD_MIGRATION_INTEGRATION=1 to run local MySQL/PostgreSQL integration")
	}
	if DB != nil {
		DB.Close()
		DB = nil
	}
	resetPostgresTestDB(t, "127.0.0.1", 55432, getenv("USER"), "lemwood_builtin_test")
	cfg := &config.Config{
		DatabaseMode: "pgsql",
		MySQLHost:    "127.0.0.1", MySQLPort: 33306, MySQLUser: "lemwood", MySQLPassword: "testpass", MySQLDatabase: "lemwood_source",
		PostgresHost: "127.0.0.1", PostgresPort: 55432, PostgresUser: getenv("USER"), PostgresDatabase: "lemwood_builtin_test", PostgresSSLMode: "disable",
		PostgresMigrationBatch: 2, PostgresMigrationDelay: "0s",
	}
	if err := InitDB(t.TempDir(), cfg); err != nil {
		t.Fatalf("InitDB first run error = %v", err)
	}
	var marker string
	if err := DB.QueryRow(`SELECT value FROM system_info WHERE "key"=$1`, postgresCleanMigrationKey).Scan(&marker); err != nil {
		t.Fatalf("migration marker error = %v", err)
	}
	var visitsBefore int64
	if err := DB.QueryRow("SELECT COALESCE(SUM(visit_count),0) FROM visits").Scan(&visitsBefore); err != nil {
		t.Fatalf("visit count error = %v", err)
	}
	DB.Close()
	DB = nil
	if err := InitDB(t.TempDir(), cfg); err != nil {
		t.Fatalf("InitDB second run error = %v", err)
	}
	defer func() { DB.Close(); DB = nil }()
	var visitsAfter int64
	if err := DB.QueryRow("SELECT COALESCE(SUM(visit_count),0) FROM visits").Scan(&visitsAfter); err != nil {
		t.Fatalf("visit count after restart error = %v", err)
	}
	if visitsAfter != visitsBefore {
		t.Fatalf("migration repeated: visits before=%d after=%d", visitsBefore, visitsAfter)
	}
}

func TestPostgresBuiltInMigrationFallsBackToSQLiteIntegration(t *testing.T) {
	if testing.Short() || getenv("LEMWOOD_MIGRATION_INTEGRATION") != "1" {
		t.Skip("set LEMWOOD_MIGRATION_INTEGRATION=1 to run local PostgreSQL integration")
	}
	storage := t.TempDir()
	source, err := sql.Open("sqlite", filepath.Join(storage, "stats.db"))
	if err != nil {
		t.Fatalf("open sqlite source: %v", err)
	}
	queries := []string{
		`CREATE TABLE visits (id INTEGER PRIMARY KEY, country TEXT, region TEXT, city TEXT, created_at DATETIME)`,
		`CREATE TABLE download_events (id INTEGER PRIMARY KEY, file_path TEXT, file_name TEXT, launcher TEXT, version TEXT, client_ip TEXT, country TEXT, bytes_served INTEGER, completed INTEGER, status_code INTEGER, date TEXT)`,
		`CREATE TABLE ip_daily_traffic (ip TEXT, date TEXT, bytes_downloaded INTEGER, PRIMARY KEY(ip,date))`,
		`CREATE TABLE daily_traffic (date TEXT PRIMARY KEY, bytes_downloaded INTEGER)`,
		`CREATE TABLE daily_completed_traffic (date TEXT PRIMARY KEY, bytes_downloaded INTEGER)`,
		`CREATE TABLE ip_blacklist (ip TEXT PRIMARY KEY, reason TEXT, source TEXT, ban_type TEXT, created_at DATETIME)`,
		`CREATE TABLE system_info (key TEXT PRIMARY KEY, value TEXT)`,
		`INSERT INTO visits VALUES (1,'CN','A','B','2026-08-20 01:00:00'),(2,'CN','A','B','2026-08-20 02:00:00')`,
		`INSERT INTO download_events VALUES (1,'a','a','fcl','1','1.2.3.4','CN',100,1,200,'2026-08-20'),(2,'a','a','fcl','1','1.2.3.4','CN',200,1,200,'2026-08-20')`,
		`INSERT INTO ip_blacklist VALUES ('203.0.113.0/24','网段封禁','local','traffic','2026-08-20 00:00:00'),('198.51.100.7','手动封禁','manual','manual','2026-08-20 00:00:00')`,
		`INSERT INTO system_info VALUES ('start_time','2026-08-20 00:00:00')`,
	}
	for _, query := range queries {
		if _, err := source.Exec(query); err != nil {
			source.Close()
			t.Fatalf("prepare sqlite source: %v", err)
		}
	}
	source.Close()

	if DB != nil {
		DB.Close()
		DB = nil
	}
	resetPostgresTestDB(t, "127.0.0.1", 55432, getenv("USER"), "lemwood_builtin_sqlite_test")
	cfg := &config.Config{
		DatabaseMode: "pgsql",
		MySQLHost:    "127.0.0.1", MySQLPort: 1, MySQLUser: "invalid", MySQLDatabase: "invalid",
		PostgresHost: "127.0.0.1", PostgresPort: 55432, PostgresUser: getenv("USER"), PostgresDatabase: "lemwood_builtin_sqlite_test", PostgresSSLMode: "disable",
		PostgresMigrationBatch: 2, PostgresMigrationDelay: "0s",
	}
	if err := InitDB(storage, cfg); err != nil {
		t.Fatalf("InitDB fallback error = %v", err)
	}
	defer func() { DB.Close(); DB = nil }()
	var visits, events, bytes int64
	if err := DB.QueryRow("SELECT COALESCE(SUM(visit_count),0) FROM visits").Scan(&visits); err != nil {
		t.Fatal(err)
	}
	if err := DB.QueryRow("SELECT COALESCE(SUM(event_count),0),COALESCE(SUM(bytes_served),0) FROM download_events").Scan(&events, &bytes); err != nil {
		t.Fatal(err)
	}
	if visits != 2 || events != 2 || bytes != 300 {
		t.Fatalf("fallback aggregates visits/events/bytes=%d/%d/%d", visits, events, bytes)
	}

	// 防火墙黑名单（含 CIDR 网段条目）应原样迁移到 PostgreSQL
	var banned int
	if err := DB.QueryRow("SELECT COUNT(*) FROM ip_blacklist WHERE ip IN ('203.0.113.0/24','198.51.100.7')").Scan(&banned); err != nil {
		t.Fatal(err)
	}
	if banned != 2 {
		t.Fatalf("ip_blacklist migrated rows = %d, want 2 (CIDR + exact)", banned)
	}
	var banType string
	if err := DB.QueryRow("SELECT ban_type FROM ip_blacklist WHERE ip = '203.0.113.0/24'").Scan(&banType); err != nil {
		t.Fatal(err)
	}
	if banType != "traffic" {
		t.Fatalf("CIDR entry ban_type = %q, want traffic", banType)
	}
}

func getenv(name string) string {
	return os.Getenv(name)
}

// resetPostgresTestDB 删除并重建目标库，使集成测试可重复运行：
// 清洗迁移有一次性完成标记（system_info.postgres_clean_migration_v1），
// 复用旧库会让后续运行跳过迁移，断言到的是上一轮的残留数据。
func resetPostgresTestDB(t *testing.T, host string, port int, user, dbname string) {
	t.Helper()
	admin, err := sql.Open("pgx", fmt.Sprintf("postgres://%s@%s:%d/postgres?sslmode=disable", user, host, port))
	if err != nil {
		t.Fatalf("open postgres admin db: %v", err)
	}
	defer admin.Close()
	if _, err := admin.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", dbname)); err != nil {
		t.Fatalf("drop test database %s: %v", dbname, err)
	}
	if _, err := admin.Exec(fmt.Sprintf("CREATE DATABASE %s OWNER %s", dbname, user)); err != nil {
		t.Fatalf("create test database %s: %v", dbname, err)
	}
}

func TestOpenPreferredMigrationSourceFallsBackToSQLite(t *testing.T) {
	storage := t.TempDir()
	sqlitePath := filepath.Join(storage, "stats.db")
	d, err := sql.Open("sqlite", sqlitePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if _, err := d.Exec("CREATE TABLE visits (id INTEGER PRIMARY KEY)"); err != nil {
		d.Close()
		t.Fatalf("create sqlite source: %v", err)
	}
	d.Close()

	source, err := openPreferredMigrationSource(storage, &config.Config{
		MySQLHost: "127.0.0.1", MySQLPort: 1,
		MySQLUser: "invalid", MySQLDatabase: "invalid",
	})
	if err != nil {
		t.Fatalf("openPreferredMigrationSource error = %v", err)
	}
	if source == nil {
		t.Fatal("expected SQLite fallback source")
	}
	defer source.db.Close()
	if source.dialect != "sqlite" {
		t.Fatalf("source dialect = %q, want sqlite", source.dialect)
	}
}

func TestOpenPreferredMigrationSourceReturnsNilWithoutSources(t *testing.T) {
	source, err := openPreferredMigrationSource(t.TempDir(), &config.Config{})
	if err != nil {
		t.Fatalf("openPreferredMigrationSource error = %v", err)
	}
	if source != nil {
		source.db.Close()
		t.Fatal("expected no migration source")
	}
}
