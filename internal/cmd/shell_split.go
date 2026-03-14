package cmd

import (
	"fmt"
	"strings"
)

type shellSplitState int

const (
	stateUnquoted shellSplitState = iota
	stateSingleQuoted
	stateDoubleQuoted
)

func splitCommandLine(input string) ([]string, error) {
	var args []string
	var current strings.Builder

	state := stateUnquoted
	tokenStarted := false

	flush := func() {
		if !tokenStarted {
			return
		}
		args = append(args, current.String())
		current.Reset()
		tokenStarted = false
	}

	for i := 0; i < len(input); i++ {
		ch := input[i]
		switch state {
		case stateUnquoted:
			if err := processUnquotedShellChar(input, &i, ch, &current, &state, &tokenStarted, flush); err != nil {
				return nil, err
			}
		case stateSingleQuoted:
			if ch == '\'' {
				state = stateUnquoted
				continue
			}
			current.WriteByte(ch)
		case stateDoubleQuoted:
			if err := processDoubleQuotedShellChar(input, &i, ch, &current, &state); err != nil {
				return nil, err
			}
		}
	}

	switch state {
	case stateSingleQuoted:
		return nil, fmt.Errorf("unterminated single quote")
	case stateDoubleQuoted:
		return nil, fmt.Errorf("unterminated double quote")
	}

	flush()
	return args, nil
}

func processUnquotedShellChar(input string, index *int, ch byte, current *strings.Builder, state *shellSplitState, tokenStarted *bool, flush func()) error {
	switch ch {
	case ' ', '\t', '\n':
		flush()
	case '\'':
		*state = stateSingleQuoted
		*tokenStarted = true
	case '"':
		*state = stateDoubleQuoted
		*tokenStarted = true
	case '\\':
		next, err := nextEscapedShellChar(input, index)
		if err != nil {
			return err
		}
		*tokenStarted = true
		current.WriteByte(next)
	default:
		*tokenStarted = true
		current.WriteByte(ch)
	}
	return nil
}

func processDoubleQuotedShellChar(input string, index *int, ch byte, current *strings.Builder, state *shellSplitState) error {
	switch ch {
	case '"':
		*state = stateUnquoted
	case '\\':
		escaped, err := nextEscapedShellChar(input, index)
		if err != nil {
			return err
		}
		writeEscapedDoubleQuotedChar(current, escaped)
	default:
		current.WriteByte(ch)
	}
	return nil
}

func nextEscapedShellChar(input string, index *int) (byte, error) {
	if *index+1 >= len(input) {
		return 0, fmt.Errorf("dangling escape")
	}
	*index++
	return input[*index], nil
}

func writeEscapedDoubleQuotedChar(current *strings.Builder, escaped byte) {
	switch escaped {
	case '"', '\\', ' ':
		current.WriteByte(escaped)
	default:
		current.WriteByte('\\')
		current.WriteByte(escaped)
	}
}
