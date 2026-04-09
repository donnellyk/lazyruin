package commands

import "github.com/donnellyk/lazyruin/pkg/testutil"

// Re-export for backward compatibility within this test package.
type MockExecutor = testutil.MockExecutor

var NewMockExecutor = testutil.NewMockExecutor
