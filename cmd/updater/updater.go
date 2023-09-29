package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"path"
	"time"

	"zktoro/zktoro-core-go/clients/health"
	"zktoro/zktoro-core-go/utils"

	"zktoro/config"
	"zktoro/healthutils"

	"zktoro/services"

	"zktoro/services/updater"

	"zktoro/store"

	log "github.com/sirupsen/logrus"
)

const minUpdateInterval = 1 * time.Minute
const maxUpdateInterval = 24 * time.Hour

func generateIntervalMs(addr string) int64 {
	interval := big.NewInt(0)
	interval.Mod(utils.ScannerIDHexToBigInt(addr), big.NewInt((maxUpdateInterval).Milliseconds()))
	return interval.Int64() + minUpdateInterval.Milliseconds()
}

type keyAddress struct {
	Address string `json:"address"`
}

func loadAddressFromKeyFile() (string, error) {
	files, err := ioutil.ReadDir(config.DefaultContainerKeyDirPath)
	if err != nil {
		return "", err
	}

	if len(files) != 1 {
		return "", errors.New("there must be only one key in key directory")
	}

	b, err := ioutil.ReadFile(path.Join(config.DefaultContainerKeyDirPath, files[0].Name()))
	if err != nil {
		return "", err
	}

	var addr keyAddress
	if err := json.Unmarshal(b, &addr); err != nil {
		return "", err
	}

	return fmt.Sprintf("0x%s", addr.Address), nil
}

func initServices(ctx context.Context, cfg config.Config) ([]services.Service, error) {
	// cfg.Registry.JsonRpc.Url = utils.ConvertToDockerHostURL(cfg.Registry.JsonRpc.Url)
	log.WithField("Keep Original URL insterad of Docker Host", cfg.Registry.JsonRpc.Url)
	cfg.Registry.IPFS.APIURL = utils.ConvertToDockerHostURL(cfg.Registry.IPFS.APIURL)
	cfg.Registry.IPFS.GatewayURL = utils.ConvertToDockerHostURL(cfg.Registry.IPFS.GatewayURL)

	log.WithFields(log.Fields{
		"developmentMode": utils.ParseBoolEnvVar(config.EnvDevelopment),
	}).Info("updater modes")

	address, err := loadAddressFromKeyFile()
	if err != nil {
		return nil, err
	}

	intervalMs := generateIntervalMs(address)
	updateDelay := int(intervalMs / 1000)
	if cfg.AutoUpdate.UpdateDelay != nil {
		updateDelay = *cfg.AutoUpdate.UpdateDelay
	}
	var srs store.ScannerReleaseStore

	if config.WithoutENS {
		srs, err = store.NewScannerReleaseStoreWithoutENS(ctx, cfg)
	} else {
		srs, err = store.NewScannerReleaseStore(ctx, cfg)
	}

	if err != nil {
		return nil, err
	}

	updaterService := updater.NewUpdaterService(
		ctx, srs, config.DefaultContainerPort, updateDelay, cfg.AutoUpdate.CheckIntervalSeconds,
	)

	return []services.Service{
		health.NewService(
			ctx, "", healthutils.DefaultHealthServerErrHandler,
			health.CheckerFrom(summarizeReports, updaterService),
		),
		updaterService,
	}, nil
}

func summarizeReports(reports health.Reports) *health.Report {
	summary := health.NewSummary()

	checkedErr, ok := reports.NameContains("event.checked.error")
	if !ok {
		summary.Fail()
		return summary.Finish()
	}
	if len(checkedErr.Details) > 0 {
		summary.Addf("auto-updater is failing to check new versions with error '%s'", checkedErr.Details)
		summary.Status(health.StatusFailing)
	}

	checkedTime, ok := reports.NameContains("event.checked.time")
	if ok {
		t, ok := checkedTime.Time()
		if ok {
			checkDelay := time.Since(*t)
			if checkDelay > time.Minute*10 {
				summary.Addf("and late for %d minutes", int64(checkDelay.Minutes()))
				summary.Status(health.StatusFailing)
			}
		}
	}
	summary.Punc(".")
	return summary.Finish()
}

func Run() {
	services.ContainerMain("updater", initServices)
}
