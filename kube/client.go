package kube

import (
	"os/user"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"

	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
	"github.com/jimi36/app-engine/store"
	"github.com/jimi36/app-engine/utils"
)

const (
	labelEdgeApp        = "edge-app"
	labelEdgeAppVersion = "edge-app-version"
)

const (
	handleTimeout = time.Second * 100
)

func NewClient(opts ...engine.Option) (engine.ClientImpl, error) {
	cli := &Client{
		ns:             "default",
		inKubeCluster:  false,
		kubeConfigPath: getDefaultKubeConfigPath(),
	}
	if err := applyOptions(cli, opts); err != nil {
		return nil, err
	}

	if err := utils.CreateFolder(cli.basePath); err != nil {
		return nil, err
	}

	if cli.store == nil {
		dbStore, err := store.NewLevelDBStore(filepath.Join(cli.basePath, "db"))
		if err != nil {
			log.Errorf("create store error: %s", err.Error())
			return nil, err
		}
		cli.store = dbStore
	}

	return cli, nil
}

type Client struct {
	ns             string
	inKubeCluster  bool
	kubeConfigPath string

	kubeCli    kubernetes.Interface
	metricsCli clientset.Interface

	// post task func
	postTaskEvent engine.PostTaskEventFunc

	basePath string

	store engine.Store
}

var _ engine.ClientImpl = (*Client)(nil)

func (cli *Client) Init(postFunc engine.PostTaskEventFunc) error {
	cli.postTaskEvent = postFunc

	go cli.monitorInstance()

	return nil
}

func getDefaultKubeConfigPath() string {
	u, err := user.Current()
	if err != nil {
		return "config"
	}
	return u.HomeDir + "/.kube/config"
}
