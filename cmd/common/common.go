package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/cnrancher/autok3s/pkg/cluster"

	"github.com/cnrancher/autok3s/pkg/providers/amazone"

	"github.com/cnrancher/autok3s/pkg/common"
	"github.com/cnrancher/autok3s/pkg/providers"
	"github.com/cnrancher/autok3s/pkg/providers/alibaba"
	"github.com/cnrancher/autok3s/pkg/providers/tencent"
	"github.com/cnrancher/autok3s/pkg/types"
	typesAli "github.com/cnrancher/autok3s/pkg/types/alibaba"
	typesAmazone "github.com/cnrancher/autok3s/pkg/types/amazone"
	typesTencent "github.com/cnrancher/autok3s/pkg/types/tencent"
	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func bindPFlags(cmd *cobra.Command, p *cluster.ProviderBase) {
	name, err := cmd.Flags().GetString("provider")
	if err != nil {
		logrus.Fatalln(err)
	}

	cmd.Flags().Visit(func(f *pflag.Flag) {
		if IsCredentialFlag(f.Name, p.BindCredentialFlags()) {
			if err := viper.BindPFlag(fmt.Sprintf(common.BindPrefix, name, f.Name), f); err != nil {
				logrus.Fatalln(err)
			}
		}
	})
}

func InitPFlags(cmd *cobra.Command, p *cluster.ProviderBase) {
	// bind env to flags
	bindEnvFlags(cmd)
	bindPFlags(cmd, p)

	// read options from config.
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalln(err)
	}

	// sync config data to local cfg path.
	if err := viper.WriteConfig(); err != nil {
		logrus.Fatalln(err)
	}
}

func bindEnvFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		envAnnotation := f.Annotations[utils.BashCompEnvVarFlag]
		if len(envAnnotation) == 0 {
			return
		}

		if os.Getenv(envAnnotation[0]) != "" {
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", os.Getenv(envAnnotation[0])))
		}
	})
}

// Borrowed from https://github.com/docker/machine/blob/master/commands/create.go#L267.
func FlagHackLookup(flagName string) string {
	// e.g. "-d" for "--driver"
	flagPrefix := flagName[1:3]

	// TODO: Should we support -flag-name (single hyphen) syntax as well?
	for i, arg := range os.Args {
		if strings.Contains(arg, flagPrefix) {
			// format '--driver foo' or '-d foo'
			if arg == flagPrefix || arg == flagName {
				if i+1 < len(os.Args) {
					return os.Args[i+1]
				}
			}

			// format '--driver=foo' or '-d=foo'
			if strings.HasPrefix(arg, flagPrefix+"=") || strings.HasPrefix(arg, flagName+"=") {
				return strings.Split(arg, "=")[1]
			}
		}
	}

	return ""
}

func IsCredentialFlag(s string, nfs *pflag.FlagSet) bool {
	found := false
	nfs.VisitAll(func(f *pflag.Flag) {
		if strings.EqualFold(s, f.Name) {
			found = true
		}
	})
	return found
}

func MakeSureCredentialFlag(flags *pflag.FlagSet, p *cluster.ProviderBase) error {
	flags.VisitAll(func(flag *pflag.Flag) {
		// if viper has set the value, make sure flag has the value set to pass require check
		if IsCredentialFlag(flag.Name, p.BindCredentialFlags()) && viper.IsSet(fmt.Sprintf(common.BindPrefix, p.GetProviderName(), flag.Name)) {
			flags.Set(flag.Name, viper.GetString(fmt.Sprintf(common.BindPrefix, p.GetProviderName(), flag.Name)))
		}
	})

	return nil
}

func GetProviderByState(c types.Cluster) (providers.Provider, error) {
	b, err := yaml.Marshal(c.Options)
	if err != nil {
		return nil, err
	}
	switch c.Provider {
	case "alibaba":
		option := &typesAli.Options{}
		if err := yaml.Unmarshal(b, option); err != nil {
			return nil, err
		}
		return &alibaba.Alibaba{
			Metadata: c.Metadata,
			Options:  *option,
			Status:   c.Status,
		}, nil
	case "tencent":
		option := &typesTencent.Options{}
		if err := yaml.Unmarshal(b, option); err != nil {
			return nil, err
		}
		return &tencent.Tencent{
			Metadata: c.Metadata,
			Options:  *option,
			Status:   c.Status,
		}, nil
	case "amazone":
		option := &typesAmazone.Options{}
		if err := yaml.Unmarshal(b, option); err != nil {
			return nil, err
		}
		return &amazone.Amazone{
			Metadata: c.Metadata,
			Options:  *option,
			Status:   c.Status,
		}, nil
	default:
		return nil, fmt.Errorf("invalid provider name %s", c.Provider)
	}
}

