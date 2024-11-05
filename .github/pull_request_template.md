!!! Please test the code with https://github.com/babylonlabs-io/finality-provider

Steps to test:
1. `cd finality-provider`
2. `git checkout base/consumer-chain-support`
3. modify the `go.mod` to use the tip commit hash of the branch for this PR
4. `go mod tidy`
5. run `make test-e2e-op`
