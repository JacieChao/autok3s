package cmd

import (
	"github.com/cnrancher/autok3s/cmd/common"
	"github.com/cnrancher/autok3s/pkg/cluster"
	"github.com/cnrancher/autok3s/pkg/types"
	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create k3s cluster",
	}
	cSSH = &types.SSH{
		Port: "22",
	}
)

func CreateCommand() *cobra.Command {
	b := cluster.NewBaseProvider()
	createCmd.Flags().AddFlagSet(utils.ConvertFlags(createCmd, b.GetClusterOptions()))
	// load dynamic provider flags.
	cProvider := common.FlagHackLookup("--provider")
	if cProvider != "" {
		err := b.GenerateProvider(cProvider)
		if err != nil {
			logrus.Fatalln(err)
		}

		cSSH = b.GetSSHConfig()
		createCmd.Flags().AddFlagSet(utils.ConvertFlags(createCmd, b.GetSSHFlags(cSSH)))
		//metadataConfig := b.GetMetadataConfig()
		//createCmd.Flags().AddFlagSet(utils.ConvertFlags(createCmd, b.GetClusterOptions(metadataConfig)))
		createCmd.Flags().AddFlagSet(utils.ConvertFlags(createCmd, b.GenerateProviderFlags()))
		createCmd.Example = b.GetUsageExample("create")
	}

	createCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if cProvider == "" {
			logrus.Fatalln("required flags(s) \"--provider\" not set")
		}
		common.InitPFlags(cmd, b)
		return common.MakeSureCredentialFlag(cmd.Flags(), b)
	}

	createCmd.Run = func(cmd *cobra.Command, args []string) {
		if err := b.CreateCheck(cSSH); err != nil {
			logrus.Fatalln(err)
		}

		if err := b.CreateK3sCluster(); err != nil {
			logrus.Errorln(err)
			if err = b.Rollback(); err != nil {
				logrus.Fatalln(err)
			}
		}
	}

	return createCmd
}
