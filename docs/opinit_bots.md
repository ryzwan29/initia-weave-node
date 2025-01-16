# Running OPinit Bots

Weave provides a streamlined way to configure and run [OPinit Bots](https://github.com/initia-labs/opinit-bots) (executor and challenger) for your rollup.

## Setting up

```bash
weave opinit init
```

This command will guide you through selecting the bot type (executor or challenger), configuring bot keys if needed, and setting up the bot's configuration.

You can also specify the bot type directly:

```bash
weave opinit init <executor|challenger>
```

## Managing Keys

To modify bot keys, use the following command to either generate new keys or restore existing ones:
```bash
weave opinit setup-keys
```

## Resetting OPinit Bots

Reset a bot's database. This will clear all the data stored in the bot's database (the configuration files are not affected).

```bash
weave opinit reset <executor|challenger>
```

## Running OPinit Bots

### Start the bot

```bash
weave opinit start <executor|challenger>
```
Specify `--detach` or `-d` to run in the background.

### Stop the bot

```bash
weave opinit stop <executor|challenger>
```

### Restart the bot

```bash
weave opinit restart <executor|challenger>
```

### See the logs

```bash
weave opinit log <executor|challenger>
```

## Help

To see all the available commands:
```bash
weave opinit --help
```
