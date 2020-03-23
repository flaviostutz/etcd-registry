# etcd-registry
ETCD Service Registry in Go.
On each service node that is launched, register itself using this library so that the registration will be kept in an ETCD server.
After registration, another client can discover which nodes are available and connect to them if desired.
You must define a TTL for the registration so that if the node goes down after a while, the registration key in ETCD gets vanished and all watchers can know it is no longer available. The TTL keep alive is performed automatically by the library

If you need a standalone application for registration, see https://github.com/flaviostutz/etcd-registrar

# Usage

1. To register a running Node:
```go
    registry, _ := etcdregistry.NewEtcdRegistry([]string{"http://etcd0:2379"}, "/services", 5*time.Second)
    node := etcdregistry.Node{}
	node.Name = "test"
	err := registry.RegisterNode(context.TODO(), "myservice", node, 5*time.Second)
```

2. To query available service nodes
```go
    registry, _ := etcdregistry.NewEtcdRegistry([]string{"http://etcd0:2379"}, "/services", 5*time.Second)
    nodes, _ := registry.GetServiceNodes("myservice")
```

See [sample/main.go](sample/main.go) for a simple example

## Run an Example

1. Create a docker-compose.yml

```yml
version: '3.5'

services:

  sample:
    build: .

  etcd0:
    image: quay.io/coreos/etcd:v3.2.25
    environment:
      - ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd0:2379

  etcd3-lucas:
    image: registry.cn-hangzhou.aliyuncs.com/ringtail/lucas:0.0.1
    ports:
      - 8888:8080
    environment:
      - ENDPOINTS=http://etcd0:2379
```

2. Run ```docker-compose up```

3. See logs

4. Run ```docker-compose scale sample=10```

5. Open browser at http://localhost:8888, click on "/" and find all registered nodes on ETCD tree

6. Run ```docker-compose scale sample=3```

7. Check on browser for nodes being reajusted

# More

* Check github.com/flaviostutz/etcd-registrar for an utility that can be used with your existing application to register and keep alive the service node in parallel to the application execution
