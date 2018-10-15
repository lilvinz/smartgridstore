package main

import (
	"io/ioutil"
	"path"

	yaml "gopkg.in/yaml.v2"
)

type SiteConfig struct {
	ApiVersion      string `yaml:"apiVersion"`
	TargetNamespace string `yaml:"targetNamespace"`
	Containers      struct {
		Channel         string `yaml:"channel"`
		ImagePullPolicy string `yaml:"imagePullPolicy"`
	} `yaml:"containers"`
	SiteInfo struct {
		Ceph struct {
			StagingPool      string `yaml:"stagingPool"`
			BTrDBDataPool    string `yaml:"btrdbDataPool"`
			BTrDBHotPool     string `yaml:"btrdbHotPool"`
			BTrDBJournalPool string `yaml:"btrdbJournalPool"`
			MountConfig      bool   `yaml:"mountConfig"`
			ConfigFile       string `yaml:"configFile"`
			ConfigPath       string `yaml:"-"`
		} `yaml:"ceph"`
		Etcd struct {
			Nodes   []string `yaml:"nodes"`
			Version string   `yaml:"version"`
		}
		ExternalIPs []string `yaml:"externalIPs"`
	} `yaml:"siteInfo"`
}

const DefaultSiteConfig = `apiVersion: smartgrid.store/v1
kind: SiteConfig
# this is the kubernetes namespace that you are deploying into
targetNamespace: sgs
containers:
  #this can be 'development' or 'release'
  channel: release
  imagePullPolicy: Always
siteInfo:
  ceph:
    # the staging pool is where ingress daemons stage data. It can be
    # ignored if you are not using ingress daemons
    stagingPool: staging

    # the btrdb data pool (or cold pool) is where most of the data
    # is stored. It is typically large and backed by spinning metal drives
    btrdbDataPool: btrdb_data
    # the btrdb hot pool is used for more performance-sensitive
    # data. It can be smaller and is usually backed by SSDs
    btrdbHotPool: btrdb_hot
    # the btrdb journal pool is small and used to aggregate write
    # load into better-performing batches. It should be backed by
    # a very high performance pool
    btrdbJournalPool: btrdb_journal
    # If you are using Rook, the recommended way to connect to ceph
    # is to mount the config file directly. Uncomment this to do so
    # configFile: /var/lib/rook/rook-ceph/rook-ceph.config
    # mountConfig: true

  etcd:
    # which nodes should run etcd servers. There must be three entries
    # here, so if you have fewer nodes, you can have duplicates. The node
    # names must match the output from 'kubectl get nodes'
    nodes:
    - host0
    - host1
    - host2
    # which version of etcd to deploy
    version: v3.3.5

  # the external IPs listed here are where the services can be contacted
  # e.g for the plotter or the BTrDB API
  externalIPs:
  - 123.123.123.1
`

func LoadSiteConfig(filename string) (*SiteConfig, error) {
	ba, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return LoadSiteConfigArray(ba)
}

func LoadSiteConfigArray(arr []byte) (*SiteConfig, error) {
	rv := SiteConfig{}
	er := yaml.Unmarshal(arr, &rv)
	if er != nil {
		return nil, er
	}
	//Some special cases
	if rv.SiteInfo.Ceph.ConfigFile == "" {
		rv.SiteInfo.Ceph.ConfigFile = "/etc/ceph/ceph.conf"
	}
	if rv.SiteInfo.Ceph.MountConfig {
		rv.SiteInfo.Ceph.ConfigPath = path.Dir(rv.SiteInfo.Ceph.ConfigFile)
	}

	return &rv, nil
}

func (sc *SiteConfig) TargetVersion() string {
	return PackageVersion
}
func (sc *SiteConfig) Pfx() string {
	if sc.Containers.Channel == "release" {
		return ""
	}
	return "dev-"
}
func (sc *SiteConfig) GenLine() string {
	return "generated by SGS MFGEN v" + PackageVersion
}

//
// 	cephStagingPool: staging
// 	cephBTrDBPool: btrdb
// 	cephRBDPool:
// stack:
//   btrdb:
//     replicas: 3
//     resources:
//       cpu: "20"
//       memory: "96Gi"
//     cephPool: btrdb
//   mrplotter:
//     replicas: 2
//   upmuReceiver:
//     replicas: 2
//     port: 1883
//     cephPool: staging
//   upmuIngester:
//     cephPool: staging
//   pmu2btrdb:
//     replicas: 2
//     port: 1884
//   storageClass:
//     createRBDStorageClass: yes
//     cephPool: rbd
// ---`
