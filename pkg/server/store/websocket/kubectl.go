package websocket

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/pkg/schemas/validation"
)

func KubeHandler(apiOp *types.APIRequest) (types.APIObject, error) {
	err := ptyHandler(apiOp)
	if err != nil {
		return types.APIObject{}, err
	}
	return types.APIObject{}, validation.ErrComplete
}

func ptyHandler(apiOp *types.APIRequest) error {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	c, err := upgrader.Upgrade(apiOp.Response, apiOp.Request, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	w := NewWriter(c)
	reader := NewReader(c)
	// symbolic link for kubectl
	os.Symlink(fmt.Sprintf("%s kubectl", os.Args[0]), "kubectl")
	if err != nil {
		return err
	}

	kubeBash := exec.Command("sh")
	// Start the command with a pty.
	ptmx, err := pty.StartWithSize(kubeBash, &pty.Winsize{
		Cols: 300,
		Rows: 150,
	})
	if err != nil {
		return err
	}

	defer ptmx.Close()

	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			_, err := w.Write([]byte("ping"))
			if err != nil {
				return err
			}
		default:
			go func() {
				io.Copy(ptmx, reader)
			}()
			io.Copy(w, ptmx)
		}
	}
}
