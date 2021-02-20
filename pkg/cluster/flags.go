package cluster

import (
	"encoding/json"

	"github.com/cnrancher/autok3s/pkg/types"
	"github.com/spf13/pflag"
)

func (b *ProviderBase) GetClusterOptions() []types.Flag {
	return []types.Flag{
		{
			Name:      "provider",
			P:         &b.Provider,
			V:         b.Provider,
			Usage:     "Provider is a module which provides an interface for managing cloud resources. (e.g. alibaba)",
			ShortHand: "p",
			EnvVar:    "AUTOK3S_PROVIDER",
			Required:  true,
		},
		{
			Name:      "name",
			P:         &b.Name,
			V:         b.Name,
			Usage:     "Set the name of the kubeconfig context",
			ShortHand: "n",
			Required:  true,
		},
		{
			Name:  "ip",
			P:     &b.IP,
			V:     b.IP,
			Usage: "Public IP of an existing k3s server",
		},
		{
			Name:  "k3s-version",
			P:     &b.K3sVersion,
			V:     b.K3sVersion,
			Usage: "Used to specify the version of k3s cluster, overrides k3s-channel",
		},
		{
			Name:  "k3s-channel",
			P:     &b.K3sChannel,
			V:     b.K3sChannel,
			Usage: "Used to specify the release channel of k3s. e.g.(stable, latest, or i.e. v1.18)",
		},
		{
			Name:  "k3s-install-script",
			P:     &b.InstallScript,
			V:     b.InstallScript,
			Usage: "Change the default upstream k3s install script address",
		},
		{
			Name:  "cloud-controller-manager",
			P:     &b.CloudControllerManager,
			V:     b.CloudControllerManager,
			Usage: "Enable cloud-controller-manager component",
		},
		{
			Name:  "master-extra-args",
			P:     &b.MasterExtraArgs,
			V:     b.MasterExtraArgs,
			Usage: "Master extra arguments for k3s installer, wrapped in quotes. e.g.(--master-extra-args '--no-deploy metrics-server')",
		},
		{
			Name:  "worker-extra-args",
			P:     &b.WorkerExtraArgs,
			V:     b.WorkerExtraArgs,
			Usage: "Worker extra arguments for k3s installer, wrapped in quotes. e.g.(--worker-extra-args '--node-taint key=value:NoExecute')",
		},
		{
			Name:  "registry",
			P:     &b.Registry,
			V:     b.Registry,
			Usage: "K3s registry file, see: https://rancher.com/docs/k3s/latest/en/installation/private-registry",
		},
		{
			Name:  "datastore",
			P:     &b.DataStore,
			V:     b.DataStore,
			Usage: "K3s datastore, HA mode `create/join` master node needed this flag",
		},
		{
			Name:  "token",
			P:     &b.Token,
			V:     b.Token,
			Usage: "K3s master token, if empty will automatically generated",
		},
		{
			Name:  "ui",
			P:     &b.UI,
			V:     b.UI,
			Usage: "Enable K3s UI.",
		},
		{
			Name:  "cluster",
			P:     &b.Cluster,
			V:     b.Cluster,
			Usage: "Form k3s cluster using embedded etcd (requires K8s >= 1.19)",
		},
	}
}

func (b *ProviderBase) GetSSHFlags(cSSH *types.SSH) []types.Flag {
	b.SSH = cSSH
	return []types.Flag{
		{
			Name:  "ssh-user",
			P:     &b.User,
			V:     b.User,
			Usage: "SSH user for host",
		},
		{
			Name:  "ssh-port",
			P:     &b.Port,
			V:     b.Port,
			Usage: "SSH port for host",
		},
		{
			Name:  "ssh-key-path",
			P:     &b.SSHKeyPath,
			V:     b.SSHKeyPath,
			Usage: "SSH private key path",
		},
		{
			Name:  "ssh-key-pass",
			P:     &b.SSHKeyPassphrase,
			V:     b.SSHKeyPassphrase,
			Usage: "SSH passphrase of private key",
		},
		{
			Name:  "ssh-key-cert-path",
			P:     &b.SSHCertPath,
			V:     b.SSHCertPath,
			Usage: "SSH private key certificate path",
		},
		{
			Name:  "ssh-password",
			P:     &b.Password,
			V:     b.Password,
			Usage: "SSH login password",
		},
		{
			Name:  "ssh-agent",
			P:     &b.SSHAgentAuth,
			V:     b.SSHAgentAuth,
			Usage: "Enable ssh agent",
		},
	}
}

