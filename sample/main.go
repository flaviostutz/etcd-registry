package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/flaviostutz/etcd-registry/etcd-registry"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	reg, err := etcdregistry.NewEtcdRegistry([]string{"http://etcd0:2379"}, "/services", 5*time.Second)
	if err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UnixNano())

	ctx := context.TODO()
	go registerNode(ctx, reg, "test", randomString(5))
	go registerNode(ctx, reg, "test", randomString(5))
	go registerNode(ctx, reg, "test", randomString(5))
	go registerNode(ctx, reg, "test", randomString(5))
	go registerNode(ctx, reg, "test", randomString(5))
	go watchService(reg, "test")

	logrus.Infof("Running...")
	<-ctx.Done()
	logrus.Infof("Done")
}

func registerNode(ctx context.Context, reg *etcdregistry.EtcdRegistry, service string, nodeName string) {

	node := etcdregistry.Node{}
	node.Name = nodeName
	err := reg.RegisterNode(ctx, service, node, 5*time.Second)
	if err != nil {
		panic(err)
	}
}

func watchService(reg *etcdregistry.EtcdRegistry, service string) {
	for {
		nodes, err0 := reg.GetServiceNodes(service)
		if err0 != nil {
			panic(err0)
		}
		logrus.Infof("Nodes: %v", nodes)
		time.Sleep(2 * time.Second)
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
