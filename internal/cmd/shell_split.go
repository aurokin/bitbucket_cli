package cmd

import (
	"fmt"
	"strings"
)

func splitCommandLine(input string) ([]string, error) {
	var args []string
	var current strings.Builder

	const (
		stateUnquoted = iota
		stateSingleQuoted
		stateDoubleQuoted
	)

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
			switch ch {
			case ' ', '\t', '\n':
				flush()
			case '\'':
				state = stateSingleQuoted
				tokenStarted = true
			case '"':
				state = stateDoubleQuoted
				tokenStarted = true
			case '\\':
				if i+1 >= len(input) {
					return nil, fmt.Errorf("dangling escape")
				}
				tokenStarted = true
				i++
				current.WriteByte(input[i])
			default:
				tokenStarted = true
				current.WriteByte(ch)
			}
		case stateSingleQuoted:
			if ch == '\'' {
				state = stateUnquoted
				continue
			}
			current.WriteByte(ch)
		case stateDoubleQuoted:
			switch ch {
			case '"':
				state = stateUnquoted
			case '\\':
				if i+1 >= len(input) {
					return nil, fmt.Errorf("dangling escape")
				}
				i++
				escaped := input[i]
				switch escaped {
				case '"', '\\', ' ':
					current.WriteByte(escaped)
				default:
					current.WriteByte('\\')
					current.WriteByte(escaped)
				}
			default:
				current.WriteByte(ch)
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
