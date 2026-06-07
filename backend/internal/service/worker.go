package service

import (
	"context"
	"log"
	"time"
)

// EmbeddingWorker 后台处理 embedding 生成
// 混合模型：in-memory channel（seed 时立即触发）+ DB fallback scan（定时兜底）
type EmbeddingWorker struct {
	embeddingSvc *EmbeddingService
	interval     time.Duration
	batchSize    int
	jobCh        chan string // 立即执行的 job channel
	stopCh       chan struct{}
}

// NewEmbeddingWorker 创建新的 EmbeddingWorker
func NewEmbeddingWorker(embeddingSvc *EmbeddingService, interval time.Duration, batchSize int) *EmbeddingWorker {
	if interval == 0 {
		interval = 30 * time.Second
	}
	if batchSize <= 0 {
		batchSize = 10
	}
	return &EmbeddingWorker{
		embeddingSvc: embeddingSvc,
		interval:     interval,
		batchSize:    batchSize,
		jobCh:        make(chan string, 1000), // 缓冲 1000 个 job
		stopCh:       make(chan struct{}),
	}
}

// Submit 提交一个立即执行的 embedding job（seed 时调用）
func (w *EmbeddingWorker) Submit(fullName string) {
	select {
	case w.jobCh <- fullName:
		log.Printf("[worker] submitted immediate job: %s", fullName)
	default:
		log.Printf("[worker] job channel full, falling back to DB scan for: %s", fullName)
	}
}

// Start 启动后台 worker
func (w *EmbeddingWorker) Start(ctx context.Context) {
	log.Printf("[worker] embedding worker started (interval=%s, batch=%d)", w.interval, w.batchSize)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// 启动后立即执行一次 DB scan
	w.processOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[worker] embedding worker stopped (context cancelled)")
			return
		case <-w.stopCh:
			log.Printf("[worker] embedding worker stopped")
			return
		case fullName := <-w.jobCh:
			// 立即执行 channel 中的 job
			if err := w.embeddingSvc.GenerateForRepo(ctx, fullName); err != nil {
				log.Printf("[worker] immediate job failed for %s: %v", fullName, err)
			}
		case <-ticker.C:
			// 定时 DB fallback scan
			w.processOnce(ctx)
		}
	}
}

// Stop 停止 worker
func (w *EmbeddingWorker) Stop() {
	close(w.stopCh)
}

func (w *EmbeddingWorker) processOnce(ctx context.Context) {
	if err := w.embeddingSvc.ProcessPending(ctx, w.batchSize); err != nil {
		log.Printf("[worker] process pending error: %v", err)
	}
}
