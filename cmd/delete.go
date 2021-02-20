package cmd

import (
	"github.com/cnrancher/autok3s/cmd/common"
	"github.com/cnrancher/autok3s/pkg/cluster"
	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete k3s cluster",
	}
	force = false
)

func init() {
	deleteCmd.Flags().BoolVarP(&force, "force", "f", force, "Force delete cluster")
}

func DeleteCommand() *cobra.Command {
	b := cluster.NewBaseProvider()
	deleteCmd.Flags().AddFlagSet(utils.ConvertFlags(deleteCmd, b.GetDeleteFlags()))
	pStr := common.FlagHackLookup("--provider")

	if pStr != "" {
		err := b.GenerateProvider(pStr)
		if err != nil {
			logrus.Fatalln(err)
		}

		deleteCmd.Flags().AddFlagSet(utils.ConvertFlags(deleteCmd, b.GetProviderDeleteFlags()))
		deleteCmd.Example = b.GetUsageExample("delete")
	}

	deleteCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if pStr == "" {
			logrus.Fatalln("required flags(s) \"[provider]\" not set")
		}
		common.InitPFlags(cmd, b)
		err := b.MergeClusterOptions()
		if err != nil {
			return err
		}

		return common.MakeSureCredentialFlag(cmd.Flags(), b)
	}

	deleteCmd.Run = func(cmd *cobra.Command, args []string) {
		if err := b.DeleteK3sCluster(force); err != nil {
			logrus.Fatalln(err)
		}
	}

	return deleteCmd
}
