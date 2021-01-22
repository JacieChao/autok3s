package cluster

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/spf13/pflag"

	"github.com/cnrancher/autok3s/pkg/providers"

	"github.com/cnrancher/autok3s/pkg/common"
	"github.com/cnrancher/autok3s/pkg/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/syncmap"
)

const (
	k3sVersion             = ""
	k3sChannel             = "stable"
	k3sInstallScript       = "https://get.k3s.io"
	master                 = "0"
	worker                 = "0"
	ui                     = false
	cloudControllerManager = false
	haMode                 = false
	defaultCidr            = "10.42.0.0/16"
)

type ProviderBase struct {
	*types.Metadata
	types.Status
	Options interface{}
	p       providers.Provider
	m       *sync.Map
	logger  *logrus.Logger
}

func NewBaseProvider() *ProviderBase {
	return &ProviderBase{
		Metadata: &types.Metadata{
			Master:                 master,
			Worker:                 worker,
			UI:                     ui,
			CloudControllerManager: cloudControllerManager,
			K3sVersion:             k3sVersion,
			K3sChannel:             k3sChannel,
			InstallScript:          k3sInstallScript,
			Cluster:                haMode,
		},
		Status: types.Status{
			MasterNodes: make([]types.Node, 0),
			WorkerNodes: make([]types.Node, 0),
		},
		m:      new(syncmap.Map),
		logger: common.NewLogger(common.Debug),
	}
}

func (b *ProviderBase) GetLogger() *logrus.Logger {
	return b.logger
}

func (b *ProviderBase) GenerateProvider(name string) error {
	p, err := providers.GetProvider(name)
	if err != nil {
		return err
	}
	config, err := json.Marshal(b)
	if err != nil {
		return err
	}
	err = p.SetConfig(config)
	if err != nil {
		return err
	}
	b.p = p
	return nil
}

func (b *ProviderBase) GenerateProviderFlags() []types.Flag {
	fs := b.p.GetCredentialFlags()
	fs = append(fs, b.p.GetOptionFlags()...)
	return fs
}

func (b *ProviderBase) GetSSHConfig() *types.SSH {
	return b.p.GetSSHConfig()
}

func (b *ProviderBase) BindCredentialFlags() *pflag.FlagSet {
	return b.p.BindCredentialFlags()
}

func (b *ProviderBase) GetUsageExample(action string) string {
	return b.p.GetUsageExample(action)
}

func (b *ProviderBase) GetProviderName() string {
	return b.Provider
}

func (b *ProviderBase) MergeClusterOptions() error {

	return nil
}

func (b *ProviderBase) GenerateClusterName() {
	if b.p.GenerateClusterName() != "" {
		b.Name = b.p.GenerateClusterName()
	}
}

func (b *ProviderBase) CreateCheck(ssh *types.SSH) error {
	return b.p.CreateCheck(ssh)
}

func (b *ProviderBase) CreateK3sCluster(ssh *types.SSH) (err error) {
	b.logger.Infof("[%s] prepare for k3s cluster...\n", b.GetProviderName())

	defer func() {
		if err == nil && len(b.Status.MasterNodes) > 0 {
			fmt.Printf(common.UsageInfo, b.Name)
			if b.UI {
				if b.CloudControllerManager {
					fmt.Printf("\nK3s UI URL: https://<using `kubectl get svc -A` get UI address>:8999\n")
				} else {
					fmt.Printf("\nK3s UI URL: https://%s:8999\n", b.Status.MasterNodes[0].PublicIPAddress[0])
				}
			}
			fmt.Println("")
		}
	}()

	// prepare instance for each custom provider
	pro, err := providers.GetProvider(b.GetProviderName())
	if err != nil {
		return err
	}

	c, err := pro.PrepareCluster(ssh)
	if err != nil {
		return err
	}

	// deploy k3s cluster
	if err = InitK3sCluster(c); err != nil {
		return err
	}

	b.logger.Infof("[%s] successfully executed create logic\n", b.GetProviderName())
	return nil
}
