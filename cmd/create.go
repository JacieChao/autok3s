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
	example = `  autok3s -d create \
    --provider amazone \
    --name <cluster name> \
    --access-key <access-key> \
    --secret-key <access-secret> \
    --master 1
`
	createCmd = &cobra.Command{
		Use:     "create",
		Short:   "Create k3s cluster",
		Example: example,
	}

	//cp        providers.Provider
	//cProvider string

	cSSH = &types.SSH{
		Port: "22",
	}
)

func CreateCommand() *cobra.Command {
	b := cluster.NewBaseProvider()
	createCmd.Flags().AddFlagSet(utils.ConvertFlags(createCmd, common.GetClusterOptions(b.Metadata)))
	createCmd.Flags().AddFlagSet(utils.ConvertFlags(createCmd, common.GetSSHConfig(cSSH)))

	// load dynamic provider flags.
	cProvider := common.FlagHackLookup("--provider")
	if cProvider != "" {
		err := b.GenerateProvider(cProvider)
		if err != nil {
			logrus.Fatalln(err)
		}

		cSSH = b.GetSSHConfig()
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
		// generate cluster name. e.g. input: "--name k3s1 --region cn-hangzhou" output: "k3s1.cn-hangzhou.<provider>"
		//cp.GenerateClusterName()
		if err := b.CreateCheck(cSSH); err != nil {
			logrus.Fatalln(err)
		}

		// create k3s cluster with generated cluster name.
		//if err := cp.CreateK3sCluster(cSSH); err != nil {
		//	logrus.Errorln(err)
		//	if rErr := cp.Rollback(); rErr != nil {
		//		logrus.Fatalln(rErr)
		//	}
		//}
	}

	return createCmd
}
