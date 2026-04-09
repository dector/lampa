package processor

import (
	"strings"

	starlexec "github.com/dector/lampa/server/core/exec/starlark"
	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
	"go.starlark.net/starlark"
)

const defaultStarlarkProgram = `def handle(req):
    return {
        "statusCode": 200,
        "headers": {
            "Content-Type": "application/json",
        },
        "body": "{\"status\":\"ok\"}",
    }
`

// StarlarkReqProcessor processes a request with a Starlark script.
//
// Script should define a callable (by default, `handle`) that accepts
// one dict argument with keys: method, url, headers, body.
//
// Expected return value:
//   - None: request is not handled
//   - dict: {statusCode:int, headers:dict[str]str|list[str], body:str|bytes}
type StarlarkReqProcessor struct {
	Script      string
	ScriptName  string
	EntryPoint  string
	Predeclared starlark.StringDict
}

func (p StarlarkReqProcessor) Process(request corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	script := p.Script
	if strings.TrimSpace(script) == "" {
		script = defaultStarlarkProgram
	}

	thread := &starlark.Thread{Name: "req-processor"}
	handler, err := starlexec.LoadHandler(thread, p.scriptName(), script, p.entryPoint(), p.Predeclared)
	if err != nil {
		return optional.None[corehttp.HttpResponse]()
	}

	requestValue, err := starlexec.ToRequestValue(request)
	if err != nil {
		return optional.None[corehttp.HttpResponse]()
	}

	result, err := starlark.Call(thread, handler, starlark.Tuple{requestValue}, nil)
	if err != nil || result == starlark.None {
		return optional.None[corehttp.HttpResponse]()
	}

	response, err := starlexec.FromResponseValue(result)
	if err != nil {
		return optional.None[corehttp.HttpResponse]()
	}

	return optional.Some(response)
}

func (p StarlarkReqProcessor) entryPoint() string {
	if p.EntryPoint == "" {
		return starlexec.DefaultEntryPoint
	}

	return p.EntryPoint
}

func (p StarlarkReqProcessor) scriptName() string {
	if p.ScriptName == "" {
		return starlexec.DefaultScriptName
	}

	return p.ScriptName
}
