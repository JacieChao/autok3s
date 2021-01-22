package amazone

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cnrancher/autok3s/pkg/cluster"
	"github.com/cnrancher/autok3s/pkg/types"
	"github.com/cnrancher/autok3s/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const createUsageExample = `  autok3s -d create \
    --provider amazone \
    --name <cluster name> \
    --access-key <access-key> \
    --secret-key <access-secret> \
    --master 1
`

func (p *Amazone) GetUsageExample(action string) string {
	switch action {
	case "create":
		return createUsageExample
	default:
		return ""
	}
}

func (p *Amazone) GetOptionFlags() []types.Flag {
	return p.sharedFlags()
}

func (p *Amazone) GetStartFlags(cmd *cobra.Command) *pflag.FlagSet {
	return nil
}

func (p *Amazone) GetStopFlags(cmd *cobra.Command) *pflag.FlagSet {
	return nil
}

func (p *Amazone) GetDeleteFlags(cmd *cobra.Command) *pflag.FlagSet {
	fs := []types.Flag{
		{
			Name:      "name",
			P:         &p.Name,
			V:         p.Name,
			Usage:     "Set the name of the kubeconfig context",
			ShortHand: "n",
			Required:  true,
		},
		{
			Name:   "region",
			P:      &p.Region,
			V:      p.Region,
			Usage:  "aws region",
			EnvVar: "AWS_DEFAULT_REGION",
		},
	}

	return utils.ConvertFlags(cmd, fs)
}

func (p *Amazone) GetJoinFlags(cmd *cobra.Command) *pflag.FlagSet {
	fs := p.sharedFlags()
	return utils.ConvertFlags(cmd, fs)
}

func (p *Amazone) GetSSHFlags(cmd *cobra.Command) *pflag.FlagSet {
	fs := []types.Flag{
		{
			Name:      "name",
			P:         &p.Name,
			V:         p.Name,
			Usage:     "Set the name of the kubeconfig context",
			ShortHand: "n",
			Required:  true,
		},
		{
			Name:   "region",
			P:      &p.Region,
			V:      p.Region,
			Usage:  "aws region",
			EnvVar: "AWS_DEFAULT_REGION",
		},
	}

	return utils.ConvertFlags(cmd, fs)
}

func (p *Amazone) GetCredentialFlags() []types.Flag {
	fs := []types.Flag{
		{
			Name:     "access-key",
			P:        &p.AccessKey,
			V:        p.AccessKey,
			Usage:    "AWS access key",
			Required: true,
			EnvVar:   "AWS_ACCESS_KEY_ID",
		},
		{
			Name:     "secret-key",
			P:        &p.SecretKey,
			V:        p.SecretKey,
			Usage:    "AWS secret key",
			Required: true,
			EnvVar:   "AWS_SECRET_ACCESS_KEY",
		},
	}

	return fs
}

func (p *Amazone) GetSSHConfig() *types.SSH {
	ssh := &types.SSH{
		User: defaultUser,
		Port: "22",
	}
	return ssh
}

func (p *Amazone) BindCredentialFlags() *pflag.FlagSet {
	nfs := pflag.NewFlagSet("", pflag.ContinueOnError)
	nfs.StringVar(&p.AccessKey, "access-key", p.AccessKey, "AWS access key")
	nfs.StringVar(&p.SecretKey, "secret-key", p.SecretKey, "AWS secret key")
	return nfs
}

func (p *Amazone) MergeClusterOptions() error {
	clusters, err := cluster.ReadFromState(&types.Cluster{
		Metadata: p.Metadata,
		Options:  p.Options,
	})
	if err != nil {
		return err
	}

	var matched *types.Cluster
	for _, c := range clusters {
		if c.Provider == p.Provider && c.Name == fmt.Sprintf("%s.%s.%s", p.Name, p.Region, p.Provider) {
			matched = &c
		}
	}

	if matched != nil {
		p.overwriteMetadata(matched)
		// delete command need merge status value.
		p.mergeOptions(*matched)
	}

	return nil
}

func (p *Amazone) mergeOptions(input types.Cluster) {
	source := reflect.ValueOf(&p.Options).Elem()
	target := reflect.Indirect(reflect.ValueOf(&input.Options)).Elem()

	p.mergeValues(source, target)
}

func (p *Amazone) mergeValues(source, target reflect.Value) {
	for i := 0; i < source.NumField(); i++ {
		for _, k := range target.MapKeys() {
			if strings.Contains(source.Type().Field(i).Tag.Get("yaml"), k.String()) {
				if source.Field(i).Kind().String() == "struct" {
					p.mergeValues(source.Field(i), target.MapIndex(k).Elem())
				} else {
					source.Field(i).SetString(fmt.Sprintf("%s", target.MapIndex(k)))
				}
			}
		}
	}
}

func (p *Amazone) overwriteMetadata(matched *types.Cluster) {
	// doesn't need to be overwrite.
	p.Status = matched.Status
	p.Token = matched.Token
	p.IP = matched.IP
	p.UI = matched.UI
	p.CloudControllerManager = matched.CloudControllerManager
	p.ClusterCIDR = matched.ClusterCIDR
	p.DataStore = matched.DataStore
	p.Mirror = matched.Mirror
	p.DockerMirror = matched.DockerMirror
	p.InstallScript = matched.InstallScript
	p.Network = matched.Network
	// needed to be overwrite.
	if p.K3sChannel == "" {
		p.K3sChannel = matched.K3sChannel
	}
	if p.K3sVersion == "" {
		p.K3sVersion = matched.K3sVersion
	}
	if p.InstallScript == "" {
		p.InstallScript = matched.InstallScript
	}
	if p.Registry == "" {
		p.Registry = matched.Registry
	}
	if p.MasterExtraArgs == "" {
		p.MasterExtraArgs = matched.MasterExtraArgs
	}
	if p.WorkerExtraArgs == "" {
		p.WorkerExtraArgs = matched.WorkerExtraArgs
	}
}

func (p *Amazone) sharedFlags() []types.Flag {
	fs := []types.Flag{
		{
			Name:   "region",
			P:      &p.Region,
			V:      p.Region,
			Usage:  "aws region",
			EnvVar: "AWS_DEFAULT_REGION",
		},
		{
			Name:   "zone",
			P:      &p.Zone,
			V:      p.Zone,
			Usage:  "Zone is physical areas with independent power grids and networks within one region. e.g.(a,b,c,d,e)",
			EnvVar: "AWS_ZONE",
		},
		{
			Name:   "keypair-name",
			P:      &p.KeypairName,
			V:      p.KeypairName,
			Usage:  "AWS keypair to use connect to instance",
			EnvVar: "AWS_KEYPAIR_NAME",
		},
		{
			Name:   "ami",
			P:      &p.AMI,
			V:      p.AMI,
			Usage:  "Used to specify the image to be used by the instance",
			EnvVar: "AWS_AMI",
		},
		{
			Name:   "instance-type",
			P:      &p.InstanceType,
			V:      p.InstanceType,
			Usage:  "Used to specify the type to be used by the instance",
			EnvVar: "AWS_INSTANCE_TYPE",
		},
		{
			Name:   "subnet-id",
			P:      &p.SubnetId,
			V:      p.SubnetId,
			Usage:  "AWS VPC subnet id",
			EnvVar: "AWS_SUBNET_ID",
		},
		{
			Name:   "volume-type",
			P:      &p.VolumeType,
			V:      p.VolumeType,
			Usage:  "Used to specify the EBS volume type",
			EnvVar: "AWS_VOLUME_TYPE",
		},
		{
			Name:   "root-size",
			P:      &p.RootSize,
			V:      p.RootSize,
			Usage:  "Used to specify the root disk size used by the instance (in GB)",
			EnvVar: "AWS_ROOT_SIZE",
		},
		{
			Name:   "security-group",
			P:      &p.SecurityGroup,
			V:      p.SecurityGroup,
			Usage:  "Used to specify the security group used by the instance",
			EnvVar: "AWS_SECURITY_GROUP",
		},
		{
			Name:  "master",
			P:     &p.Master,
			V:     p.Master,
			Usage: "Number of master node",
		},
		{
			Name:  "worker",
			P:     &p.Worker,
			V:     p.Worker,
			Usage: "Number of worker node",
		},
	}

	return fs
}
