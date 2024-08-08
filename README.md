# Babylon Finality Gadget

The Babylon Finality Gadget is a program that can be run by users of OP stack L2s to track consecutive L2 block quorum and query the BTC-finalised status of blocks.

See our [proposal](https://github.com/ethereum-optimism/specs/discussions/218) on Optimism for more details.

## Modules

- `cmd` : entry point for `opfgd` finality gadget daemon
- `finalitygadget` : top-level umbrella module that exposes query methods and coordinates calls to other clients
- `client` : grpc client to query the finality gadget
- `server` : grpc server for the finality gadget
- `proto` : protobuf definitions for the grpc server
- `config` : configs for the finality gadget
- `btcclient` : wrapper around Bitcoin RPC client
- `bbnclient` : wrapper around Babylon RPC client
- `ethl2client` : wrapper around OP stack L2 ETH RPC client
- `cwclient` : client to query CosmWasm smart contract deployed on BabylonChain
- `db` : handler for local database to store finalized block state
- `types` : common types
- `log` : custom logger
- `testutil` : test utilities and helpers

## Instructions

### Download and configuration

To get started, clone the repository.

```bash
git clone https://github.com/babylonlabs-io/finality-gadget.git
```

Copy the `config.toml.example` file to `config.toml`:

```bash
cp config.toml.example config.toml
```

Configure the `config.toml` file with the following parameters:

```toml
L2RPCHost = # RPC URL of OP stack L2 chain
BitcoinRPCHost = # Bitcoin RPC URL
DBFilePath = # Path to local bbolt DB file
FGContractAddress = # Babylon finality gadget contract address
BBNChainID = # Babylon chain id
BBNRPCAddress = # Babylon RPC host URL
GRPCServerPort = # Port to run the gRPC server on
PollInterval = # Interval to poll for new L2 blocks
```

### Building and installing the binary

At the top-level directory of the project

```bash
make install
```

The above command will build and install the `opfgd` binary to
`$GOPATH/bin`.

If your shell cannot find the installed binaries, make sure `$GOPATH/bin` is in
the `$PATH` of your shell. Usually these commands will do the job

```bash
export PATH=$HOME/go/bin:$PATH
echo 'export PATH=$HOME/go/bin:$PATH' >> ~/.profile
```

### Running the daemon

To start the daemon, run:

```bash
opfgd start --cfg config.toml
```

### Running tests

To run tests:

```bash
make test
```
