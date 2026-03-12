package cmd

import (
	"sync"
	"testing"
)

var commandTestHooksMu sync.Mutex

func lockCommandTestHooks(t *testing.T) {
	t.Helper()

	commandTestHooksMu.Lock()
	t.Cleanup(commandTestHooksMu.Unlock)
}
