package etcdregistry

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

//EtcdRegistry lib
type EtcdRegistry struct {
	etcdBasePath   string
	etcdEndpoints  []string
	defaultTimeout time.Duration
}

//Node registered service node
type Node struct {
	Name string
	Info map[string]string
}

//NewEtcdRegistry EtcdRegistry factory method
func NewEtcdRegistry(etcdEndpoints []string, etcdBasePath string, defaultTimeout time.Duration) (*EtcdRegistry, error) {
	r := &EtcdRegistry{}
	r.defaultTimeout = defaultTimeout
	r.etcdBasePath = etcdBasePath
	r.etcdEndpoints = etcdEndpoints
	return r, nil
}

//RegisterNode registers a new Node to a service with a TTL. After registration, TTL lease will be kept alive until node is unregistered or process killed
func (r *EtcdRegistry) RegisterNode(ctx context.Context, serviceName string, node Node, ttl time.Duration) error {
	if serviceName == "" {
		return fmt.Errorf("service name must be non empty")
	}
	if node.Name == "" {
		return fmt.Errorf("node.Name must be non empty")
	}
	if ttl.Seconds() <= 0 {
		return fmt.Errorf("ttl must be > 0")
	}

	r.keepRegistered(ctx, serviceName, node, ttl)
	return nil
}

func (r *EtcdRegistry) keepRegistered(ctx context.Context, serviceName string, node Node, ttl time.Duration) {
	for {
		err := r.registerNode(ctx, serviceName, node, ttl)
		if err != nil {
			logrus.Warnf("Registration got errors. Restarting. err=%s", err)
			time.Sleep(5 * time.Second)
		} else {
			logrus.Infof("Registration stopped with no errors")
			return
		}
	}
}

func (r *EtcdRegistry) registerNode(ctx context.Context, serviceName string, node Node, ttl time.Duration) error {
	cli, err := r.initializeETCDClient()
	if err != nil {
		return err
	}

	logrus.Debugf("Creating lease grant for %s/%s", serviceName, node.Name)
	ctx0, cancel := context.WithTimeout(context.Background(), r.defaultTimeout)
	defer cancel()
	lgr, err := cli.Grant(ctx0, int64(ttl.Seconds()))
	if err != nil {
		return err
	}

	logrus.Debugf("Creating node attribute for %s/%s", serviceName, node.Name)
	ctx0, cancel = context.WithTimeout(context.Background(), r.defaultTimeout)
	defer cancel()
	_, err = cli.Put(ctx0, r.nodePath(serviceName, node.Name), encode(node.Info), clientv3.WithLease(lgr.ID))
	if err != nil {
		return err
	}

	logrus.Debugf("Starting automatic keep alive for %s/%s", serviceName, node.Name)
	c, err := cli.KeepAlive(ctx, lgr.ID)
	if err != nil {
		return err
	}
	for m := range c {
		logrus.Debugf("[%s %s] %s", serviceName, node.Name, m)
	}
	return nil
}

//GetServiceNodes returns a list of active service nodes
func (r *EtcdRegistry) GetServiceNodes(serviceName string) ([]Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.defaultTimeout)
	defer cancel()

	cli, err := r.initializeETCDClient()
	if err != nil {
		return nil, err
	}

	rsp, err := cli.Get(ctx, r.servicePath(serviceName)+"/", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
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
		node.Name = string(n.Key)
		node.Info = decode(n.Value)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (r *EtcdRegistry) initializeETCDClient() (*clientv3.Client, error) {
	logrus.Debugf("Initializing ETCD client")
	cli, err := clientv3.New(clientv3.Config{Endpoints: r.etcdEndpoints, DialTimeout: r.defaultTimeout})
	if err != nil {
		logrus.Errorf("Could not initialize ETCD client. err=%s", err)
		return nil, err
	}
	logrus.Debugf("EtcdRegistry initialized with endpoints=%s, basePath=%s, defaultTimeout=%s", r.etcdEndpoints, r.etcdBasePath, r.defaultTimeout)
	return cli, nil
}

func encode(m map[string]string) string {
	if m != nil {
		b, _ := json.Marshal(m)
		return string(b)
	}
	return ""
}

func decode(ds []byte) map[string]string {
	if ds != nil && len(ds) > 0 {
		var s map[string]string
		json.Unmarshal(ds, &s)
		return s
	}
	return nil
}

func (r *EtcdRegistry) servicePath(serviceName string) string {
	service := strings.Replace(serviceName, "/", "-", -1)
	return path.Join(r.etcdBasePath, service)
}

func (r *EtcdRegistry) nodePath(serviceName string, nodeName string) string {
	service := strings.Replace(serviceName, "/", "-", -1)
	node := strings.Replace(nodeName, "/", "-", -1)
	return path.Join(r.etcdBasePath, service, node)
}
