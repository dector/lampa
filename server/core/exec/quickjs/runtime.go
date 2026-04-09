package quickjsexec

import (
	"fmt"

	"modernc.org/quickjs"
)

const EntryPoint = "handle"

const responseDSLPrelude = `
(function initResponseDSL(global) {
	function ResponseBuilder(body, statusCode, headers) {
		this.statusCode = (typeof statusCode === "number") ? statusCode : 200;
		this.headers = headers || {};
		this.body = body;
	}

	ResponseBuilder.prototype.status = function(code) {
		this.statusCode = code;
		return this;
	};

	ResponseBuilder.prototype.header = function(name, value) {
		this.headers[name] = value;
		return this;
	};

	ResponseBuilder.prototype.contentType = function(value) {
		this.headers["Content-Type"] = value;
		return this;
	};

	function mergeHeaders(base, extra) {
		var merged = {};
		if (base) {
			for (var key in base) {
				if (Object.prototype.hasOwnProperty.call(base, key)) {
					merged[key] = base[key];
				}
			}
		}
		if (extra) {
			for (var key in extra) {
				if (Object.prototype.hasOwnProperty.call(extra, key)) {
					merged[key] = extra[key];
				}
			}
		}
		return merged;
	}

	global.Response = {
		json: function(data, statusCode, headers) {
			return new ResponseBuilder(
				JSON.stringify(data),
				(typeof statusCode === "number") ? statusCode : 200,
				mergeHeaders({"Content-Type": "application/json"}, headers)
			);
		},
		text: function(data, statusCode, headers) {
			return new ResponseBuilder(
				String(data),
				(typeof statusCode === "number") ? statusCode : 200,
				mergeHeaders({"Content-Type": "text/plain; charset=utf-8"}, headers)
			);
		},
		bytes: function(data, statusCode, headers) {
			return new ResponseBuilder(
				data,
				(typeof statusCode === "number") ? statusCode : 200,
				headers || {}
			);
		},
		empty: function(statusCode, headers) {
			return new ResponseBuilder(
				"",
				(typeof statusCode === "number") ? statusCode : 204,
				headers || {}
			);
		},
	};
})(globalThis);
`

func LoadHandler(vm *quickjs.VM, script string) error {
	if _, err := vm.Eval(responseDSLPrelude, quickjs.EvalGlobal); err != nil {
		return err
	}

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
