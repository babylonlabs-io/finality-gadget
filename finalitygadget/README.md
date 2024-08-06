# Babylon Finality Gadget

This is a peripheral program that can be run by users of OP stack L2s to track consecutive L2 block quorum and query the BTC-finalised status of blocks via RESTful API endpoints.

## Downloading and configuring the daemon

To get started, clone the repository.

```bash
git clone https://github.com/babylonlabs-io/finality-gadget.git
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

### Building and installing the binary

At the top-level directory of the project

```bash
make install
```

The above command will build and install the `opfgd` binary to
`$GOPATH/bin`:

- `eotsd`: The daemon program for the EOTS manager.
- `fpd`: The daemon program for the finality-provider with overall commands.

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
