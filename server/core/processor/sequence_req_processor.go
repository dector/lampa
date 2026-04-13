package processor

import (
	"fmt"
	"sync"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

// SequenceReqProcessor processes requests with sub-processors one by one.
//
// CurrentIndex points to the currently active sub-processor index.
// After each Process call it moves to the next index (cyclic).
type SequenceReqProcessor struct {
	mu sync.Mutex

	Processors   []ReqProcessor
	CurrentIndex int
}

// SequenceState represents current runtime state of sequence processor.
type SequenceState struct {
	Size      int
	NextIndex int
}

// NewSequenceReqProcessor creates sequence processor with copied processors list.
func NewSequenceReqProcessor(processors []ReqProcessor) *SequenceReqProcessor {
	cloned := make([]ReqProcessor, len(processors))
	copy(cloned, processors)
	return &SequenceReqProcessor{Processors: cloned}
}

// Snapshot returns current sequence runtime state.
func (p *SequenceReqProcessor) Snapshot() SequenceState {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.snapshotLocked()
}

// Reset sets sequence next active index.
//
// For empty sequence only index 0 is valid.
func (p *SequenceReqProcessor) Reset(index int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if index < 0 {
		return fmt.Errorf("invalid index")
	}

	size := len(p.Processors)
	if size == 0 {
		if index != 0 {
			return fmt.Errorf("invalid index")
		}
		p.CurrentIndex = 0
		return nil
	}

	if index >= size {
		return fmt.Errorf("invalid index")
	}

	p.CurrentIndex = index
	return nil
}

func (p *SequenceReqProcessor) Process(request corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	p.mu.Lock()
	defer p.mu.Unlock()

	state := p.snapshotLocked()
	if state.Size == 0 {
		return optional.None[corehttp.HttpResponse]()
	}

	activeIndex := state.NextIndex
	active := p.Processors[activeIndex]
	p.CurrentIndex = (activeIndex + 1) % state.Size

	if active == nil {
		return optional.None[corehttp.HttpResponse]()
	}

	return active.Process(request)
}

func (p *SequenceReqProcessor) snapshotLocked() SequenceState {
	size := len(p.Processors)
	if size == 0 {
		p.CurrentIndex = 0
		return SequenceState{Size: 0, NextIndex: 0}
	}

	if p.CurrentIndex < 0 || p.CurrentIndex >= size {
		p.CurrentIndex = 0
	}

	return SequenceState{Size: size, NextIndex: p.CurrentIndex}
}
