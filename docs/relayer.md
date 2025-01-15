# Running IBC Relayer

An IBC relayer is a software component that facilitates communication between two distinct blockchain networks that support the Inter-Blockchain Communication (IBC) protocol.

It is required for built-in oracle, Minitswap, and other cross-chain services to function with your rollup.

While setting up a relayer is traditionally one of the most complex tasks in the Cosmos ecosystem, 
Weave simplifies this process significantly, reducing the typical 10+ step IBC Relayer setup to just a few simple steps.

> Weave only supports IBC relayer setup between Initia L1 and Interwoven Rollups. Setting up relayers between other arbitrary networks is not supported.

> Currently, Weave only supports the `Hermes` relayer. For detailed information about Hermes, please refer to the [Hermes documentation](https://github.com/informalsystems/hermes).

## Setting up

```bash
weave relayer init
```

When initializing, Weave will present you with three options:
1. Set up a relayer between Initia L1 and a whitelisted Rollup (those available in [Initia Registry](https://github.com/initia-labs/initia-registry))
2. Configure manually
3. Configure using artifacts from `weave rollup launch` (recommended for users who have just launched their rollup)

> Please ensure that your Gas Station account has sufficient funds to cover the relayer's account funding requirements for both Initia L1 and your rollup.

> For advanced configuration options, you can refer to the [Hermes Configuration Guide](https://hermes.informal.systems/documentation/configuration/configure-hermes.html) and customize the relayer's configuration file located at `~/.hermes/config.toml`.

## Running Relayer

### Start the relayer

```bash
weave relayer start
```

Specify `--detach` or `-d` to run in the background.
Specify `--update-client false` to disable update IBC clients on relayer starts. Default to `true` 

### Stop the relayer

```bash
weave relayer stop
```

### Restart the relayer

```bash
weave relayer restart
```

### See the logs

```bash
weave relayer log
```

## Help

To see all the available commands: 
```bash
weave relayer --help
```
