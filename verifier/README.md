# Babylon Verifier Program

This is a peripheral program that can be run by users of OP stack L2s to track consecutive L2 block quorum and query the BTC-finalised status of blocks via RESTful API endpoints. It is served as a Docker container for local use or hosted on a server.

## Running the Docker container

To get started, clone the repository and navigate to the `/verifier` directory.

```bash
git clone https://github.com/babylonchain/babylon-finality-gadget.git
cd verifier
```

Create a `.env` file in the `/verifier` directory with the following content:

```bash
L2_RPC_HOST=
BITCOIN_RPC_HOST=
PG_CONNECTION_STRING=
FG_CONTRACT_ADDRESS=
BBN_CHAIN_TYPE=
POLL_INTERVAL_IN_SECS=
```

Make sure you have Docker installed locally. If you don't, you can download it [here](https://www.docker.com/products/docker-desktop).

To run the Docker container, run:

```bash
docker compose up
```
