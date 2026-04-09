package starlexec

import (
	"fmt"

	"go.starlark.net/starlark"
)

const (
	DefaultEntryPoint = "handle"
	DefaultScriptName = "handler.star"
)

func LoadHandler(
	thread *starlark.Thread,
	scriptName string,
	script string,
	entryPoint string,
	predeclared starlark.StringDict,
) (starlark.Value, error) {
	globals, err := starlark.ExecFile(thread, scriptName, script, predeclared)
	if err != nil {
		return nil, err
	}

	handler, ok := globals[entryPoint]
	if !ok {
		return nil, fmt.Errorf("entry point %q not found", entryPoint)
	}

	return handler, nil
}
