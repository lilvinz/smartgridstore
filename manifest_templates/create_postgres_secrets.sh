#!/bin/bash
# {{ .GenLine }}

TDIR=$(mktemp -d pg.XXXXXXXX)

head -c 5000 /dev/urandom | md5sum | head -c 20 > ${TDIR}/pgpassword
kubectl -n {{ .TargetNamespace }} create secret generic postgres-root-password --from-file=${TDIR}/pgpassword

head -c 5000 /dev/urandom | md5sum | head -c 20 > ${TDIR}/pgpassword
kubectl -n {{ .TargetNamespace }} create secret generic postgres-repl-password --from-file=${TDIR}/pgpassword

head -c 5000 /dev/urandom | md5sum | head -c 20 > ${TDIR}/pgpassword
kubectl -n {{ .TargetNamespace }} create secret generic postgres-read-password --from-file=${TDIR}/pgpassword

head -c 5000 /dev/urandom | md5sum | head -c 20 > ${TDIR}/pgpassword
kubectl -n {{ .TargetNamespace }} create secret generic postgres-readwrite-password --from-file=${TDIR}/pgpassword

rm ${TDIR}/pgpassword
rmdir ${TDIR}
