package commands

import "kvnd/lazyruin/pkg/testutil"

// Re-export for backward compatibility within this test package.
type MockExecutor = testutil.MockExecutor

var NewMockExecutor = testutil.NewMockExecutor
