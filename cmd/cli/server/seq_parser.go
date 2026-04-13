package server

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type seqStepInput struct {
	Index   int
	Body    string
	Status  *int
	Content *string
	Headers []string
}

func parseSeqStepInputsFromRawArgv(rawArgv []string) ([]seqStepInput, error) {
	type stepState struct {
		seqStepInput
		hasBody    bool
		hasStatus  bool
		hasContent bool
	}

	stepsByIndex := map[int]*stepState{}
	foundAny := false

	for i := 0; i < len(rawArgv); i++ {
		token := rawArgv[i]
		if !strings.HasPrefix(token, "--response.") {
			continue
		}

		flagName, index, inlineValue, isIndexed, err := parseIndexedSeqFlagToken(token)
		if err != nil {
			return nil, err
		}
		if !isIndexed {
			continue
		}
		foundAny = true

		value := inlineValue
		if value == "" {
			if i+1 >= len(rawArgv) {
				return nil, fmt.Errorf("missing value for --%s-%d", flagName, index)
			}
			i++
			value = rawArgv[i]
		}

		step := stepsByIndex[index]
		if step == nil {
			step = &stepState{seqStepInput: seqStepInput{Index: index}}
			stepsByIndex[index] = step
		}

		switch flagName {
		case OptResponseBody:
			if step.hasBody {
				return nil, fmt.Errorf("duplicate declaration for --%s-%d", OptResponseBody, index)
			}
			step.hasBody = true
			step.Body = value
		case OptResponseStatus:
			if step.hasStatus {
				return nil, fmt.Errorf("duplicate declaration for --%s-%d", OptResponseStatus, index)
			}
			parsedStatus, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("invalid value for --%s-%d: %q", OptResponseStatus, index, value)
			}
			step.hasStatus = true
			step.Status = &parsedStatus
		case OptResponseContent:
			if step.hasContent {
				return nil, fmt.Errorf("duplicate declaration for --%s-%d", OptResponseContent, index)
			}
			parsedContent := strings.TrimSpace(value)
			step.hasContent = true
			step.Content = &parsedContent
		case OptResponseHeader:
			step.Headers = append(step.Headers, value)
		}
	}

	if !foundAny {
		return nil, fmt.Errorf("missing indexed sequence flags: expected --%s-N", OptResponseBody)
	}

	indices := make([]int, 0, len(stepsByIndex))
	for index := range stepsByIndex {
		indices = append(indices, index)
	}
	sort.Ints(indices)

	if len(indices) == 0 {
		return nil, fmt.Errorf("no sequence steps declared")
	}

	for expected := 1; expected <= len(indices); expected++ {
		if indices[expected-1] != expected {
			return nil, fmt.Errorf("sequence index gap: expected step %d", expected)
		}
	}

	steps := make([]seqStepInput, 0, len(indices))
	for _, index := range indices {
		step := stepsByIndex[index]
		if !step.hasBody || strings.TrimSpace(step.Body) == "" {
			return nil, fmt.Errorf("step %d: missing response body", index)
		}
		steps = append(steps, step.seqStepInput)
	}

	return steps, nil
}

func parseIndexedSeqFlagToken(token string) (flagName string, index int, inlineValue string, isIndexed bool, err error) {
	for _, candidate := range []string{OptResponseBody, OptResponseStatus, OptResponseContent, OptResponseHeader} {
		prefix := "--" + candidate + "-"
		if !strings.HasPrefix(token, prefix) {
			continue
		}

		suffix := token[len(prefix):]
		if suffix == "" {
			return "", 0, "", true, fmt.Errorf("invalid indexed flag %q: expected positive numeric suffix", token)
		}

		indexPart := suffix
		inlinePart := ""
		if cut := strings.Index(indexPart, "="); cut >= 0 {
			inlinePart = indexPart[cut+1:]
			indexPart = indexPart[:cut]
		}

		parsedIndex, parseErr := strconv.Atoi(indexPart)
		if parseErr != nil || parsedIndex < 1 {
			return "", 0, "", true, fmt.Errorf("invalid indexed flag %q: expected positive numeric suffix", token)
		}

		return candidate, parsedIndex, inlinePart, true, nil
	}

	for _, candidate := range []string{OptResponseBody, OptResponseStatus, OptResponseContent, OptResponseHeader} {
		if strings.HasPrefix(token, "--"+candidate) {
			return "", 0, "", false, fmt.Errorf("invalid indexed flag %q: expected -N suffix", token)
		}
	}

	return "", 0, "", false, nil
}
