#!/bin/bash
echo "Invoking BTrDB Custom Postgres Script"
set -ex

DATADIR=/var/lib/postgresql/data
BACKUPDIR=/var/lib/postgresql/backups

if [[ $ROLE == "master" ]]
then
  REPDIR=${DATADIR}/rep_data
  mkdir -p $REPDIR

  ls ${DATADIR}
  if [ ! -f ${DATADIR}/initial_setup_complete ]
  then
    echo "RUNNING MASTER INITIIAL SETUP"
    echo "host     replication     repuser         0.0.0.0/0        md5" >> ${DATADIR}/pg_hba.conf
    echo "wal_level = hot_standby" >> ${DATADIR}/postgresql.conf
    echo "archive_mode = on" >> ${DATADIR}/postgresql.conf
    echo "archive_command = 'test ! -f ${REPDIR}/%f && cp %p ${REPDIR}/%f'" >> ${DATADIR}/postgresql.conf
    echo "max_wal_senders = 10" >> ${DATADIR}/postgresql.conf

    /usr/lib/postgresql/11/bin/createuser -U postgres repuser -w -c 15 --replication
    psql -U postgres postgres --command "ALTER USER repuser WITH PASSWORD '${POSTGRES_RPASSWORD}';"
    touch ${DATADIR}/initial_setup_complete
  else
    echo "SKIPPING MASTER INIITIAL SETUP"
  fi

elif [[ $ROLE == "slave" ]]
then
  ls ${DATADIR}
  if [[ $POSTGRES_MASTER == "" ]]
  then
    echo "Missing \$POSTGRES_MASTER"
    exit 1
  fi
  echo "*:*:*:repuser:${POSTGRES_RPASSWORD}" > ~/.pgpass
  chmod 0600 ~/.pgpass

  if [ ! -f ${DATADIR}/initial_setup_complete ]
  then
    echo "RUNNING SLAVE INITIAL SETUP"
    pg_ctl -D "${DATADIR}" -m fast -w stop
    mkdir /tmp/ptmp
    cp ${DATADIR}/*.conf /tmp/ptmp
    backupdir=${BACKUPDIR}/backup_$(date +%s)
    mkdir -p ${backupdir}
    mv ${DATADIR}/* ${backupdir}/
    echo "PG_BASEBACKUP_START"
    unset POSTGRES_PASSWORD
    unset PGPASSWORD
    pg_basebackup -h ${POSTGRES_MASTER} -U repuser -D ${DATADIR} -v --wal-method=stream
    echo "PG_BASEBACKUP_END"
    echo "hot_standby = on" >> ${DATADIR}/postgresql.conf
    echo "standby_mode = on" >> ${DATADIR}/recovery.conf
    echo "primary_conninfo = 'host=${POSTGRES_MASTER} port=5432 user=repuser password=${POSTGRES_RPASSWORD}'" >> ${DATADIR}/recovery.conf
    touch ${DATADIR}/initial_setup_complete
    pg_ctl -D "${DATADIR}" -o "-c listen_addresses=''" -w start
  else
    echo "SKIPPING SLAVE INITIAL SETUP"
  fi

else
echo "Missing \$ROLE"
exit 1
fi
