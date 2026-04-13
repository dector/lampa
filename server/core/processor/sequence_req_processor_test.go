package processor

import (
	"testing"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

func TestSequenceReqProcessor_Process_EmptySequence(t *testing.T) {
	p := NewSequenceReqProcessor(nil)

	response := p.Process(corehttp.HttpRequest{})
	if response.IsPresent() {
		t.Fatal("expected None response")
	}
	if got, want := p.CurrentIndex, 0; got != want {
		t.Fatalf("unexpected current index: got %d, want %d", got, want)
	}
}

func TestSequenceReqProcessor_Process_RotatesProcessors(t *testing.T) {
	p := NewSequenceReqProcessor([]ReqProcessor{
		StaticReqProcessor{StatusCode: 200, Body: []byte("first")},
		StaticReqProcessor{StatusCode: 201, Body: []byte("second")},
	})

	first := p.Process(corehttp.HttpRequest{})
	if !first.IsPresent() {
		t.Fatal("expected first response")
	}
	if got, want := string(first.OrElse(corehttp.HttpResponse{}).Body), "first"; got != want {
		t.Fatalf("unexpected first body: got %q, want %q", got, want)
	}
	if got, want := p.CurrentIndex, 1; got != want {
		t.Fatalf("unexpected current index after first call: got %d, want %d", got, want)
	}

	second := p.Process(corehttp.HttpRequest{})
	if !second.IsPresent() {
		t.Fatal("expected second response")
	}
	if got, want := string(second.OrElse(corehttp.HttpResponse{}).Body), "second"; got != want {
		t.Fatalf("unexpected second body: got %q, want %q", got, want)
	}
	if got, want := p.CurrentIndex, 0; got != want {
		t.Fatalf("unexpected current index after second call: got %d, want %d", got, want)
	}
}

func TestSequenceReqProcessor_Process_AdvancesOnNoneResponse(t *testing.T) {
	p := NewSequenceReqProcessor([]ReqProcessor{
		&sequenceNoneReqProcessor{},
		StaticReqProcessor{StatusCode: 200, Body: []byte("handled")},
	})

	first := p.Process(corehttp.HttpRequest{})
	if first.IsPresent() {
		t.Fatal("expected None from first processor")
	}
	if got, want := p.CurrentIndex, 1; got != want {
		t.Fatalf("unexpected current index after first call: got %d, want %d", got, want)
	}

	second := p.Process(corehttp.HttpRequest{})
	if !second.IsPresent() {
		t.Fatal("expected Some response from second processor")
	}
	if got, want := string(second.OrElse(corehttp.HttpResponse{}).Body), "handled"; got != want {
		t.Fatalf("unexpected second body: got %q, want %q", got, want)
	}
}

func TestSequenceReqProcessor_Process_HandlesInvalidCurrentIndex(t *testing.T) {
	p := &SequenceReqProcessor{
		Processors: []ReqProcessor{
			StaticReqProcessor{StatusCode: 200, Body: []byte("ok")},
		},
		CurrentIndex: 42,
	}

	response := p.Process(corehttp.HttpRequest{})
	if !response.IsPresent() {
		t.Fatal("expected response")
	}
	if got, want := string(response.OrElse(corehttp.HttpResponse{}).Body), "ok"; got != want {
		t.Fatalf("unexpected body: got %q, want %q", got, want)
	}
	if got, want := p.CurrentIndex, 0; got != want {
		t.Fatalf("unexpected current index after call: got %d, want %d", got, want)
	}
}

func TestSequenceReqProcessor_Snapshot(t *testing.T) {
	p := &SequenceReqProcessor{
		Processors: []ReqProcessor{
			StaticReqProcessor{StatusCode: 200, Body: []byte("first")},
			StaticReqProcessor{StatusCode: 201, Body: []byte("second")},
		},
		CurrentIndex: 99,
	}

	state := p.Snapshot()
	if got, want := state.Size, 2; got != want {
		t.Fatalf("unexpected size: got %d, want %d", got, want)
	}
	if got, want := state.NextIndex, 0; got != want {
		t.Fatalf("unexpected next index: got %d, want %d", got, want)
	}
}

func TestSequenceReqProcessor_Reset(t *testing.T) {
	p := NewSequenceReqProcessor([]ReqProcessor{
		StaticReqProcessor{StatusCode: 200, Body: []byte("first")},
		StaticReqProcessor{StatusCode: 201, Body: []byte("second")},
	})

	if err := p.Reset(1); err != nil {
		t.Fatalf("unexpected reset error: %v", err)
	}
	if got, want := p.Snapshot().NextIndex, 1; got != want {
		t.Fatalf("unexpected next index after reset: got %d, want %d", got, want)
	}

	if err := p.Reset(2); err == nil {
		t.Fatal("expected reset error for out-of-bounds index")
	}
}

var _ ReqProcessor = (*sequenceNoneReqProcessor)(nil)

type sequenceNoneReqProcessor struct{}

func (*sequenceNoneReqProcessor) Process(_ corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	return optional.None[corehttp.HttpResponse]()
}
