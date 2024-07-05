# Babylon DA SDK

We proposed a [Babylon finality gadget](https://github.com/ethereum-optimism/specs/discussions/218) for OP-stack chains. The finality gadget depends on the EOTS data in a CosmWasm contract deployed on the Babylon chain. This essentially makes Babylon an extended DA for the OP-stack chains.

We will modify the OP-stack codebase to use this SDK for additional finalty checks.

In the future, we will also move the CosmWasm contract code here.

## Dependencies

The DA SDK requires a BTC RPC client defined in https://github.com/btcsuite/btcd/tree/master/rpcclient. We wrap it in our own BTC Client to make it easier to use.

## Usages

To run the demo app

```
make run
```

To run tests

```
make test
```
