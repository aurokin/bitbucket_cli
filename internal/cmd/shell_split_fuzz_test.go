package cmd

import "testing"

func FuzzSplitCommandLine(f *testing.F) {
	for _, seed := range []string{
		`pr view 1`,
		`pr create --title "Add feature" --body 'Needs review'`,
		`pr create --title "Quote: \"x\""`,
		`pr create --repo acme/widgets\ beta`,
		`broken "quote`,
		`dangling\`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_, _ = splitCommandLine(input)
	})
}
