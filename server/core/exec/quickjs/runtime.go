package quickjsexec

import (
	"fmt"

	"modernc.org/quickjs"
)

const EntryPoint = "handle"

func LoadHandler(vm *quickjs.VM, script string) error {
	if _, err := vm.Eval(script, quickjs.EvalGlobal); err != nil {
		return err
	}

	hasHandlerScript := fmt.Sprintf(`(typeof globalThis[%q] === "function")`, EntryPoint)
	hasHandler, err := vm.Eval(hasHandlerScript, quickjs.EvalGlobal)
	if err != nil {
		return err
	}

	isCallable, ok := hasHandler.(bool)
	if !ok || !isCallable {
		return fmt.Errorf("entry point %q not found", EntryPoint)
	}

	return nil
}
