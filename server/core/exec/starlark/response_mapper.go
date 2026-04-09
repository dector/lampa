package starlexec

import (
	"fmt"
	"net/http"

	corehttp "github.com/dector/lampa/server/core/http"
	"go.starlark.net/starlark"
)

func FromResponseValue(value starlark.Value) (corehttp.HttpResponse, error) {
	dict, ok := value.(*starlark.Dict)
	if !ok {
		return corehttp.HttpResponse{}, fmt.Errorf("response must be dict")
	}

	response := corehttp.HttpResponse{
		StatusCode: http.StatusOK,
		Headers:    make(http.Header),
	}

	if statusValue, found, err := dict.Get(starlark.String("statusCode")); err != nil {
		return corehttp.HttpResponse{}, err
	} else if found {
		var statusCode int64
		if err := starlark.AsInt(statusValue, &statusCode); err != nil {
			return corehttp.HttpResponse{}, fmt.Errorf("statusCode must be int: %w", err)
		}
		response.StatusCode = int(statusCode)
	}

	if headersValue, found, err := dict.Get(starlark.String("headers")); err != nil {
		return corehttp.HttpResponse{}, err
	} else if found {
		headers, err := parseHeaders(headersValue)
		if err != nil {
			return corehttp.HttpResponse{}, err
		}
		response.Headers = headers
	}

	if bodyValue, found, err := dict.Get(starlark.String("body")); err != nil {
		return corehttp.HttpResponse{}, err
	} else if found {
		body, err := parseBody(bodyValue)
		if err != nil {
			return corehttp.HttpResponse{}, err
		}
		response.Body = body
	}

	return response, nil
}

func parseBody(value starlark.Value) ([]byte, error) {
	switch typed := value.(type) {
	case starlark.String:
		return []byte(string(typed)), nil
	case starlark.Bytes:
		return []byte(typed), nil
	default:
		return nil, fmt.Errorf("body must be string or bytes")
	}
}

func parseHeaders(value starlark.Value) (http.Header, error) {
	headersDict, ok := value.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("headers must be dict")
	}

	headers := make(http.Header)
	for _, item := range headersDict.Items() {
		name, ok := item[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("header name must be string")
		}

		nameText := string(name)
		switch typed := item[1].(type) {
		case starlark.String:
			headers.Add(nameText, string(typed))
		case *starlark.List:
			values, err := parseStringList(typed)
			if err != nil {
				return nil, err
			}
			for _, v := range values {
				headers.Add(nameText, v)
			}
		default:
			return nil, fmt.Errorf("header value must be string or list of strings")
		}
	}

	return headers, nil
}

func parseStringList(list *starlark.List) ([]string, error) {
	iter := list.Iterate()
	defer iter.Done()

	values := make([]string, 0, list.Len())
	var v starlark.Value
	for iter.Next(&v) {
		s, ok := v.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("list item must be string")
		}
		values = append(values, string(s))
	}

	return values, nil
}
