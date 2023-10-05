package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"zktoro/config"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const minPassphraseLength = 12

func handlezktoroInit(cmd *cobra.Command, args []string) error {
	if isInitialized() {
		greenBold("Already initialized - please ensure that your configuration at %s is correct!\n", cfg.ConfigFilePath())
		return nil
	}

	if !isDirInitialized() {
		if err := os.Mkdir(cfg.ZktoroDir, 0755); err != nil {
			return err
		}
	}

	if !isConfigFileInitialized() {
		tmpl, err := template.New("config-template").Parse(defaultConfig)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, config.GetEnvDefaults(cfg.Development)); err != nil {
			return err
		}
		if err := os.WriteFile(cfg.ConfigFilePath(), buf.Bytes(), 0644); err != nil {
			return err
		}
	}

	if !isKeyDirInitialized() {
		if err := os.Mkdir(cfg.KeyDirPath, 0755); err != nil {
			return err
		}
	}

	if !isKeyInitialized() {
		if len(cfg.Passphrase) == 0 {
			yellowBold("Please provide a passphrase and do not lose it.\n\n")
			return cmd.Help()
		}

		if !isValidPassphrase(cfg.Passphrase) || len(cfg.Passphrase) < minPassphraseLength {
			yellowBold("Please provide an alphanumeric passphrase (a-z, A-Z, 0-9) with at least %d characters.\n\n", minPassphraseLength)
			return errors.New("invalid passphrase")
		}

		ks := keystore.NewKeyStore(cfg.KeyDirPath, keystore.StandardScryptN, keystore.StandardScryptP)
		acct, err := ks.NewAccount(cfg.Passphrase)
		if err != nil {
			return err
		}
		printScannerAddress(acct.Address.Hex())
	}

	color.Green("\nSuccessfully initialized at %s\n", cfg.ZktoroDir)
	whiteBold("\n%s\n", strings.Join([]string{
		"- Please make sure that all of the values in config.yml are set correctly.",
		"- Please register this node after making sure that you have staked enough.",
	}, "\n"))

	return nil
}

func isValidPassphrase(passphrase string) bool {
	matches, _ := regexp.MatchString(`([a-zA-Z0-9]+)`, passphrase)
	return matches
}

func printScannerAddress(address string) {
	fmt.Printf("\nScanner address: %s\n", color.New(color.FgYellow).Sprintf(address))
}

const defaultConfig = `# Auto generated by 'zktoro init' - safe to modify

# Chain ID of the network that is analyzed (1=mainnet)
# Set this before registering the node
chainId: 1

# Used for retrieving the blocks and transactions of the chain that is scanned
scan:
  jsonRpc:
    url: <required>

# Used for retrieving traces of all transactions in a block
# Must support trace_block (e.g. Alchemy)
trace:
  jsonRpc:
    url: <required>

# Used for loading assigned bots and detecting newer node versions
# Always set this as a reliable Polygon JSON-RPC API
# registry:
#  jsonRpc: run
#    url: <polygon-json-rpc-api>

# Used for allowing bots to make JSON-RPC requests
# If not set, defaults to scan.jsonRpc.url value above
# jsonRpcProxy:
#   jsonRpc:
#     url: <enter-if-different-from-scan-value>

# Adjust log level and sizes of scan node service containers
# log:
#  level: info
#  maxLogSize: 50m
#  maxLogFiles: 10
`

func isDirInitialized() bool {
	info, err := os.Stat(cfg.ZktoroDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isConfigFileInitialized() bool {
	info, err := os.Stat(cfg.ConfigFilePath())
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func isKeyDirInitialized() bool {
	info, err := os.Stat(cfg.KeyDirPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isKeyInitialized() bool {
	if !isKeyDirInitialized() {
		return false
	}
	entries, err := os.ReadDir(cfg.KeyDirPath)
	if err != nil {
		return false
	}
	for i, entry := range entries {
		if i > 0 {
			return false // There must be one key file
		}
		return !entry.IsDir() // so it should be a geth key file
	}
	return false // No keys found in dir
}

func isInitialized() bool {
	return isDirInitialized() && isConfigFileInitialized() && isKeyInitialized()
}
