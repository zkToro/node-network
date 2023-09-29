package nodecmd

import (
	inspector "zktoro/cmd/inspector"
	json_rpc "zktoro/cmd/json-rpc"
	jwt_provider "zktoro/cmd/jwt-provider"
	public_api "zktoro/cmd/public-api"
	"zktoro/cmd/publisher"
	"zktoro/cmd/scanner"
	"zktoro/cmd/storage"
	"zktoro/cmd/supervisor"
	"zktoro/cmd/updater"

	"github.com/spf13/cobra"
)

var (
	cmdzktoroNode = &cobra.Command{
		Use: "zktoro-node",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage: true,
	}

	cmdUpdater = &cobra.Command{
		Use: "updater",
		RunE: func(cmd *cobra.Command, args []string) error {
			updater.Run()
			return nil
		},
	}

	cmdSupervisor = &cobra.Command{
		Use: "supervisor",
		RunE: func(cmd *cobra.Command, args []string) error {
			supervisor.Run()
			return nil
		},
	}

	cmdScanner = &cobra.Command{
		Use: "scanner",
		RunE: func(cmd *cobra.Command, args []string) error {
			scanner.Run()
			return nil
		},
	}

	cmdJWTProvider = &cobra.Command{
		Use: "jwt-provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			jwt_provider.Run()
			return nil
		},
	}

	cmdPublisher = &cobra.Command{
		Use: "publisher",
		RunE: func(cmd *cobra.Command, args []string) error {
			publisher.Run()
			return nil
		},
	}

	cmdInspector = &cobra.Command{
		Use: "inspector",
		RunE: func(cmd *cobra.Command, args []string) error {
			inspector.Run()
			return nil
		},
	}

	cmdJsonRpc = &cobra.Command{
		Use: "json-rpc",
		RunE: func(cmd *cobra.Command, args []string) error {
			json_rpc.Run()
			return nil
		},
	}

	cmdPublicAPI = &cobra.Command{
		Use: "public-api",
		RunE: func(cmd *cobra.Command, args []string) error {
			public_api.Run()
			return nil
		},
	}

	cmdStorage = &cobra.Command{
		Use: "storage",
		RunE: func(cmd *cobra.Command, args []string) error {
			storage.Run()
			return nil
		},
	}
)

func init() {
	cmdzktoroNode.AddCommand(cmdUpdater)
	cmdzktoroNode.AddCommand(cmdSupervisor)
	cmdzktoroNode.AddCommand(cmdScanner)
	cmdzktoroNode.AddCommand(cmdPublisher)
	cmdzktoroNode.AddCommand(cmdInspector)
	cmdzktoroNode.AddCommand(cmdJsonRpc)
	cmdzktoroNode.AddCommand(cmdPublicAPI)
	cmdzktoroNode.AddCommand(cmdJWTProvider)
	cmdzktoroNode.AddCommand(cmdStorage)
}

func Run() error {
	return cmdzktoroNode.Execute()
}
