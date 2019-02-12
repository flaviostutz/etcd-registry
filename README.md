# etcd-registry
ETCD Service Registry in Go.
When service node go live, it registers itself with a TTL in an ETCD server. If it dies, after TTL the registration vanishes.
At any time another process can query for that server which nodes are live.

# Usage