func GetClusterOptions(opt *types.Metadata) []types.Flag {
	return []types.Flag{
		{
			Name:      "provider",
			P:         &opt.Provider,
			V:         opt.Provider,
			Usage:     "Provider is a module which provides an interface for managing cloud resources. (e.g. amazone)",
			ShortHand: "p",
			EnvVar:    "AUTOK3S_PROVIDER",
			Required:  true,
		},
		{
			Name:      "name",
			P:         &opt.Name,
			V:         opt.Name,
			Usage:     "Set the name of the kubeconfig context",
			ShortHand: "n",
			Required:  true,
		},
		{
			Name:  "ip",
			P:     &opt.IP,
			V:     opt.IP,
			Usage: "Public IP of an existing k3s server",
		},
		{
			Name:  "k3s-version",
			P:     &opt.K3sVersion,
			V:     opt.K3sVersion,
			Usage: "Used to specify the version of k3s cluster, overrides k3s-channel",
		},
		{
			Name:  "k3s-channel",
			P:     &opt.K3sChannel,
			V:     opt.K3sChannel,
			Usage: "Used to specify the release channel of k3s. e.g.(stable, latest, or i.e. v1.18)",
		},
		{
			Name:  "k3s-install-script",
			P:     &opt.InstallScript,
			V:     opt.InstallScript,
			Usage: "Change the default upstream k3s install script address",
		},
		{
			Name:  "cloud-controller-manager",
			P:     &opt.CloudControllerManager,
			V:     opt.CloudControllerManager,
			Usage: "Enable cloud-controller-manager component",
		},
		{
			Name:  "master-extra-args",
			P:     &opt.MasterExtraArgs,
			V:     opt.MasterExtraArgs,
			Usage: "Master extra arguments for k3s installer, wrapped in quotes. e.g.(--master-extra-args '--no-deploy metrics-server')",
		},
		{
			Name:  "worker-extra-args",
			P:     &opt.WorkerExtraArgs,
			V:     opt.WorkerExtraArgs,
			Usage: "Worker extra arguments for k3s installer, wrapped in quotes. e.g.(--worker-extra-args '--node-taint key=value:NoExecute')",
		},
		{
			Name:  "registry",
			P:     &opt.Registry,
			V:     opt.Registry,
			Usage: "K3s registry file, see: https://rancher.com/docs/k3s/latest/en/installation/private-registry",
		},
		{
			Name:  "datastore",
			P:     &opt.DataStore,
			V:     opt.DataStore,
			Usage: "K3s datastore, HA mode `create/join` master node needed this flag",
		},
		{
			Name:  "token",
			P:     &opt.Token,
			V:     opt.Token,
			Usage: "K3s master token, if empty will automatically generated",
		},
		{
			Name:  "ui",
			P:     &opt.UI,
			V:     opt.UI,
			Usage: "Enable K3s UI.",
		},
		{
			Name:  "cluster",
			P:     &opt.Cluster,
			V:     opt.Cluster,
			Usage: "Form k3s cluster using embedded etcd (requires K8s >= 1.19)",
		},
		//{
		//	Name:  "master",
		//	P:     &opt.Master,
		//	V:     opt.Master,
		//	Usage: "Number of master node",
		//},
		//{
		//	Name:  "worker",
		//	P:     &opt.Worker,
		//	V:     opt.Worker,
		//	Usage: "Number of worker node",
		//},
	}
}

func GetSSHConfig(cSSH *types.SSH) []types.Flag {
	return []types.Flag{
		{
			Name:  "ssh-user",
			P:     &cSSH.User,
			V:     cSSH.User,
			Usage: "SSH user for host",
		},
		{
			Name:  "ssh-port",
			P:     &cSSH.Port,
			V:     cSSH.Port,
			Usage: "SSH port for host",
		},
		{
			Name:  "ssh-key-path",
			P:     &cSSH.SSHKeyPath,
			V:     cSSH.SSHKeyPath,
			Usage: "SSH private key path",
		},
		{
			Name:  "ssh-key-pass",
			P:     &cSSH.SSHKeyPassphrase,
			V:     cSSH.SSHKeyPassphrase,
			Usage: "SSH passphrase of private key",
		},
		{
			Name:  "ssh-key-cert-path",
			P:     &cSSH.SSHCertPath,
			V:     cSSH.SSHCertPath,
			Usage: "SSH private key certificate path",
		},
		{
			Name:  "ssh-password",
			P:     &cSSH.Password,
			V:     cSSH.Password,
			Usage: "SSH login password",
		},
		{
			Name:  "ssh-agent",
			P:     &cSSH.SSHAgentAuth,
			V:     &cSSH.SSHAgentAuth,
			Usage: "Enable ssh agent",
		},
	}
}
