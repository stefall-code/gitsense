package trend

import (
	"context"
	"log"
	"time"
)

// Worker 定时刷新 Trend Redis 缓存
type Worker struct {
	service  *Service
	interval time.Duration
	stopCh   chan struct{}
}

// NewWorker 创建 Trend Worker
func NewWorker(service *Service, interval time.Duration) *Worker {
	if interval == 0 {
		interval = 6 * time.Hour
	}
	return &Worker{
		service:  service,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动定时刷新
func (w *Worker) Start(ctx context.Context) {
	log.Printf("[trend-worker] started (interval=%s)", w.interval)

	// 启动时立即刷新一次
	if err := w.refresh(ctx); err != nil {
		log.Printf("[trend-worker] initial refresh error: %v", err)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[trend-worker] stopped")
			return
		case <-w.stopCh:
			log.Printf("[trend-worker] stopped")
			return
		case <-ticker.C:
			if err := w.refresh(ctx); err != nil {
				log.Printf("[trend-worker] refresh error: %v", err)
			}
		}
	}
}

// Stop 停止 worker
func (w *Worker) Stop() {
	close(w.stopCh)
}

// refresh 刷新所有窗口的缓存
func (w *Worker) refresh(ctx context.Context) error {
	windows := []TimeWindow{Window7d, Window30d}
	for _, window := range windows {
		if err := w.service.RefreshCache(ctx, window); err != nil {
			log.Printf("[trend-worker] refresh %s error: %v", window, err)
		}
	}
	return nil
}
