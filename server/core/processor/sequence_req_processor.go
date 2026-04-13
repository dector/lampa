package processor

import (
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

// NewSequenceReqProcessor creates sequence processor with copied processors list.
func NewSequenceReqProcessor(processors []ReqProcessor) *SequenceReqProcessor {
	cloned := make([]ReqProcessor, len(processors))
	copy(cloned, processors)
	return &SequenceReqProcessor{Processors: cloned}
}

func (p *SequenceReqProcessor) Process(request corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.Processors) == 0 {
		return optional.None[corehttp.HttpResponse]()
	}

	if p.CurrentIndex < 0 || p.CurrentIndex >= len(p.Processors) {
		p.CurrentIndex = 0
	}

	activeIndex := p.CurrentIndex
	active := p.Processors[activeIndex]
	p.CurrentIndex = (activeIndex + 1) % len(p.Processors)

	if active == nil {
		return optional.None[corehttp.HttpResponse]()
	}

	return active.Process(request)
}
