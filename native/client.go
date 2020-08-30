package native

import (
	"path/filepath"

	engine "github.com/jimi36/app-engine"
	"github.com/jimi36/app-engine/log"
	"github.com/jimi36/app-engine/store"
	"github.com/jimi36/app-engine/utils"
)

func NewClient(opts ...engine.Option) (engine.ClientImpl, error) {
	cli := &Client{
		basePath:     "/var/lib/engine/native",
		appInstances: make(map[string]*Instance),
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
	// store
	store engine.Store
	// base path
	basePath string
	// post task func
	postTaskEvent engine.PostTaskEventFunc
	// application instances
	appInstances map[string]*Instance
}

var _ engine.ClientImpl = (*Client)(nil)

func (cli *Client) Init(postFunc engine.PostTaskEventFunc) error {
	cli.postTaskEvent = postFunc
	return nil
}
