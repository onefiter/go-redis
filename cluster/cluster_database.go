package cluster

import (
	"context"
	"fmt"
	"github.com/go-redis/config"
	"github.com/go-redis/database"
	databaseface "github.com/go-redis/interface/database"
	"github.com/go-redis/interface/resp"
	"github.com/go-redis/lib/consistenthash"
	"github.com/go-redis/lib/logger"
	"github.com/go-redis/resp/reply"

	pool "github.com/jolestar/go-commons-pool/v2"
	"runtime/debug"
	"strings"
)

// ClusterDatabase represents a node of godis cluster
// it holds part of data and coordinates other nodes to finish transactions
type ClusterDatabase struct {
	self string

	nodes          []string
	peerPicker     *consistenthash.NodeMap
	peerConnection map[string]*pool.ObjectPool
	db             databaseface.Database
}

// MakeClusterDatabase creates and starts a node of cluster
func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self: config.Properties.Self,

		db:             database.NewStandaloneDatabase(),
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	cluster.peerPicker.AddNode(nodes...)
	ctx := context.Background()
	for _, peer := range config.Properties.Peers { // 新建一个连接池，连接工厂connectionFactory
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}
	cluster.nodes = nodes
	return cluster
}

// CmdFunc represents the handler of a redis command
type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply

// Close stops current node of cluster
func (cluster *ClusterDatabase) Close() {
	cluster.db.Close()
}

var router = makeRouter()

// Exec executes command on cluster
func (cluster *ClusterDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "', or not supported in cluster mode")
	}
	result = cmdFunc(cluster, c, cmdLine)
	return
}

// AfterClientClose does some clean after client close connection
func (cluster *ClusterDatabase) AfterClientClose(c resp.Connection) {
	cluster.db.AfterClientClose(c)
}
