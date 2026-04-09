package starlexec

import (
	"fmt"

	corehttp "github.com/dector/lampa/server/core/http"
	"go.starlark.net/starlark"
)

func ToRequestValue(request corehttp.HttpRequest) (starlark.Value, error) {
	headers, err := toHeadersValue(request.Headers)
	if err != nil {
		return nil, err
	}

	d := starlark.NewDict(4)
	if err := d.SetKey(starlark.String("method"), starlark.String(request.Method)); err != nil {
		return nil, err
	}
	if err := d.SetKey(starlark.String("url"), starlark.String(request.Url.String())); err != nil {
		return nil, err
	}
	if err := d.SetKey(starlark.String("headers"), headers); err != nil {
		return nil, err
	}
	if err := d.SetKey(starlark.String("body"), starlark.Bytes(request.Body)); err != nil {
		return nil, err
	}

	return d, nil
}

func toHeadersValue(headers map[string][]string) (*starlark.Dict, error) {
	d := starlark.NewDict(len(headers))
	for name, values := range headers {
		list := make([]starlark.Value, 0, len(values))
		for _, value := range values {
			list = append(list, starlark.String(value))
		}

		if err := d.SetKey(starlark.String(name), starlark.NewList(list)); err != nil {
			return nil, fmt.Errorf("set header key %q: %w", name, err)
		}
	}

	return d, nil
}