func (b *ProviderBase) GenerateProviderFlags() []types.Flag {
	fs := b.p.GetCredentialFlags()
	fs = append(fs, b.p.GetOptionFlags()...)
	return fs
}

func (b *ProviderBase) BindCredentialFlags() *pflag.FlagSet {
	return b.p.BindCredentialFlags()
}

func (b *ProviderBase) GetUsageExample(action string) string {
	return b.p.GetUsageExample(action)
}

func (b *ProviderBase) GetDeleteFlags() []types.Flag {
	return []types.Flag{
		{
			Name:      "name",
			P:         &b.Name,
			V:         b.Name,
			Usage:     "Set the name of the kubeconfig context",
			ShortHand: "n",
			Required:  true,
		},
		{
			Name:      "provider",
			P:         &b.Provider,
			V:         b.Provider,
			Usage:     "Provider is a module which provides an interface for managing cloud resources. (e.g. alibaba)",
			ShortHand: "p",
			EnvVar:    "AUTOK3S_PROVIDER",
			Required:  true,
		},
	}
}

func (b *ProviderBase) GetProviderDeleteFlags() []types.Flag {
	fs := b.p.GetCredentialFlags()
	fs = append(fs, b.p.GetDeleteFlags()...)
	return fs
}

func (b *ProviderBase) MergeClusterOptions() error {
	state, err := GetClusterState(b.Name, b.Provider)
	if err != nil {
		return err
	}
	b.overwriteMetadata(state)
	b.Status = types.Status{
		Status: state.Status,
	}
	masterNodes := []types.Node{}
	err = json.Unmarshal(state.MasterNodes, &masterNodes)
	if err != nil {
		return err
	}
	workerNodes := []types.Node{}
	err = json.Unmarshal(state.WorkerNodes, &workerNodes)
	b.Status.MasterNodes = masterNodes
	b.Status.WorkerNodes = workerNodes
	cluster := &types.Cluster{
		Metadata: *b.Metadata,
		Status:   b.Status,
	}
	b.p.MergeClusterOptions(cluster)
	if err = b.p.SetOptions(state.Options); err != nil {
		return err
	}
	b.Options = b.p.GetOptions()
	return nil
}

func (b *ProviderBase) overwriteMetadata(matched *types.ClusterState) {
	// doesn't need to be overwrite.
	b.Token = matched.Token
	b.IP = matched.IP
	b.UI = matched.UI
	b.CloudControllerManager = matched.CloudControllerManager
	b.ClusterCidr = matched.ClusterCidr
	b.DataStore = matched.DataStore
	b.Mirror = matched.Mirror
	b.DockerMirror = matched.DockerMirror
	b.InstallScript = matched.InstallScript
	b.Network = matched.Network
	// needed to be overwrite.
	if b.K3sChannel == "" {
		b.K3sChannel = matched.K3sChannel
	}
	if b.K3sVersion == "" {
		b.K3sVersion = matched.K3sVersion
	}
	if b.InstallScript == "" {
		b.InstallScript = matched.InstallScript
	}
	if b.Registry == "" {
		b.Registry = matched.Registry
	}
	if b.MasterExtraArgs == "" {
		b.MasterExtraArgs = matched.MasterExtraArgs
	}
	if b.WorkerExtraArgs == "" {
		b.WorkerExtraArgs = matched.WorkerExtraArgs
	}
}
