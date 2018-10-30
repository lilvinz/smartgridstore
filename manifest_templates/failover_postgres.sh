#!/bin/bash
# {{ .GenLine }}

# This script kills the master postgres and activates the slave as the master
# It is up to you to deploy a new slave
# (you can use postgresnewslave.daemonset.yaml)

MASTER=postgres-0
SLAVE=postgres-1
NAMESPACE={{ .TargetNamespace }}

set -ex

if [[ "$MASTER" == "" ]]
then
  echo "\$MASTER must be set"
  exit 1
fi

if [[ "$SLAVE" == "" ]]
then
  echo "\$SLAVE must be set"
  exit 1
fi

kubectl -n $NAMESPACE delete daemonset ${MASTER}
kubectl -n $NAMESPACE delete service ${MASTER}

slavepod=$(kubectl -n $NAMESPACE get pods -l cluster=postgres,role=slave -o go-template --template '{{`{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}`}}')
kubectl -n $NAMESPACE exec -it ${slavepod} -- su postgres -c '/usr/lib/postgresql/11/bin/pg_ctl promote'
kubectl -n $NAMESPACE patch daemonset ${SLAVE} --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/env/2/value", "value":"master"}]'
kubectl -n $NAMESPACE patch daemonset ${SLAVE} --type='json' -p='[{"op": "replace", "path": "/spec/template/metadata/labels/role", "value":"master"}]'

echo "Done"
