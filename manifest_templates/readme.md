# ordering

The full documentation can be found on smartgrid.store, but here is the order
you should create things:

### core

Create the secrets first

Secrets:
```
core/create_admin_key.sh
core/create_postgres_secret.sh
```

Then do etcd:
```
core/etcd-cluster.daemonset.yaml
```

Then do the metadata database:
```
core/postgres.daemonset.yaml
```

After those five pods are up and running, you can then create the database:

```
core/ensuredb.job.yaml
```

Wait for that to finish and then do btrdb

```
core/btrdb.statefulset.yaml
```

Wait for that to be running and do the rest:

```
core/adminconsole.deployment.yaml
core/mrplotter.deployment/yaml
```

### ingress

you can do the ingress at any stage after the core has been created
