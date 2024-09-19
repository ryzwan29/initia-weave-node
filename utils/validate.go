package utils

import (
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/initia-labs/weave/bip39"
)

var (
	// Denominations can be 3 ~ 128 characters long and support letters, followed by either
	// a letter, a number or a separator ('/', ':', '.', '_' or '-').
	reDnmString = `[a-zA-Z][a-zA-Z0-9/:._-]{2,127}`
	reDecAmt    = `[[:digit:]]+(?:\.[[:digit:]]+)?|\.[[:digit:]]+`
	reSpc       = `[[:space:]]*`
	reDnm       *regexp.Regexp
	reDecCoin   *regexp.Regexp

	MaxBitLen = 256

	// number of decimal places
	LegacyPrecision = 18

	// bits required to represent the above precision
	// Ceiling[Log2[10^Precision - 1]]
	LegacyDecimalPrecisionBits = 60

	// decimalTruncateBits is the minimum number of bits removed
	// by a truncate operation. It is equal to
	// Floor[Log2[10^Precision - 1]].
	decimalTruncateBits = LegacyDecimalPrecisionBits - 1

	maxDecBitLen = MaxBitLen + decimalTruncateBits

	// Decimal errors
	ErrLegacyEmptyDecimalStr      = errors.New("decimal string cannot be empty")
	ErrLegacyInvalidDecimalLength = errors.New("invalid decimal length")
	ErrLegacyInvalidDecimalStr    = errors.New("invalid decimal string")
)

func init() {
	SetCoinDenomRegex(DefaultCoinDenomRegex)
}

// DefaultCoinDenomRegex returns the default regex string
func DefaultCoinDenomRegex() string {
	return reDnmString
}

// coinDenomRegex returns the current regex string and can be overwritten for custom validation
var coinDenomRegex = DefaultCoinDenomRegex

// SetCoinDenomRegex allows for coin's custom validation by overriding the regular
// expression string used for denom validation.
func SetCoinDenomRegex(reFn func() string) {
	coinDenomRegex = reFn

	reDnm = regexp.MustCompile(fmt.Sprintf(`^%s$`, coinDenomRegex()))
	reDecCoin = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reDecAmt, reSpc, coinDenomRegex()))
}

// ValidateDenom is the default validation function for Coin.Denom.
func ValidateDenom(denom string) error {
	if !reDnm.MatchString(denom) {
		return fmt.Errorf("invalid denom: %s", denom)
	}
	return nil
}

func NoOps(_ string) error {
	return nil
}

func ValidateDecCoin(coinStr string) (err error) {
	coinStr = strings.TrimSpace(coinStr)

	matches := reDecCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return fmt.Errorf("invalid decimal coin expression: %s", coinStr)
	}

	amountStr, denomStr := matches[1], matches[2]

	err = ValidateDecFromStr(amountStr)
	if err != nil {
		return fmt.Errorf("failed to parse decimal coin amount: %s", amountStr)
	}

	if err := ValidateDenom(denomStr); err != nil {
		return fmt.Errorf("invalid denom cannot contain spaces: %s", err)
	}

	return nil
}

func ValidateDecFromStr(str string) error {
	if str[0] == '-' {
		return fmt.Errorf("decimal string cannot be positve")
	}

	if len(str) == 0 {
		return ErrLegacyEmptyDecimalStr
	}

	strs := strings.Split(str, ".")
	lenDecs := 0
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		lenDecs = len(strs[1])
		if lenDecs == 0 || len(combinedStr) == 0 {
			return ErrLegacyInvalidDecimalLength
		}
		combinedStr += strs[1]
	} else if len(strs) > 2 {
		return ErrLegacyInvalidDecimalStr
	}

	if lenDecs > LegacyPrecision {
		return fmt.Errorf("value '%s' exceeds max precision by %d decimal places: max precision %d", str, LegacyPrecision-lenDecs, LegacyPrecision)
	}

	// add some extra zero's to correct to the Precision factor
	zerosToAdd := LegacyPrecision - lenDecs
	zeros := strings.Repeat("0", zerosToAdd)
	combinedStr += zeros

	combined, ok := new(big.Int).SetString(combinedStr, 10) // base 10
	if !ok {
		return fmt.Errorf("failed to set decimal string with base 10: %s", combinedStr)
	}
	if combined.BitLen() > maxDecBitLen {
		return fmt.Errorf("decimal '%s' out of range; bitLen: got %d, max %d", str, combined.BitLen(), maxDecBitLen)
	}

	return nil
}

func ValidateMnemonic(mnemonic string) error {
	if !bip39.IsMnemonicValid(mnemonic) {
		return errors.New("invalid bip39 mnemonic")
	}
	return nil
}

// Function to validate if a string is a valid URL and return an error if invalid
func ValidateURL(str string) error {
	u, err := url.Parse(str)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("URL is missing scheme or host")
	}
	return nil
}

// IsValidDNS checks if a given string is a valid DNS name
func IsValidDNS(dns string) bool {
	// Regular expression for validating DNS names
	dnsRegex := `^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	re := regexp.MustCompile(dnsRegex)

	// Validate DNS name
	return re.MatchString(dns)
}

// IsValidPeerOrSeed checks if each address in a comma-separated list is valid.
// It allows empty strings and returns an error with detailed reasons if any address is invalid.
// It accepts both IP addresses and DNS names.
func IsValidPeerOrSeed(addresses string) error {
	// Compile the regular expression for node ID
	nodeIDRegex, err := regexp.Compile(`^[a-f0-9]{40}$`)
	if err != nil {
		return fmt.Errorf("failed to compile nodeID regex: %v", err)
	}

	// Split the input string by commas to handle multiple addresses
	addressList := strings.Split(addresses, ",")

	var invalidAddresses []string

	// Iterate over each address and validate
	for _, address := range addressList {
		address = strings.TrimSpace(address) // Remove any leading/trailing spaces

		// Skip empty strings, as they're considered valid
		if address == "" {
			continue
		}

		parts := strings.Split(address, "@")
		if len(parts) != 2 {
			invalidAddresses = append(invalidAddresses, fmt.Sprintf("'%s': must be in format nodeID@ip_or_dns:port", address))
			continue
		}

		nodeID := parts[0]
		peerAddr := parts[1]

		// Validate node ID
		if !nodeIDRegex.MatchString(nodeID) {
			invalidAddresses = append(invalidAddresses, fmt.Sprintf("'%s': invalid node ID (must be 40-character hex string)", address))
			continue
		}

		// Split peer address into host (IP or DNS) and port
		host, port, err := net.SplitHostPort(peerAddr)
		if err != nil && !strings.Contains(err.Error(), "missing port in address") {
			invalidAddresses = append(invalidAddresses, fmt.Sprintf("'%s': invalid address (IP:Port or DNS:Port format required)", address))
			continue
		}

		// Validate host (can be IP or DNS)
		if net.ParseIP(host) == nil && !IsValidDNS(host) {
			invalidAddresses = append(invalidAddresses, fmt.Sprintf("'%s': invalid IP or DNS name", address))
			continue
		}

		// Validate port if present
		if port != "" {
			if _, err := fmt.Sscanf(port, "%d", new(int)); err != nil {
				invalidAddresses = append(invalidAddresses, fmt.Sprintf("'%s': invalid port", address))
				continue
			}
		}
	}

	// If there are any invalid addresses, return an error with detailed messages
	if len(invalidAddresses) > 0 {
		return errors.New("invalid peer/seed addresses: " + strings.Join(invalidAddresses, ", "))
	}

	return nil
}
