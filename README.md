# Babylon DA SDK

We propoed a [Babylon finality gadget](https://github.com/ethereum-optimism/specs/discussions/218) for OP-stack chain. The finality gadget depends on EOTS data in a CosmWasm contract deployed on the Babylon chain. This essentially makes Babylon an extended DA for the OP-stack chains.

We will modify the OP-stack codebase to use this Go module for additional finalty checks.

In the future, we will also move the CosmWasm contract code here.

## Usages

To run the demo app
```
make run
```

To run tests
```
make test
```