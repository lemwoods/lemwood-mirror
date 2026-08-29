package bandwidth

import (
	"testing"
	"time"
)

func TestNewTrackerDefaultsPeakBandwidth(t *testing.T) {
	status := NewTracker(0).Snapshot()
	if status.PeakBandwidthMbps != 200 {
		t.Fatalf("peak bandwidth = %d, want 200", status.PeakBandwidthMbps)
	}
}

func TestTrackerRecordsBytesAndDownloads(t *testing.T) {
	tracker := NewTracker(200)
	tracker.StartDownload()
	tracker.RecordBytes(25_000_000)
	status := tracker.Snapshot()
	if status.TotalBytesServed != 25_000_000 {
		t.Fatalf("total bytes = %d, want 25000000", status.TotalBytesServed)
	}
	if status.ActiveDownloads != 1 {
		t.Fatalf("active downloads = %d, want 1", status.ActiveDownloads)
	}
	if status.CurrentBandwidthMbps <= 0 {
		t.Fatalf("current bandwidth should be positive: %+v", status)
	}
	tracker.FinishDownload()
	if got := tracker.Snapshot().ActiveDownloads; got != 0 {
		t.Fatalf("active downloads after finish = %d, want 0", got)
	}
}

func TestRecentDownloadsWindowKeepsFinishedConnections(t *testing.T) {
	tracker := NewTracker(200)

	// 两个连接开始，其中一个已结束：瞬时归零，recent 窗口仍应保留 60s
	tracker.StartDownload()
	tracker.StartDownload()
	tracker.FinishDownload()

	status := tracker.Snapshot()
	if status.RecentWindowSeconds != 60 {
		t.Fatalf("recent window = %d, want 60", status.RecentWindowSeconds)
	}
	if status.RecentDownloads != 2 {
		t.Fatalf("recent downloads = %d, want 2 (窗口内已结束的连接也应计入)", status.RecentDownloads)
	}
	if status.ActiveDownloads != 1 {
		t.Fatalf("active downloads = %d, want 1", status.ActiveDownloads)
	}
}

func TestRecentDownloadsPrunesExpiredConnections(t *testing.T) {
	tracker := NewTracker(200)

	// 手工注入一条 61s 前的连接记录，应被窗口清理
	tracker.connMu.Lock()
	tracker.connTimes = append(tracker.connTimes, time.Now().Add(-61*time.Second))
	tracker.connMu.Unlock()

	tracker.StartDownload()

	if got := tracker.Snapshot().RecentDownloads; got != 1 {
		t.Fatalf("recent downloads = %d, want 1 (过期连接应被清理)", got)
	}
}
