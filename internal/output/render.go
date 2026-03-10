package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/itchyny/gojq"
)

type FormatOptions struct {
	AllFields  bool
	JSONFields []string
	JQ         string
}

func ParseFormatOptions(jsonFlag, jq string) (FormatOptions, error) {
	opts := FormatOptions{
		JQ: strings.TrimSpace(jq),
	}

	switch trimmed := strings.TrimSpace(jsonFlag); trimmed {
	case "":
	case "*":
		opts.AllFields = true
	default:
		for _, field := range strings.Split(trimmed, ",") {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			opts.JSONFields = append(opts.JSONFields, field)
		}
		if len(opts.JSONFields) == 0 {
			return FormatOptions{}, fmt.Errorf("no JSON fields selected")
		}
	}

	if opts.JQ != "" && !opts.AllFields && len(opts.JSONFields) == 0 {
		return FormatOptions{}, fmt.Errorf("--jq requires --json")
	}

	return opts, nil
}

func (o FormatOptions) UsesStructuredOutput() bool {
	return o.AllFields || len(o.JSONFields) > 0 || o.JQ != ""
}

func Render(w io.Writer, opts FormatOptions, data any, humanRenderer func(io.Writer) error) error {
	if !opts.UsesStructuredOutput() {
		return humanRenderer(w)
	}

	value, err := NormalizeValue(data)
	if err != nil {
		return err
	}

	if !opts.AllFields && len(opts.JSONFields) > 0 {
		value, err = projectFields(value, opts.JSONFields)
		if err != nil {
			return err
		}
	}

	if opts.JQ != "" {
		value, err = ApplyJQ(value, opts.JQ)
		if err != nil {
			return err
		}
	}

	return WriteJSON(w, value)
}

func NormalizeValue(data any) (any, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON output: %w", err)
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("normalize JSON output: %w", err)
	}

	return value, nil
}

func WriteJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	return encoder.Encode(value)
}

func projectFields(value any, fields []string) (any, error) {
	switch typed := value.(type) {
	case []any:
		projected := make([]any, 0, len(typed))
		for _, item := range typed {
			itemValue, err := projectFields(item, fields)
			if err != nil {
				return nil, err
			}
			projected = append(projected, itemValue)
		}
		return projected, nil
	case map[string]any:
		fieldSet := make(map[string]struct{}, len(fields))
		for _, field := range fields {
			fieldSet[field] = struct{}{}
		}

		projected := make(map[string]any, len(fieldSet))
		for key := range fieldSet {
			if value, ok := typed[key]; ok {
				projected[key] = value
			}
		}
		return projected, nil
	default:
		return nil, fmt.Errorf("cannot select JSON fields from %T", value)
	}
}

func ApplyJQ(value any, query string) (any, error) {
	parsed, err := gojq.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("parse jq expression: %w", err)
	}

	iter := parsed.Run(value)

	var results []any
	for {
		next, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := next.(error); ok {
			return nil, fmt.Errorf("evaluate jq expression: %w", err)
		}
		results = append(results, next)
	}

	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		return results[0], nil
	default:
		return results, nil
	}
}
