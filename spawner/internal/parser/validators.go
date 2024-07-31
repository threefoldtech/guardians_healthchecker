package parser

import (
	"fmt"

	"github.com/cosmos/go-bip39"
)

func validateMnemonic(mnemonic string) error {
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("invalid mnemonic: '%s'", mnemonic)
	}
	return nil
}
