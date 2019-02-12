FROM golang:1.10 AS BUILD

#doing dependency build separated from source build optimizes time for developer, but is not required
#install external dependencies first
ADD /etcdregistry.dep $GOPATH/src/github.com/flaviostutz/etcd-registry/etcd-registry/etcdregistry.go
RUN go get -v github.com/flaviostutz/etcd-registry/etcd-registry

ADD /main.dep $GOPATH/src/etcd-registrar/main.go
RUN go get -v etcd-registrar

ADD /main.dep $GOPATH/src/etcd-watcher/main.go
RUN go get -v etcd-watcher

#now build source code
ADD /etcd-registry $GOPATH/src/github.com/flaviostutz/etcd-registry/etcd-registry
RUN go get -v github.com/flaviostutz/etcd-registry/etcd-registry
#RUN go test -v etcd-registry

ADD /etcd-registrar $GOPATH/src/etcd-registrar
RUN go get -v etcd-registrar

ADD /etcd-watcher $GOPATH/src/etcd-watcher
RUN go get -v etcd-watcher



FROM golang:1.10

ENV LOG_LEVEL 'info'

COPY --from=BUILD /go/bin/* /bin/
ADD startup.sh /

ENV ETCD_URL ""
ENV ETCD_BASE ""
ENV SERVICE ""
ENV NAME ""
ENV INFO ""
ENV TTL 60

CMD [ "/startup.sh" ]

