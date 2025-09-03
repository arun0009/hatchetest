#Hatchet Bug Client Initialization

Build and Run Tests:

go build -o hatchetest ./cmd
go test ./... -v (This should show below error)

```
    panic.go:262: test panicked: runtime error: invalid memory address or nil pointer dereference
        goroutine 8 [running]:
        runtime/debug.Stack()
        	/Users/arun/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.darwin-arm64/src/runtime/debug/stack.go:26 +0x64
        github.com/stretchr/testify/suite.failOnPanic(0x14000482fc0, {0x103d9c080, 0x10494c870})
        	/Users/arun/go/pkg/mod/github.com/stretchr/testify@v1.11.1/suite/suite.go:89 +0x38
        github.com/stretchr/testify/suite.recoverAndFailOnPanic(0x14000482fc0)
        	/Users/arun/go/pkg/mod/github.com/stretchr/testify@v1.11.1/suite/suite.go:83 +0x40
        panic({0x103d9c080?, 0x10494c870?})
        	/Users/arun/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.darwin-arm64/src/runtime/panic.go:792 +0x124
        github.com/hatchet-dev/hatchet/pkg/client/loader.GetClientConfigFromConfigFile(0x140005fd0e0)
        	/Users/arun/go/pkg/mod/github.com/hatchet-dev/hatchet@v0.71.14/pkg/client/loader/loader.go:89 +0x148
        github.com/hatchet-dev/hatchet/pkg/client/loader.(*ConfigLoader).LoadClientConfig(0x1400051f968?, 0x0)
        	/Users/arun/go/pkg/mod/github.com/hatchet-dev/hatchet@v0.71.14/pkg/client/loader/loader.go:36 +0xa0
        github.com/hatchet-dev/hatchet/pkg/client.defaultClientOpts(0x140000f4540?, 0x1?)
        	/Users/arun/go/pkg/mod/github.com/hatchet-dev/hatchet@v0.71.14/pkg/client/client.go:104 +0x74
        github.com/hatchet-dev/hatchet/pkg/client.New({0x1400051fbd0, 0x1, 0x1400051fc28?})
        	/Users/arun/go/pkg/mod/github.com/hatchet-dev/hatchet@v0.71.14/pkg/client/client.go:223 +0x7c
        github.com/arun0009/hatchetest/pkg/testsuite.(*SharedTestSuite).createHatchetClient(0x140002aa5b0, {0x103fb0518, 0x1049c24a0})
        	/Users/arun/development/workspace/github.com/arun0009/hatchetest/pkg/testsuite/shared_suite.go:425 +0x508
        github.com/arun0009/hatchetest/pkg/testsuite.(*SharedTestSuite).SetupSuite(0x140002aa5b0)
        	/Users/arun/development/workspace/github.com/arun0009/hatchetest/pkg/testsuite/shared_suite.go:286 +0x190
        github.com/arun0009/hatchetest.(*IntegrationTestSuite).SetupSuite(0x104950810?)
        	/Users/arun/development/workspace/github.com/arun0009/hatchetest/integration_test.go:20 +0x20
        github.com/stretchr/testify/suite.Run(0x14000482fc0, {0x103fad900, 0x140002aa5b0})
        	/Users/arun/go/pkg/mod/github.com/stretchr/testify@v1.11.1/suite/suite.go:211 +0x5fc
        github.com/arun0009/hatchetest.TestIntegrationSuite(0x14000482fc0)
        	/Users/arun/development/workspace/github.com/arun0009/hatchetest/integration_test.go:61 +0x3c
        testing.tRunner(0x14000482fc0, 0x103f998f0)
        	/Users/arun/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.darwin-arm64/src/testing/testing.go:1792 +0xe4
        created by testing.(*T).Run in goroutine 1
        	/Users/arun/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.6.darwin-arm64/src/testing/testing.go:1851 +0x374
--- FAIL: TestIntegrationSuite (10.19s)
```