#!/bin/bash
# {{ .GenLine }}

TDIR=$(mktemp -d pg.XXXXXXXX)

head -c 5000 /dev/urandom | md5sum | head -c 20 > ${TDIR}/pgpassword
kubectl -n {{ .TargetNamespace }} create secret generic postgres-password --from-file=${TDIR}/pgpassword

head -c 5000 /dev/urandom | md5sum | head -c 20 > ${TDIR}/pgpassword
kubectl -n {{ .TargetNamespace }} create secret generic postgres-rpassword --from-file=${TDIR}/pgpassword

rm ${TDIR}/pgpassword
rmdir ${TDIR}
