package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"zktoro/store"

	"zktoro/zktoro-core-go/registry"
	"zktoro/zktoro-core-go/security"
	"zktoro/zktoro-core-go/security/eip712"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func handlezktoroAuthorizePool(cmd *cobra.Command, args []string) error {
	poolIDStr, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	poolID, err := strconv.ParseInt(poolIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to decode pool ID: %v", err)
	}

	polygonscan, _ := cmd.Flags().GetBool("polygonscan")
	force, _ := cmd.Flags().GetBool("force")
	clean, _ := cmd.Flags().GetBool("clean")

	scannerKey, err := security.LoadKeyWithPassphrase(cfg.KeyDirPath, cfg.Passphrase)
	if err != nil {
		return fmt.Errorf("failed to load scanner key: %v", err)
	}
	scannerPrivateKey := scannerKey.PrivateKey

	fmt.Println("ENSConfig ", cfg.ENSConfig.ContractAddress)
	fmt.Println("JsonRpcUrl ", cfg.Registry.JsonRpc.Url)
	fmt.Println("scannerPrivateKey ", scannerPrivateKey)

	regClient, err := store.GetRegistryClientWithoutENS(context.Background(), cfg, registry.ClientConfig{
		JsonRpcUrl: cfg.Registry.JsonRpc.Url,
		ENSAddress: cfg.ENSConfig.ContractAddress,
		Name:       "registry-client",
		PrivateKey: scannerPrivateKey,
	})

	if err != nil {
		return fmt.Errorf("failed to create registry client: %v", err)
	}

	return authorizePoolWithRegistry(regClient, scannerKey, poolID, polygonscan, force, clean)
}

func authorizePoolWithRegistry(
	regClient registry.Client,
	scannerKey *keystore.Key,
	poolID int64, polygonscan, force, clean bool,
) error {
	regClient.SetRegistryChainID(cfg.Registry.ChainID)

	scanner, err := regClient.GetPoolScanner(scannerKey.Address.Hex())
	if err != nil {
		return fmt.Errorf("failed to get scanner from registry: %v", err)
	}
	if scanner != nil && !force {
		color.New(color.FgYellow).Printf("This scanner is already registered to pool %s!\n", scanner.PoolID)
		return nil
	}
	fmt.Println("All contracts ", regClient.Contracts().Addresses)

	willShutdown, err := regClient.WillNewScannerShutdownPool(big.NewInt(poolID))
	if err != nil {
		return fmt.Errorf("failed to check pool shutdown condition: %v", err)
	}
	if willShutdown && !force {
		redBold("Registering this scanner will shutdown the pool! Please stake more on the pool (id = %d) first.\n", poolID)
		return nil
	}

	ts := time.Now().Unix()
	regInfo, err := regClient.GenerateScannerRegistrationSignature(&eip712.ScannerNodeRegistration{
		Scanner:       scannerKey.Address,
		ScannerPoolId: big.NewInt(poolID),
		ChainId:       big.NewInt(int64(cfg.ChainID)),
		Metadata:      "",
		Timestamp:     big.NewInt(ts),
	})

	fmt.Println("###")
	fmt.Println("Scanner: ", scannerKey.Address.Hex())
	fmt.Println("ScannerPoolId: ", big.NewInt(poolID))
	fmt.Println("ChainId: ", big.NewInt(int64(cfg.ChainID)))
	fmt.Println("Timestamp: ", big.NewInt(ts))
	fmt.Println("###")

	if err != nil {
		return fmt.Errorf("failed to generate registration signature: %v", err)
	}

	infoB, err := json.Marshal(regInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal registration info: %v", err)
	}

	infoStr := base64.URLEncoding.EncodeToString(infoB)

	if clean {
		fmt.Println(infoStr)
		return nil
	}

	if polygonscan {
		whiteBold("Please use the registerScannerNode() inputs below on https://polygonscan.com as soon as possible and do not share with anyone!\n\n")
		color.New(color.FgYellow).Println("req      :", makeArgsTuple(scannerKey.Address.Hex(), poolID, cfg.ChainID, ts))
		color.New(color.FgYellow).Println("signature:", regInfo.Signature)
	} else {
		whiteBold("Please use the registration signature below on https://app.zktoro.network as soon as possible and do not share with anyone!\n\n")
		color.New(color.FgYellow).Println(infoStr)
	}

	return nil
}

//	struct ScannerNodeRegistration {
//		address scanner;
//		uint256 scannerPoolId;
//		uint256 chainId;
//		string metadata;
//		uint256 timestamp;
//	}
func makeArgsTuple(scannerAddr string, poolID int64, chainID int, ts int64) string {
	tuple := make([]string, 5)
	tuple[0] = scannerAddr
	tuple[1] = (*hexutil.Big)(big.NewInt(poolID)).String()
	tuple[2] = (*hexutil.Big)(big.NewInt(int64(chainID))).String()
	tuple[4] = (*hexutil.Big)(big.NewInt(int64(ts))).String()
	b, _ := json.Marshal(tuple)
	return string(b)
}
