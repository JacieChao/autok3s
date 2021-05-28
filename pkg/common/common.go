package common

import (
	"path/filepath"
	"time"

	"github.com/cnrancher/autok3s/pkg/utils"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	KubeCfgFile        = ".kube/config"
	KubeCfgTempName    = "autok3s-temp-*"
	K3sManifestsDir    = "/var/lib/rancher/k3s/server/manifests"
	MasterInstanceName = "autok3s.%s.master"
	WorkerInstanceName = "autok3s.%s.worker"
	TagClusterPrefix   = "autok3s-"
	StatusRunning      = "Running"
	StatusCreating     = "Creating"
	StatusMissing      = "Missing"
	StatusFailed       = "Failed"
	StatusUpgrading    = "Upgrading"
	UsageInfoTitle     = "=========================== Prompt Info ==========================="
	UsageContext       = "Use 'autok3s kubectl config use-context %s'"
	UsagePods          = "Use 'autok3s kubectl get pods -A' get POD status`"
	DBFolder           = ".db"
	DBFile             = "autok3s.db"
)

var (
	Debug   = false
	CfgPath = utils.UserHome() + "/.autok3s"
	Backoff = wait.Backoff{
		Duration: 30 * time.Second,
		Factor:   1,
		Steps:    5,
	} // retry 5 times, total 120 seconds.
	DefaultDB *Store
)

// GetDefaultSSHKeyPath returns the default path for generated ssh file
func GetDefaultSSHKeyPath(clusterName, providerName string) string {
	return filepath.Join(CfgPath, providerName, "clusters", clusterName, "id_rsa")
}

// GetDefaultSSHPublicKeyPath returns the default path for generated ssh public key file
func GetDefaultSSHPublicKeyPath(clusterName, providerName string) string {
	return filepath.Join(CfgPath, providerName, "clusters", clusterName, "id_rsa.pub")
}

// GetClusterPath returns cluster file path for provider
func GetClusterPath(clusterName, providerName string) string {
	return filepath.Join(CfgPath, providerName, "clusters", clusterName)
}

// GetDataSource returns DB file path
func GetDataSource() string {
	return filepath.Join(CfgPath, DBFolder, DBFile)
}
