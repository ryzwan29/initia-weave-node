# Launching your Rollup

Weave simplifies complicated rollup launch steps into a single command.

> Weave will send some funds from Gas Station to the OPinit Bot accounts during this process. Please make sure that your Gas Station account has enough funds to cover the total amount of funds to be sent (this amount will be shown to youbefore sending the funds).

```bash
weave rollup launch
```

Once the process completes, your rollup node will be running and ready to process queries and transactions.
The command also provides an [Initia Scan](https://scan.testnet.initia.xyz/) magic link that automatically adds your local rollup to the explorer, allowing you to instantly view your rollup's transactions and state.


> This command only sets up the bot addresses but does not start the OPinit Bots (executor and challenger). To complete the setup, proceed to the [OPinit Bots setup](/docs/opinit_bots.md) section to configure and run the OPinit Bots.

## Running your Rollup node

### Start the node

```bash
weave rollup start
```
Specify `--detach` or `-d` to run in the background.

### Stop the node

```bash
weave rollup stop
```

### Restart the node

```bash
weave rollup restart
```

### See the logs

```bash
weave rollup log
```

## Help

To see all the available commands:

```bash
weave rollup --help
```
