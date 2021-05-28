package native

import (
	"github.com/cnrancher/autok3s/pkg/types"
)

const createUsageExample = `  autok3s -d create \
    --provider native \
    --ssh-key-path <ssh-key-path> \
    --master-ips <master-ips> \
    --worker-ips <worker-ips>
`

const joinUsageExample = `  autok3s -d join \
    --provider native \
    --ssh-key-path <ssh-key-path> \
    --ip <existing server ip> \
    --master-ips <master-ips> \
    --worker-ips <worker-ips>
`

// GetUsageExample returns cli usage example for native provider
func (p *Native) GetUsageExample(action string) string {
	switch action {
	case "create":
		return createUsageExample
	case "join":
		return joinUsageExample
	default:
		return "not support"
	}
}

// GetCreateFlags return create flags for native provider
func (p *Native) GetCreateFlags() []types.Flag {
	cSSH := p.GetSSHConfig()
	p.SSH = *cSSH
	fs := p.GetClusterOptions()
	fs = append(fs, p.GetCreateOptions()...)
	return fs
}

// GetOptionFlags return flags for option
func (p *Native) GetOptionFlags() []types.Flag {
	return p.sharedFlags()
}

// GetJoinFlags return join flags for native provider
func (p *Native) GetJoinFlags() []types.Flag {
	fs := p.sharedFlags()
	fs = append(fs, p.GetClusterOptions()...)
	fs = append(fs, types.Flag{
		Name:  "ip",
		P:     &p.IP,
		V:     p.IP,
		Usage: "IP for an existing k3s server",
	})
	return fs
}

// GetSSHFlags not supported
func (p *Native) GetSSHFlags() []types.Flag {
	return []types.Flag{}
}

// GetDeleteFlags not supported
func (p *Native) GetDeleteFlags() []types.Flag {
	return []types.Flag{}
}

// GetCredentialFlags not supported
func (p *Native) GetCredentialFlags() []types.Flag {
	return []types.Flag{}
}

// GetSSHConfig return default ssh config for native provider
func (p *Native) GetSSHConfig() *types.SSH {
	ssh := &types.SSH{
		SSHUser:    defaultUser,
		SSHPort:    "22",
		SSHKeyPath: defaultSSHKeyPath,
	}
	return ssh
}

// BindCredential not supported
func (p *Native) BindCredential() error {
	return nil
}

// MergeClusterOptions not supported
func (p *Native) MergeClusterOptions() error {
	return nil
}

func (p *Native) sharedFlags() []types.Flag {
	fs := []types.Flag{
		{
			Name:  "master-ips",
			P:     &p.MasterIps,
			V:     p.MasterIps,
			Usage: "Public IPs of master nodes on which to install agent, multiple IPs are separated by commas",
		},
		{
			Name:  "worker-ips",
			P:     &p.WorkerIps,
			V:     p.WorkerIps,
			Usage: "Public IPs of worker nodes on which to install agent, multiple IPs are separated by commas",
		},
	}

	return fs
}
