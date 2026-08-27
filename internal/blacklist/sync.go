package blacklist

import (
	"bufio"
	"fmt"
	"io"
	"lemwood_mirror/internal/db"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	externalIPs   map[string]bool
	externalIPsMu sync.RWMutex
	lastSyncTime  time.Time
	syncMu        sync.Mutex
)

func init() {
	externalIPs = make(map[string]bool)
}

func SyncExternalBlacklist(url string) error {
	if url == "" {
		return nil
	}

	syncMu.Lock()
	defer syncMu.Unlock()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取外部黑名单失败，状态码: %d", resp.StatusCode)
	}

	var ips []string
	var newExternalIPs = make(map[string]bool)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 剥离行内注释（# 之后的内容）
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}
		// 仅当形如 ipv4:port 或 ip:注释（单个冒号）时切掉冒号后内容；
		// IPv6 地址含多个冒号，原样保留，避免被截断成垃圾条目。
		if strings.Count(line, ":") == 1 {
			line = strings.TrimSpace(strings.SplitN(line, ":", 2)[0])
		}
		if line != "" {
			ips = append(ips, line)
			newExternalIPs[line] = true
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}

	externalIPsMu.Lock()
	externalIPs = newExternalIPs
	lastSyncTime = time.Now()
	externalIPsMu.Unlock()

	if err := db.AddExternalBlacklist(ips); err != nil {
		return err
	}

	log.Printf("[黑名单同步] 成功同步 %d 个外部黑名单IP", len(ips))
	return nil
}

func IsExternalBlacklisted(ip string) bool {
	externalIPsMu.RLock()
	defer externalIPsMu.RUnlock()
	return externalIPs[ip]
}

func GetExternalBlacklistCount() int {
	externalIPsMu.RLock()
	defer externalIPsMu.RUnlock()
	return len(externalIPs)
}

func GetLastSyncTime() time.Time {
	externalIPsMu.RLock()
	defer externalIPsMu.RUnlock()
	return lastSyncTime
}
