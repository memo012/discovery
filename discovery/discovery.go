package discovery

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/bilibili/discovery/conf"
	"github.com/bilibili/discovery/registry"
	http "github.com/bilibili/kratos/pkg/net/http/blademaster"
)

// Discovery discovery.
type Discovery struct {
	c         *conf.Config
	protected bool
	client    *http.Client
	registry  *registry.Registry
	nodes     atomic.Value
}

// New get a discovery.
func New(c *conf.Config) (d *Discovery, cancel context.CancelFunc) {
	d = &Discovery{
		protected: c.EnableProtect,
		c:         c,
		client:    http.NewClient(c.HTTPClient),
		registry:  registry.NewRegistry(c),
	}
	// 读取配置 初始化对等节点(Node/Nodes)
	d.nodes.Store(registry.NewNodes(c))
	// 自发现1 拉去discovery集群其他discovery节点信息
	d.syncUp()
	// 自发现2 注册自己
	cancel = d.regSelf()
	// 自发现3 遵循增量拉取discovery集群其他discovery节点信息
	go d.nodesproc()
	go d.exitProtect()
	return
}

func (d *Discovery) exitProtect() {
	// exist protect mode after two renew cycle
	time.Sleep(time.Second * 60)
	d.protected = false
}
