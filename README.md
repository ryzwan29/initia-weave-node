# Weave

Weave is the only CLI tool you'll need for developing, testing, and running Interwoven Rollup successfully in production.

## Key Features

- **Bootstrap and Run Initia Node**: Set up an Initia node effortlessly. From configuring the desired network and syncing with an official node to running [Cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/main/tools/cosmovisor) for automatic upgrades, all with a single command.
- **Launch Interwoven Rollup**: Create your own Interwoven Rollup in minutes. Choose your preferred VM, customize the challenge period, select the DA option, and more, all with a single command.
- **Set Up and Run OPinit Bots**: [OPinit bots](https://github.com/initia-labs/opinit-bots) are essential for Initia's Optimistic bridge. Weave simplifies the setup process for you.
- **Set Up and Run Relayer**: Easily run a relayer between Initia L1 and any Rollup without navigating through extensive documentation. Currently, only [Hermes](https://github.com/informalsystems/hermes) is supported.

## Building from Scratch

This project requires Go `1.22` or higher.

To get started with Weave, ensure Go is installed on your system. Then, clone the repository and build and install the project by running:

```bash
make install
```

## Dependencies

Weave utilizes several external libraries:

- [`github.com/spf13/cobra`](https://github.com/spf13/cobra): For creating powerful modern CLI applications.
- [`github.com/charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea): For building terminal user interfaces.

## Contributing

We welcome contributions! Please feel free to submit a pull request.
