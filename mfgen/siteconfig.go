package main

import (
	"io/ioutil"

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
			StagingPool   string `yaml:"stagingPool"`
			BTrDBDataPool string `yaml:"btrdbDataPool"`
			BTrDBHotPool  string `yaml:"btrdbHotPool"`
			RBDPool       string `yaml:"rbdPool"`
		} `yaml:"ceph"`
		ExternalIPs []string `yaml:"externalIPs"`
	} `yaml:"siteInfo"`
}

const DefaultSiteConfig = `apiVersion: smartgrid.store/v1
kind: SiteConfig
# this is the kubernetes namespace that you are deploying into
targetNamespace: sgs
containers:
  #this can be 'development'
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
    # the RBD pool is used to provision persistent storage for
    # kubernetes pods. It can use spinning metal.
    rbdPool: rbd

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
