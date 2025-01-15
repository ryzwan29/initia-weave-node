# Bootstrapping Initia Node

Setting up a node for a Cosmos SDK chain has traditionally been a complex process requiring multiple steps:
- Locating the correct repository and version of the node binary compatible with your target network
- Either cloning and building the source code or downloading a pre-built binary from the release page
- Configuring the node with appropriate `config.toml` and `app.toml` files, which involves:
    - Setting correct values for `min_gas_price`, `seeds`, `persistent_peers`, and `pruning`
    - Navigating through numerous other parameters that rarely need modification
- Finding and implementing the correct genesis file to sync with the network
- Setting up cosmovisor for automatic updates or manually maintaining the node binary

Weave streamlines this entire process into a simple command.

## Initialize your node

```bash
weave initia init
```
This command guides you through the node setup process, taking you from an empty directory to a fully synced node ready for operation.
Once complete, you can run the node using `weave initia start`.


## Running your node

### Start the node

```bash
weave initia start
```
Specify `--detach` or `-d` to run in the background.

### Stop the node

```bash
weave initia stop
```

### Restart the node

```bash
weave initia restart
```

### See the logs

```bash
weave initia log
```

## Help

To see all the available commands: 
```bash
weave initia --help
```
