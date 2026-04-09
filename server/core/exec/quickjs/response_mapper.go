package quickjsexec

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"

	corehttp "github.com/dector/lampa/server/core/http"
	"modernc.org/quickjs"
)

type InvalidResponseTypeError struct {
	Actual any
}

func (e InvalidResponseTypeError) Error() string {
	return fmt.Sprintf("response must be object, got %T", e.Actual)
}

type InvalidStatusCodeTypeError struct {
	Actual any
}

func (e InvalidStatusCodeTypeError) Error() string {
	return fmt.Sprintf("statusCode must be number, got %T", e.Actual)
}

type InvalidHeadersTypeError struct {
	Actual any
}

func (e InvalidHeadersTypeError) Error() string {
	return fmt.Sprintf("headers must be object, got %T", e.Actual)
}

type InvalidHeaderValueTypeError struct {
	Name   string
	Actual any
}

func (e InvalidHeaderValueTypeError) Error() string {
	return fmt.Sprintf("header %q value must be string or string array, got %T", e.Name, e.Actual)
}

type InvalidBodyTypeError struct {
	Actual any
}

func (e InvalidBodyTypeError) Error() string {
	return fmt.Sprintf("body must be string or Uint8Array, got %T", e.Actual)
}

func FromResponseValue(value any) (corehttp.HttpResponse, error) {
	mapped, err := toResponseMap(value)
	if err != nil {
		return corehttp.HttpResponse{}, err
	}

	response := corehttp.HttpResponse{
		StatusCode: http.StatusOK,
		Headers:    make(http.Header),
	}

	if statusValue, found := mapped["statusCode"]; found {
		statusCode, err := parseStatusCode(statusValue)
		if err != nil {
			return corehttp.HttpResponse{}, err
		}
		response.StatusCode = statusCode
	}

	if headersValue, found := mapped["headers"]; found {
		headers, err := parseHeaders(headersValue)
		if err != nil {
			return corehttp.HttpResponse{}, err
		}
		response.Headers = headers
	}

	if bodyValue, found := mapped["body"]; found {
		body, err := parseBody(bodyValue)
		if err != nil {
			return corehttp.HttpResponse{}, err
		}
		response.Body = body
	}

	return response, nil
}

func toResponseMap(value any) (map[string]any, error) {
	switch typed := value.(type) {
	case *quickjs.Object:
		mapped := make(map[string]any)
		if err := typed.Into(&mapped); err != nil {
			return nil, err
		}
		return mapped, nil
	case map[string]any:
		return typed, nil
	default:
		return nil, InvalidResponseTypeError{Actual: value}
	}
}

func parseStatusCode(value any) (int, error) {
	number, ok := asNumber(value)
	if !ok {
		return 0, InvalidStatusCodeTypeError{Actual: value}
	}
	if math.Trunc(number) != number {
		return 0, InvalidStatusCodeTypeError{Actual: value}
	}

	return int(number), nil
}

func parseHeaders(value any) (http.Header, error) {
	headerMap, ok := value.(map[string]any)
	if !ok {
		return nil, InvalidHeadersTypeError{Actual: value}
	}

	headers := make(http.Header)
	for name, value := range headerMap {
		switch typed := value.(type) {
		case string:
			headers.Add(name, typed)
		case []any:
			for _, item := range typed {
				text, ok := item.(string)
				if !ok {
					return nil, InvalidHeaderValueTypeError{Name: name, Actual: value}
				}
				headers.Add(name, text)
			}
		default:
			return nil, InvalidHeaderValueTypeError{Name: name, Actual: value}
		}
	}

	return headers, nil
}

func parseBody(value any) ([]byte, error) {
	switch typed := value.(type) {
	case string:
		return []byte(typed), nil
	case []byte:
		return typed, nil
	case map[string]any:
		return parseByteMap(typed)
	case []any:
		return parseByteArray(typed)
	default:
		return nil, InvalidBodyTypeError{Actual: value}
	}
}

func parseByteMap(items map[string]any) ([]byte, error) {
	type indexedValue struct {
		index int
		value byte
	}

	values := make([]indexedValue, 0, len(items))
	for key, value := range items {
		index, err := strconv.Atoi(key)
		if err != nil {
			return nil, InvalidBodyTypeError{Actual: items}
		}
		byteValue, err := asByte(value)
		if err != nil {
			return nil, InvalidBodyTypeError{Actual: items}
		}
		values = append(values, indexedValue{index: index, value: byteValue})
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].index < values[j].index
	})

	result := make([]byte, 0, len(values))
	for _, item := range values {
		result = append(result, item.value)
	}
	return result, nil
}

func parseByteArray(items []any) ([]byte, error) {
	result := make([]byte, 0, len(items))
	for _, item := range items {
		value, err := asByte(item)
		if err != nil {
			return nil, InvalidBodyTypeError{Actual: items}
		}
		result = append(result, value)
	}

	return result, nil
}

func asNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case float32:
		return float64(typed), true
	case float64:
		return typed, true
	default:
		return 0, false
	}
}

func asByte(value any) (byte, error) {
	number, ok := asNumber(value)
	if !ok {
		return 0, fmt.Errorf("value is not number")
	}
	if math.Trunc(number) != number || number < 0 || number > 255 {
		return 0, fmt.Errorf("value is not byte")
	}

	return byte(number), nil
}
