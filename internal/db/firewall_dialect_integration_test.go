package db

// 本文件是真实 MySQL/PostgreSQL 上的防火墙数据适配集成测试，
// 与 postgres_migrate_test.go 相同，需本机数据库并设置
// LEMWOOD_MIGRATION_INTEGRATION=1 才会运行（默认跳过）。
//
// 本地环境约定（见 AGENTS.md）：PostgreSQL 127.0.0.1:55432（信任本机回环），
// MySQL 127.0.0.1:33306（用户 lemwood / testpass，mysql_native_password）。
import (
	"os"
	"path/filepath"
	"testing"

	"lemwood_mirror/internal/config"
)

// 方言集成环境（与既有集成测试一致）。
func dialectIntegrationConfigs() map[string]*config.Config {
	user := getenv("USER")
	return map[string]*config.Config{
		"sqlite": {DatabaseMode: "sqlite"},
		"mysql": {
			DatabaseMode: "mysql",
			MySQLHost:    "127.0.0.1", MySQLPort: 33306, MySQLUser: "lemwood", MySQLPassword: "testpass", MySQLDatabase: "lemwood_fw_test",
		},
		"postgres": {
			DatabaseMode:    "pgsql",
			PostgresHost:    "127.0.0.1", PostgresPort: 55432, PostgresUser: user, PostgresDatabase: "lemwood_fw_test", PostgresSSLMode: "disable",
		},
	}
}

func closeTestDB() {
	if DB != nil {
		_ = DB.Close()
		DB = nil
	}
}

// 防火墙黑名单（含 CIDR 网段条目）在三种方言下的读写适配：
// ip_blacklist.ip 为 TEXT，CIDR 值按字符串精确存取；网段展开匹配在内存层（firewall 包）完成。
func TestBlacklistCIDRAcrossDialectsIntegration(t *testing.T) {
	if testing.Short() || getenv("LEMWOOD_MIGRATION_INTEGRATION") != "1" {
		t.Skip("set LEMWOOD_MIGRATION_INTEGRATION=1 to run local MySQL/PostgreSQL integration")
	}

	for name, cfg := range dialectIntegrationConfigs() {
		t.Run(name, func(t *testing.T) {
			closeTestDB()
			if err := InitDB(t.TempDir(), cfg); err != nil {
				t.Fatalf("InitDB(%s) error = %v", name, err)
			}
			defer closeTestDB()

			// 幂等：清掉历史数据
			if _, err := DB.Exec("DELETE FROM ip_blacklist"); err != nil {
				t.Fatalf("清空 ip_blacklist 失败: %v", err)
			}

			if err := AddIPToBlacklistWithSource("203.0.113.0/24", "网段封禁", "local", "manual"); err != nil {
				t.Fatalf("AddIPToBlacklistWithSource(CIDR) error = %v", err)
			}
			if err := AddIPToBlacklistWithSource("198.51.100.7", "精确 IP", "local", "traffic"); err != nil {
				t.Fatalf("AddIPToBlacklistWithSource(exact) error = %v", err)
			}

			// 重复写入走 upsert 分支，不应报错或产生重复行
			if err := AddIPToBlacklistWithSource("203.0.113.0/24", "网段封禁-更新", "local", "manual"); err != nil {
				t.Fatalf("AddIPToBlacklistWithSource(CIDR upsert) error = %v", err)
			}

			if !IsIPBlacklisted("203.0.113.0/24") {
				t.Fatal("CIDR 条目应按字符串精确命中")
			}
			if IsIPBlacklisted("203.0.113.55") {
				t.Fatal("IP 展开匹配不属于 DB 层职责（由 firewall 内存网段匹配承担）")
			}

			list, err := GetIPBlacklist()
			if err != nil {
				t.Fatalf("GetIPBlacklist() error = %v", err)
			}
			if len(list) != 2 {
				t.Fatalf("GetIPBlacklist() len = %d, want 2: %v", len(list), list)
			}
			for _, item := range list {
				if item["ip"] == "203.0.113.0/24" {
					if item["reason"] != "网段封禁-更新" || item["source"] != "local" || item["ban_type"] != "manual" {
						t.Fatalf("CIDR 条目 upsert 后字段不符: %v", item)
					}
				}
			}

			if err := RemoveIPFromBlacklist("203.0.113.0/24"); err != nil {
				t.Fatalf("RemoveIPFromBlacklist(CIDR) error = %v", err)
			}
			if IsIPBlacklisted("203.0.113.0/24") {
				t.Fatal("删除 CIDR 条目后不应再命中")
			}
		})
	}
}

// SQLite → MySQL 一次性迁移（mysql_migration: true）应携带 CIDR 黑名单条目。
func TestSQLiteToMySQLMigrationCarriesCIDRBlacklistIntegration(t *testing.T) {
	if testing.Short() || getenv("LEMWOOD_MIGRATION_INTEGRATION") != "1" {
		t.Skip("set LEMWOOD_MIGRATION_INTEGRATION=1 to run local MySQL/PostgreSQL integration")
	}

	storage := t.TempDir()
	closeTestDB()
	if err := InitDB(storage, &config.Config{DatabaseMode: "sqlite"}); err != nil {
		t.Fatalf("InitDB(sqlite) error = %v", err)
	}
	if err := AddIPToBlacklistWithSource("203.0.113.0/24", "网段封禁", "local", "traffic"); err != nil {
		t.Fatalf("seed CIDR blacklist error = %v", err)
	}
	closeTestDB()

	cfg := dialectIntegrationConfigs()["mysql"]
	cfg.MySQLMigration = true
	if err := InitDB(storage, cfg); err != nil {
		t.Fatalf("InitDB(mysql with migration) error = %v", err)
	}
	defer closeTestDB()

	if !IsIPBlacklisted("203.0.113.0/24") {
		t.Fatal("CIDR 黑名单条目应随 SQLite→MySQL 迁移保留")
	}
	if _, err := os.Stat(filepath.Join(storage, "stats.db")); !os.IsNotExist(err) {
		t.Fatal("迁移完成后 stats.db 应被重命名为 .bak")
	}
}
