package processor

import (
	"strings"
	"time"

	quickjsexec "github.com/dector/lampa/server/core/exec/quickjs"
	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
	"modernc.org/quickjs"
)

const defaultQuickJSProgram = `function handle(req) {
  return Response.json({ status: "ok" });
}`

// QuickJSReqProcessor processes a request with a QuickJS script.
//
// Script should define `handle(req)` callable that accepts
// one object argument with fields: method, url, headers, body.
//
// Expected return value:
//   - null/undefined: request is not handled
//   - object: {statusCode:number, headers:Record<string,string|string[]>, body:string|Uint8Array}
type QuickJSReqProcessor struct {
	Script      string
	EvalTimeout time.Duration
	MemoryLimit uintptr
}

func (p QuickJSReqProcessor) Process(request corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	script := p.Script
	if strings.TrimSpace(script) == "" {
		script = defaultQuickJSProgram
	}

	vm, err := quickjs.NewVM()
	if err != nil {
		return optional.None[corehttp.HttpResponse]()
	}
	defer vm.Close()

	if p.MemoryLimit > 0 {
		vm.SetMemoryLimit(p.MemoryLimit)
	}
	if p.EvalTimeout > 0 {
		if err := vm.SetEvalTimeout(p.EvalTimeout); err != nil {
			return optional.None[corehttp.HttpResponse]()
		}
	}

	if err := quickjsexec.LoadHandler(vm, script); err != nil {
		return optional.None[corehttp.HttpResponse]()
	}

	requestValue := quickjsexec.ToRequestValue(request)
	result, err := vm.Call(quickjsexec.EntryPoint, requestValue)
	if err != nil {
		return optional.None[corehttp.HttpResponse]()
	}

	if result == nil {
		return optional.None[corehttp.HttpResponse]()
	}
	if _, ok := result.(quickjs.Undefined); ok {
		return optional.None[corehttp.HttpResponse]()
	}

	response, err := quickjsexec.FromResponseValue(result)
	if err != nil {
		return optional.None[corehttp.HttpResponse]()
	}

	return optional.Some(response)
}
