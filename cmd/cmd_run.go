package cmd

import (
	"context"
	"errors"
	"fmt"

	"zktoro/zktoro-core-go/registry"

	"zktoro/zktoro-core-go/security"

	"zktoro/cmd/runner"

	"zktoro/store"

	"github.com/spf13/cobra"
)

// errors
var (
	ErrCannotRunScanner = errors.New("cannot run scanner")
)

func handlezktoroRun(cmd *cobra.Command, args []string) error {
	if err := checkScannerState(); err != nil {
		return err
	}
	if cfg.LocalModeConfig.Enable {
		whiteBold("Running in local mode...\n")
		if len(cfg.LocalModeConfig.WebhookURL) > 0 {
			yellowBold("Sending alerts to %s\n", cfg.LocalModeConfig.WebhookURL)
		} else {
			yellowBold("No webhook URL specified! Logging alerts in %s/logs/\n", cfg.ZktoroDir)
		}
	}
	runner.Run(cfg)
	return nil
}

func checkScannerState() error {
	scannerKey, err := security.LoadKeyWithPassphrase(cfg.KeyDirPath, cfg.Passphrase)
	if err != nil {
		return fmt.Errorf("failed to load scanner key: %v", err)
	}

	// disable registration and staking check in local mode
	if cfg.LocalModeConfig.Enable {
		return nil
	}
	// disable if flag was provided
	if parsedArgs.NoCheck {
		return nil
	}

	scannerAddressStr := scannerKey.Address.Hex()

	registry, err := store.GetRegistryClientWithoutENS(context.Background(), cfg, registry.ClientConfig{
		JsonRpcUrl: cfg.Registry.JsonRpc.Url,
		ENSAddress: cfg.ENSConfig.ContractAddress,
		Name:       "registry-client",
	})
	if err != nil {
		return fmt.Errorf("failed to create registry client: %v", err)
	}
	scanner, err := registry.GetScanner(scannerAddressStr)
	if err != nil {
		return fmt.Errorf("failed to check scanner state: %v", err)
	}

	// treat reverts the same as non-registered
	if scanner == nil {
		yellowBold("Scanner not registered - please make sure you register first.\n")
		toStderr("You can disable this behaviour with --no-check flag.\n")
		return ErrCannotRunScanner
	}
	if !scanner.Enabled {
		yellowBold("Warning! Your scan node is either disabled or does not meet with the minimum stake requirement. It will not receive any detection bots yet.\n")
	}
	return nil
}
