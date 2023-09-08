package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"zktoro/config"

	"github.com/creasty/defaults"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	keyZktoroDir         = "zktoro_dir"
	keyZktoroPassphrase  = "zktoro_passphrase"
	keyZktoroDevelopment = "zktoro_development"
	keyZktoroExposeNats  = "zktoro_expose_nats"
)

var (
	cfg       config.Config
	cmdZktoro = &cobra.Command{
		Use:   "zktoro",
		Short: "Zktoro node command line interface",
		Long:  `Zktoro node host bot to execute trading strategies`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage: true,
	}

	cmdZktoroInit = &cobra.Command{
		Use:   "init",
		Short: "initialize a config file and a private key (doesn't overwrite)",
		RunE:  handlezktoroInit,
	}

	cmdzktoroAuthorize = &cobra.Command{
		Use:   "authorize",
		Short: "generate a signature for a specific action",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmdzktoroAuthorizePool = &cobra.Command{
		Use:   "pool",
		Short: "generate a pool registration signature",
		RunE:  withInitialized(withValidConfig(handlezktoroAuthorizePool)),
	}
)

func Execute() error {
	return cmdZktoro.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	cmdZktoro.AddCommand(cmdZktoroInit)

	cmdZktoro.PersistentFlags().String("passphrase", "", "passphrase to decrypt the private key (overrides $zktoro_PASSPHRASE)")
	viper.BindPFlag(keyZktoroPassphrase, cmdZktoro.PersistentFlags().Lookup("passphrase"))

	cmdZktoro.AddCommand(cmdzktoroAuthorize)
	cmdzktoroAuthorize.AddCommand(cmdzktoroAuthorizePool)

	// zktoro authorize pool
	cmdzktoroAuthorizePool.Flags().String("id", "", "scanner pool ID (integer)")
	cmdzktoroAuthorizePool.MarkFlagRequired("id")
	cmdzktoroAuthorizePool.Flags().Bool("polygonscan", false, "see the registerScannerNode() inputs to use in Polygonscan")
	cmdzktoroAuthorizePool.Flags().BoolP("force", "f", false, "ignore warning(s)")
	cmdzktoroAuthorizePool.Flags().Bool("clean", false, "output only the encoded registration info")
}

func initConfig() {
	viper.SetConfigType("yaml")

	viper.BindEnv(keyZktoroDir)
	viper.BindEnv(keyZktoroPassphrase)
	viper.BindEnv(keyZktoroDevelopment)
	viper.BindEnv(keyZktoroExposeNats)
	viper.AutomaticEnv()

	zktoroDir := viper.GetString(keyZktoroDir)
	if zktoroDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			logrus.Panicf("failed to get home dir: %v", err)
		}
		zktoroDir = path.Join(home, ".zktoro")
	}

	configPath := path.Join(zktoroDir, config.DefaultConfigFileName)
	configBytes, _ := ioutil.ReadFile(configPath)
	if err := yaml.Unmarshal(configBytes, &cfg); err != nil {
		yellowBold("Your config file is invalid! Please check the values and fix any formatting issues.\n")
		logrus.WithError(err).Fatal("failed to read config")
	}

	if err := defaults.Set(&cfg); err != nil {
		panic(err)
	}

	cfg.ZktoroDir = zktoroDir
	cfg.KeyDirPath = path.Join(cfg.ZktoroDir, config.DefaultKeysDirName)
	cfg.Development = viper.GetBool(keyZktoroDevelopment)
	cfg.Passphrase = viper.GetString(keyZktoroPassphrase)

	viper.ReadConfig(bytes.NewBuffer(configBytes))
	config.InitLogLevel(cfg)
}

func withValidConfig(handler func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := validateConfig(); err != nil {
			return err
		}
		return handler(cmd, args)
	}
}

func withInitialized(handler func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if !isInitialized() {
			yellowBold("Please make sure you do 'zktoro init' first and check your configuration at %s/config.yml\n", cfg.ZktoroDir)
			return errors.New("not initialized")
		}
		return handler(cmd, args)
	}
}

func validateConfig() error {
	validate := validator.New()

	// Use the YAML names while validating the struct.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	if err := validate.Struct(&cfg); err != nil {
		validationErrs := err.(validator.ValidationErrors)
		fmt.Fprintln(os.Stderr, "The config file has invalid or missing fields:")
		for _, validationErr := range validationErrs {
			fmt.Fprintf(os.Stderr, "  - %s\n", validationErr.Namespace()[7:])
		}
		return errors.New("invalid config file")
	}

	return nil
}
