# Babylon Verifier Daemon

This is a peripheral program that can be run by users of OP stack L2s to track consecutive L2 block quorum and query the BTC-finalised status of blocks via RESTful API endpoints.

## Running the daemon

To get started, clone the repository.

```bash
git clone https://github.com/babylonchain/babylon-finality-gadget.git
```

Configure the `config.toml` file with the following content:

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

To start the daemon, navigate to the `/cmd` directory:

```bash
cd cmd
go run . start --cfg ../config.toml
```

<!-- ## Running the Docker container -->

<!-- Make sure you have Docker installed locally. If you don't, you can download it [here](https://www.docker.com/products/docker-desktop).

To run the Docker container, run:

```bash
docker compose up
``` -->
