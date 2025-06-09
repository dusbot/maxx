package progressbar

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

type ConcurrentProgressBar struct {
	Total   uint32
	current uint32
	mu      sync.Mutex
}

func (p *ConcurrentProgressBar) Add(n uint32) {
	atomic.AddUint32(&p.current, n)
	p.print()
}

func (p *ConcurrentProgressBar) print() {
	p.mu.Lock()
	defer p.mu.Unlock()

	percent := float64(atomic.LoadUint32(&p.current)) / float64(p.Total) * 100
	fmt.Printf("\r[%-50s] %d/%d (%.1f%%)",
		strings.Repeat("=", int(percent/2)),
		p.current, p.Total, percent)
}