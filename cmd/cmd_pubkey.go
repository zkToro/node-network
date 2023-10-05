package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func handleZktoroPubKey(cmd *cobra.Command, args []string) error {
	// ks := keystore.NewKeyStore(cfg.KeyDirPath, keystore.StandardScryptN, keystore.StandardScryptP)
	// acct := ks.Accounts()[0]
	// fmt.Println("Public Key: ", acct.Address.Hex())
	prvKeyBytes, err := os.ReadFile(cfg.DIDKeyPath)

	if err != nil {
		pub, prv, _ := ed25519.GenerateKey(rand.Reader)

		_ = os.WriteFile(cfg.DIDKeyPath, prv, 0644)
		fmt.Println("Private/Private key generated")
		fmt.Println("Node Private Key:")
		fmt.Println(hex.EncodeToString(prv))
		fmt.Println("Node Public Key:")
		fmt.Println(hex.EncodeToString(pub))
		// fmt.Println(hex.EncodeToString(prv))
	} else {
		fmt.Println("Node Public Key:")
		fmt.Println(hex.EncodeToString(prvKeyBytes)[64:])
	}
	return nil
}
