# Babylon Finality Gadget

We proposed a [Babylon finality gadget](https://github.com/ethereum-optimism/specs/discussions/218) for OP-stack chains. The finality gadget depends on the EOTS data in a CosmWasm contract deployed on the Babylon chain. 

We have modified the OP-stack codebase to use the SDK in this codebase for additional finality checks.

In the future, we will also move the CosmWasm contract code here.

## Dependencies

The SDK requires a BTC RPC client defined in https://github.com/btcsuite/btcd/tree/master/rpcclient. We wrap it in our own BTC Client to make it easier to use.

## Usage

To run tests

```
make test
```
