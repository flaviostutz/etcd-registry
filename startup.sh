#!/bin/bash
set -e
set -x

echo "Starting etcd-registrar..."
etcd-registrar \
    --loglevel=$LOG_LEVEL \
    --etcd-url=$ETCD_URL \
    --etcd-base=$ETCD_BASE \
    --service=$SERVICE \
    --name=$NAME \
    --info=$INFO \
    --ttl=$TTL

