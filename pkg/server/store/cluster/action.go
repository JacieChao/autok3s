package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cnrancher/autok3s/pkg/providers"

	com "github.com/cnrancher/autok3s/cmd/common"
	"github.com/cnrancher/autok3s/pkg/cluster"
	"github.com/cnrancher/autok3s/pkg/common"
	"github.com/cnrancher/autok3s/pkg/types/apis"
	"github.com/cnrancher/autok3s/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/rancher/apiserver/pkg/apierror"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/pkg/schemas/validation"
	"github.com/sirupsen/logrus"
)

const (
	actionJoin = "join"
	linkNodes  = "nodes"
	linkLog    = "logs"
)

func Formatter(request *types.APIRequest, resource *types.RawResource) {
	resource.Links[linkNodes] = request.URLBuilder.Link(resource.Schema, resource.ID, linkNodes)
	resource.AddAction(request, actionJoin)
}

func HandleCluster() map[string]http.Handler {
	return map[string]http.Handler{
		actionJoin: joinHandler(),
	}
}

func LinkCluster(request *types.APIRequest) (types.APIObject, error) {
	if request.Link != "" {
		return nodesHandler(request, request.Schema, request.Name)
	}

	return request.Schema.Store.ByID(request, request.Schema, request.Name)
}

func joinHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		clusterID := vars["name"]
		if clusterID == "" {
			rw.WriteHeader(http.StatusUnprocessableEntity)
			rw.Write([]byte("clusterID cannot be empty"))
			return
		}
		content := strings.Split(clusterID, ".")
		providerName := content[len(content)-1]
		provider, err := providers.GetProvider(providerName)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(fmt.Sprintf("provider %s is not found", providerName)))
		}
		//c := &autok3stypes.Cluster{}
		body, err := ioutil.ReadAll(req.Body)
		//if err != nil {
		//	rw.WriteHeader(http.StatusInternalServerError)
		//	rw.Write([]byte(err.Error()))
		//	return
		//}
		//err = json.Unmarshal(body, c)
		//if err != nil {
		//	rw.WriteHeader(http.StatusInternalServerError)
		//	rw.Write([]byte(err.Error()))
		//	return
		//}
		//provider, err := com.GetProviderByState(*c)
		//if err != nil {
		//	rw.WriteHeader(http.StatusInternalServerError)
		//	rw.Write([]byte(err.Error()))
		//	return
		//}
		apiCluster := &apis.Cluster{}
		err = json.Unmarshal(body, apiCluster)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}
		err = provider.SetConfig(body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}
		err = provider.MergeClusterOptions()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}

		go func() {
			err := provider.JoinK3sNode(&apiCluster.SSH)
			if err != nil {
				logrus.Errorf("join cluster error: %v", err)
				err = provider.Rollback()
				logrus.Errorf("rollback cluster error: %v", err)
			}
		}()

		rw.WriteHeader(http.StatusOK)
	})
}

func nodesHandler(apiOp *types.APIRequest, schema *types.APISchema, id string) (types.APIObject, error) {
	content := strings.Split(id, ".")
	providerName := content[len(content)-1]
	v := common.CfgPath
	if v == "" {
		return types.APIObject{}, fmt.Errorf("[cluster] cfg path is empty")
	}

	clusters, err := utils.ReadYaml(v, common.StateFile)
	if err != nil {
		return types.APIObject{}, err
	}

	converts, err := cluster.ConvertToClusters(clusters)
	if err != nil {
		return types.APIObject{}, fmt.Errorf("[cluster] failed to unmarshal state file, msg: %s", err)
	}
	for _, con := range converts {
		if con.Provider == providerName && con.Name == id {
			provider, err := com.GetProviderByState(con)
			if err != nil {
				return types.APIObject{}, apierror.NewAPIError(validation.InvalidOption, fmt.Sprintf("cluster %s is not valid", id))
			}
			kubeCfg := fmt.Sprintf("%s/%s", common.CfgPath, common.KubeCfgFile)
			clusterInfo := provider.DescribeCluster(kubeCfg)
			return types.APIObject{
				Type:   schema.ID,
				ID:     id,
				Object: clusterInfo,
			}, nil
		}
	}
	return types.APIObject{}, apierror.NewAPIError(validation.NotFound, fmt.Sprintf("cluster %s is not found", id))
}
