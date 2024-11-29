package crypto

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
)

const (
	CosmosHDPath   string = "m/44'/118'/0'/0/0"
	HardenedOffset int    = 0x80000000
)

// MnemonicToBech32Address converts a mnemonic to a Cosmos SDK Bech32 address.
func MnemonicToBech32Address(hrp, mnemonic string) (string, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return "", fmt.Errorf("failed to generate seed: %w", err)
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return "", fmt.Errorf("failed to derive master key: %w", err)
	}

	derivedKey, err := deriveKey(masterKey, CosmosHDPath)
	if err != nil {
		return "", fmt.Errorf("failed to derive child key: %w", err)
	}

	_, pubKey := btcec.PrivKeyFromBytes(derivedKey.Key)
	pubKeyBytes := pubKey.SerializeCompressed()

	shaHash := sha256.Sum256(pubKeyBytes)
	ripemd := ripemd160.New()
	ripemd.Write(shaHash[:])
	addressHash := ripemd.Sum(nil)

	converted, err := bech32.ConvertBits(addressHash, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("failed to convert to Bech32: %w", err)
	}

	bech32Addr, err := bech32.Encode(hrp, converted)
	if err != nil {
		return "", fmt.Errorf("failed to encode to Bech32: %w", err)
	}

	return bech32Addr, nil
}

// deriveKey derives the private key along the given HD path.
func deriveKey(masterKey *bip32.Key, path string) (*bip32.Key, error) {
	key := masterKey
	var err error

	for _, n := range parseHDPath(path) {
		key, err = key.NewChildKey(n)
		if err != nil {
			return nil, err
		}
	}
	return key, nil
}

// parseHDPath parses the hd path string
func parseHDPath(path string) []uint32 {
	parts := strings.Split(path, "/")[1:]
	keys := make([]uint32, len(parts))

	for i, part := range parts {
		hardened := strings.HasSuffix(part, "'")
		if hardened {
			part = strings.TrimSuffix(part, "'")
		}

		n, _ := strconv.Atoi(part)
		if hardened {
			n = n + HardenedOffset
		}
		keys[i] = uint32(n)
	}
	return keys
}

// GenerateMnemonic generates new fresh mnemonic
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	return mnemonic, nil
}
