package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func resolveCommandVariableValue(stdin io.Reader, value, valueFile, kind string) (string, error) {
	if strings.TrimSpace(valueFile) != "" {
		data, err := readRequestBody(stdin, valueFile)
		if err != nil {
			return "", err
		}
		value = string(data)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("provide a %s value with --value or --value-file", kind)
	}
	return value, nil
}

func parseBoolString(raw string) (bool, error) {
	return strconv.ParseBool(strings.TrimSpace(raw))
}

func resolveVariableReference[T any](
	raw string,
	kind string,
	getByUUID func(string) (T, error),
	list func() ([]T, error),
	matches func(string, T) bool,
) (T, error) {
	var zero T

	reference := strings.TrimSpace(raw)
	if reference == "" {
		return zero, fmt.Errorf("%s reference is required", kind)
	}
	if strings.HasPrefix(reference, "{") && strings.HasSuffix(reference, "}") && getByUUID != nil {
		return getByUUID(reference)
	}

	values, err := list()
	if err != nil {
		return zero, err
	}

	var matchesFound []T
	for _, item := range values {
		if matches(reference, item) {
			matchesFound = append(matchesFound, item)
		}
	}

	switch len(matchesFound) {
	case 1:
		return matchesFound[0], nil
	case 0:
		return zero, fmt.Errorf("%s %q was not found", kind, reference)
	default:
		return zero, fmt.Errorf("%s %q is ambiguous; use a UUID instead", kind, reference)
	}
}
