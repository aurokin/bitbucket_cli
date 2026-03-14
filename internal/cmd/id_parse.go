package cmd

import (
	"fmt"
	"strconv"
)

func parsePositiveInt(label, raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s ID must be a positive integer", label)
	}
	return value, nil
}
