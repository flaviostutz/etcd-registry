package etcdregistry

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

//EtcdRegistry lib
type EtcdRegistry struct {
	client *clientv3.Client
	sync.Mutex
	register       map[string]uint64
	leases         map[string]clientv3.LeaseID
	basePath       string
	defaultTimeout time.Duration
}

//Node registered service node
type Node struct {
	name string
	info map[string]string
}

//NewEtcdRegistry EtcdRegistry factory method
func NewEtcdRegistry(etcdEndpoints []string, etcdBasePath string, defaultTimeout time.Duration) (*EtcdRegistry, error) {
	r := &EtcdRegistry{}
	r.defaultTimeout = defaultTimeout
	r.basePath = etcdBasePath

	cli, err := clientv3.New(clientv3.Config{Endpoints: etcdEndpoints, DialTimeout: defaultTimeout})
	if err != nil {
		logrus.Errorf("Could not initialize ETCD client. err=%s", err)
		return nil, err
	}
	r.client = cli
	logrus.Debugf("EtcdRegistry initialized with endpoints=%s, basePath=%s, defaultTimeout=%s", etcdEndpoints, etcdBasePath, defaultTimeout)
	return r, nil
}

//RegisterNode registers a new Node to a service with a TTL. After registration, TTL lease will be kept alive until node is unregistered or process killed
func (r *EtcdRegistry) RegisterNode(ctx context.Context, serviceName string, node *Node, ttl time.Duration) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name must be non empty")
	}
	if node == nil {
		return nil, fmt.Errorf("node must be defined")
	}
	if node.name == "" {
		return nil, fmt.Errorf("node.name must be non empty")
	}
	if ttl.Seconds() <= 0 {
		return nil, fmt.Errorf("ttl must be > 0")
	}

	logrus.Debugf("Creating lease grant for %s/%s", serviceName, node.name)
	ctx0, cancel := context.WithTimeout(context.Background(), r.defaultTimeout)
	defer cancel()
	lgr, err := r.client.Grant(ctx0, int64(ttl.Seconds()))
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Creating node attribute for %s/%s", serviceName, node.name)
	ctx0, cancel = context.WithTimeout(context.Background(), r.defaultTimeout)
	defer cancel()
	_, err = r.client.Put(ctx0, r.nodePath(serviceName, node.name), encode(node.info), clientv3.WithLease(lgr.ID))
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Starting automatic keep alive for %s/%s", serviceName, node.name)
	c, err := r.client.KeepAlive(ctx, lgr.ID)
	if err != nil {
		return nil, err
	}
	return c, nil
}

//GetServiceNodes returns a list of active service nodes
func (r *EtcdRegistry) GetServiceNodes(serviceName string) ([]Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.defaultTimeout)
	defer cancel()

	rsp, err := r.client.Get(ctx, r.servicePath(serviceName)+"/", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, 0)

	if len(rsp.Kvs) == 0 {
		logrus.Debugf("No services nodes were found under %s", r.servicePath(serviceName)+"/")
		return nodes, nil
	}

	for _, n := range rsp.Kvs {
		node := Node{}
		node.name = string(n.Key)
		node.info = decode(n.Value)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func encode(m map[string]string) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func decode(ds []byte) map[string]string {
	var s map[string]string
	json.Unmarshal(ds, &s)
	return s
}

func (r *EtcdRegistry) servicePath(serviceName string) string {
	service := strings.Replace(serviceName, "/", "-", -1)
	return path.Join(r.basePath, service)
}

func (r *EtcdRegistry) nodePath(serviceName string, nodeName string) string {
	service := strings.Replace(serviceName, "/", "-", -1)
	node := strings.Replace(nodeName, "/", "-", -1)
	return path.Join(r.basePath, service, node)
}
