package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/utils"
)

const NoBalancesText string = "No Balances"

func GasStationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "gas-station",
		Short:                      "Gas Station subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		gasStationSetupCommand(),
		gasStationShowCommand(),
	)

	return cmd
}

func gasStationSetupCommand() *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup Gas Station account on Initia and Celestia for funding the OPinit-bots or relayer to send transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := tea.NewProgram(models.NewGasStationMnemonicInput("")).Run()
			if err != nil {
				return err
			}

			return showGasStationBalance()
		},
	}

	return setupCmd
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Coins []Coin

func (cs *Coins) Render(maxWidth int) string {
	if len(*cs) == 0 {
		return createFrame(NoBalancesText, maxWidth)
	}

	maxAmountLen := cs.getMaxAmountLength()

	var content strings.Builder
	for _, coin := range *cs {
		line := fmt.Sprintf("%-*s %s", maxAmountLen, coin.Amount, coin.Denom)
		content.WriteString(line + "\n")
	}

	contentStr := strings.TrimSuffix(content.String(), "\n")
	return createFrame(contentStr, maxWidth)
}

func createFrame(text string, maxWidth int) string {
	lines := strings.Split(text, "\n")
	top := "┌" + strings.Repeat("─", maxWidth+2) + "┐"
	bottom := "└" + strings.Repeat("─", maxWidth+2) + "┘"

	var framedContent strings.Builder
	for _, line := range lines {
		framedContent.WriteString(fmt.Sprintf("│ %-*s │\n", maxWidth, line))
	}

	return fmt.Sprintf("%s\n%s%s", top, framedContent.String(), bottom)
}

func (cs *Coins) getMaxAmountLength() int {
	maxLen := 0
	for _, coin := range *cs {
		if len(coin.Amount) > maxLen {
			maxLen = len(coin.Amount)
		}
	}
	return maxLen
}

func getBalance(chainType registry.ChainType, address string) (*Coins, error) {
	chainRegistry, err := registry.GetChainRegistry(chainType)
	if err != nil {
		return nil, fmt.Errorf("failed to load chainRegistry: %v", err)
	}

	baseUrl, err := chainRegistry.GetActiveLcd()
	if err != nil {
		return nil, fmt.Errorf("failed to get active lcd for %s: %v", chainType, err)
	}

	client := utils.NewHTTPClient()
	var result map[string]interface{}
	_, err = client.Get(
		baseUrl,
		fmt.Sprintf("/cosmos/bank/v1beta1/balances/%s", address),
		map[string]string{"pagination.limit": "100"},
		&result,
	)
	if err != nil {
		return nil, err
	}

	rawBalances, ok := result["balances"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse balances field")
	}

	balancesJSON, err := json.Marshal(rawBalances)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal balances: %w", err)
	}

	var balances Coins
	err = json.Unmarshal(balancesJSON, &balances)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal balances into Coins: %w", err)
	}

	return &balances, nil
}

func getMaxWidth(coinGroups ...*Coins) int {
	maxAmountWidth := 0
	maxDenomWidth := 0

	for _, coins := range coinGroups {
		for _, coin := range *coins {
			if len(coin.Amount) > maxAmountWidth {
				maxAmountWidth = len(coin.Amount)
			}
			if len(coin.Denom) > maxDenomWidth {
				maxDenomWidth = len(coin.Denom)
			}
		}
	}

	// Add 1 space here for the space between amount and denom
	return maxAmountWidth + maxDenomWidth + 1
}

func showGasStationBalance() error {
	gasStationMnemonic := utils.GetGasStationMnemonic()
	initiaGasStationAddress, err := utils.MnemonicToBech32Address("init", gasStationMnemonic)
	if err != nil {
		return err
	}
	celestiaGasStationAddress, err := utils.MnemonicToBech32Address("celestia", gasStationMnemonic)
	if err != nil {
		return err
	}

	initiaL1TestnetBalances, err := getBalance(registry.InitiaL1Testnet, initiaGasStationAddress)
	if err != nil {
		return err
	}

	celestiaTestnetBalance, err := getBalance(registry.CelestiaTestnet, celestiaGasStationAddress)
	if err != nil {
		return err
	}

	celestiaMainnetBalance, err := getBalance(registry.CelestiaMainnet, celestiaGasStationAddress)
	if err != nil {
		return err
	}

	maxWidth := getMaxWidth(initiaL1TestnetBalances, celestiaTestnetBalance, celestiaMainnetBalance)
	if maxWidth < len(NoBalancesText) {
		maxWidth = len(NoBalancesText)
	}

	fmt.Println(fmt.Sprintf("\n⛽️ Initia Testnet Address: %s\n%s\n", initiaGasStationAddress, initiaL1TestnetBalances.Render(maxWidth)))
	fmt.Println(fmt.Sprintf("⛽️ Celestia Testnet Address: %s\n%s\n", celestiaGasStationAddress, celestiaTestnetBalance.Render(maxWidth)))
	fmt.Println(fmt.Sprintf("⛽️ Celestia Mainnet Address: %s\n%s\n", celestiaGasStationAddress, celestiaMainnetBalance.Render(maxWidth)))

	return nil
}

func gasStationShowCommand() *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show Initia and Celestia Gas Station addresses and balances",
		RunE: func(cmd *cobra.Command, args []string) error {
			if utils.IsFirstTimeSetup() {
				fmt.Println("Please setup Gas Station first, by running `gas-station setup`")
				return nil
			}

			return showGasStationBalance()
		},
	}

	return showCmd
}
