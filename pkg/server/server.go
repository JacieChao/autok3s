package server

import (
	"net/http"

	"github.com/cnrancher/autok3s/pkg/server/ui"
	"github.com/gorilla/mux"
	responsewriter "github.com/rancher/apiserver/pkg/middleware"
	"github.com/rancher/apiserver/pkg/server"
	"github.com/rancher/apiserver/pkg/store/apiroot"
	"github.com/rancher/apiserver/pkg/types"
)

func Start() http.Handler {
	s := server.DefaultAPIServer()
	initMutual(s.Schemas)
	initProvider(s.Schemas)
	initCluster(s.Schemas)
	initCredential(s.Schemas)
	apiroot.Register(s.Schemas, []string{"v1"})
	router := mux.NewRouter()
	router.UseEncodedPath()
	router.StrictSlash(true)

	chain := responsewriter.Chain{
		responsewriter.Gzip,
		responsewriter.NoCache,
		responsewriter.DenyFrameOptions,
		responsewriter.ContentType,
		ui.UI,
	}
	router.Handle("/", chain.Handler(s))
	uiContent := responsewriter.Chain{
		responsewriter.Gzip,
		responsewriter.DenyFrameOptions,
		responsewriter.CacheMiddleware("json", "js", "css")}.Handler(ui.Content())
	router.PathPrefix("/css/").Handler(uiContent)
	router.PathPrefix("/js/").Handler(uiContent)
	router.PathPrefix("/img/").Handler(uiContent)
	router.PathPrefix("/fonts/").Handler(uiContent)

	router.Path("/{prefix}/{type}").Handler(s)
	router.Path("/{prefix}/{type}/{name}").Queries("link", "{link}").Handler(s)
	router.Path("/{prefix}/{type}/{name}").Queries("action", "{action}").Handler(s)
	router.Path("/{prefix}/{type}/{name}").Handler(s)

	router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Handle(&types.APIRequest{
			Request:   r,
			Response:  rw,
			Type:      "apiRoot",
			URLPrefix: "v1",
		})
	})

	return router
}
