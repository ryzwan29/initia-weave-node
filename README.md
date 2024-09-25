# Weave

Weave is a powerful command-line tool designed for managing Initia deployments and interacting with Initia and Minitia nodes.

## Available Commands

Weave currently offers the following main commands and subcommands:

### `weave init`

Initializes the Weave CLI, funding the gas station and setting up the configuration.

### `weave initia`

Manages Initia full node operations with the following subcommands:

- `weave initia init`: Bootstraps your Initia full node.
- `weave initia start`: Starts the initiad full node application.
- `weave initia stop`: Stops the initiad full node application.
- `weave initia log`: Streams the logs of the initiad full node application.

### `weave minitia`

Manages Minitia operations with the following subcommand:

- `weave minitia launch`: Launches a new Minitia from scratch.

## Building from scratch

To get started with Weave, make sure you have Go installed on your system. Then, clone the repository then build and install the project by calling

```
make install
```

## Dependencies

Weave uses a few external libraries:

- `github.com/spf13/cobra`: For creating powerful modern CLI applications.
- `github.com/charmbracelet/bubbletea`: For building terminal user interfaces.

## Contributing

We welcome contributions! Please feel free to submit a Pull Request.
