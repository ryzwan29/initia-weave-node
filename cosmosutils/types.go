package cosmosutils

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/initia-labs/weave/styles"
)

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func (coin Coin) IsZero() bool {
	normalizedAmount := strings.TrimSpace(coin.Amount)
	if normalizedAmount == "" || normalizedAmount == "0" {
		return true
	}

	amountBigInt := new(big.Int)
	amountBigInt, success := amountBigInt.SetString(normalizedAmount, 10)
	if !success {
		return false
	}

	return amountBigInt.Cmp(big.NewInt(0)) == 0
}

type Coins []Coin

func (cs *Coins) Render(maxWidth int) string {
	if len(*cs) == 0 {
		return styles.CreateFrame(NoBalancesText, maxWidth)
	}

	maxAmountLen := cs.getMaxAmountLength()

	var content strings.Builder
	for _, coin := range *cs {
		line := fmt.Sprintf("%-*s %s", maxAmountLen, coin.Amount, coin.Denom)
		content.WriteString(line + "\n")
	}

	contentStr := strings.TrimSuffix(content.String(), "\n")
	return styles.CreateFrame(contentStr, maxWidth)
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

func (cs *Coins) IsZero() bool {
	for _, coin := range *cs {
		if !coin.IsZero() {
			return false
		}
	}
	return true
}

type NodeInfoResponse struct {
	ApplicationVersion struct {
		Version string `json:"version"`
	} `json:"application_version"`
}
