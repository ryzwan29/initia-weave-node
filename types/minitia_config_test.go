package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the creation of a new SystemAccount using NewSystemAccount function
func TestNewSystemAccount(t *testing.T) {
	// Create new system account with both L1 and L2 balances
	account := NewSystemAccount("some-mnemonic", "l1-address")

	// Validate account has the correct L1 and L2 addresses
	assert.Equal(t, "l1-address", account.L1Address, "Expected L1 address to be 'l1-address'")
	assert.Equal(t, "some-mnemonic", account.Mnemonic, "Expected mnemonic to be 'some-mnemonic'")
	assert.Equal(t, "l1-address", account.L2Address, "Expected L2 address to be 'l1-address'")

	// Create system account with only L1 balance
	accountL1 := NewSystemAccount("mnemonic-l1", "l1-only-address")
	assert.Equal(t, "l1-only-address", accountL1.L1Address, "Expected L1 address to be 'l1-only-address'")
	assert.Equal(t, "l1-only-address", accountL1.L2Address, "Expected L2 address to be empty")

	// Create system account with only L2 balance
	accountL2 := NewSystemAccount("mnemonic-l2", "l2-only-address")
	assert.Equal(t, "l2-only-address", accountL2.L2Address, "Expected L2 address to be 'l2-only-address'")
	assert.Equal(t, "l2-only-address", accountL2.L1Address, "Expected L1 address to be empty")
}

// Test the behavior of the GenesisAccounts struct
func TestGenesisAccounts(t *testing.T) {
	// Create a slice of GenesisAccounts
	accounts := GenesisAccounts{
		{Address: "address1", Coins: "100coins"},
		{Address: "address2", Coins: "200coins"},
	}

	// Validate the genesis accounts
	assert.Equal(t, 2, len(accounts), "Expected 2 genesis accounts")
	assert.Equal(t, "address1", accounts[0].Address, "Expected first account address to be 'address1'")
	assert.Equal(t, "100coins", accounts[0].Coins, "Expected first account coins to be '100coins'")
	assert.Equal(t, "address2", accounts[1].Address, "Expected second account address to be 'address2'")
	assert.Equal(t, "200coins", accounts[1].Coins, "Expected second account coins to be '200coins'")
}
