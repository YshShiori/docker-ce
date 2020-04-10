module github.com/docker/docker

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/BurntSushi/toml v0.3.1
	github.com/Graylog2/go-gelf v0.0.0-20171211094031-414364622654
	github.com/Microsoft/go-winio v0.4.6
	github.com/Microsoft/hcsshim v0.6.8
	github.com/Microsoft/opengcs v0.3.6
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5
	github.com/RackSec/srslog v0.0.0-20161121191927-456df3a81436
	github.com/aws/aws-sdk-go v1.25.41
	github.com/bfirsh/funker-go v0.0.0-20161231111542-eaa0a2e06f30
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/boltdb/bolt v1.3.1
	github.com/bsphere/le_go v0.0.0-20170215134836-7a984a84b549
	github.com/cloudflare/cfssl v0.0.0-20160825002822-7fb22c8cba7e
	github.com/containerd/cgroups v0.0.0-20180309154421-fe281dd26576
	github.com/containerd/console v0.0.0-20180220200639-2748ece16665 // indirect
	github.com/containerd/containerd v1.0.1-0.20171227210547-3fa104f843ec
	github.com/containerd/continuity v0.0.0-20180216233310-d8fb8589b0e8
	github.com/containerd/fifo v0.0.0-20170714183902-fbfb6a11ec67
	github.com/containerd/go-runc v0.0.0-20180125000618-4f6e87ae043f // indirect
	github.com/containerd/typeurl v0.0.0-20170912152501-f6943554a7e7
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/deckarep/golang-set v0.0.0-20141123011944-ef32fa3046d9 // indirect
	github.com/dmcgowan/go-tar v0.0.0-20171110022031-01eea4a853c4 // indirect
	github.com/docker/distribution v2.6.0-rc.1.0.20170726174610-edc3ab29cdff+incompatible
	github.com/docker/go-connections v0.3.1-0.20180212134524-7beb39f0b969
	github.com/docker/go-events v0.0.0-20170721190031-9461782956ad // indirect
	github.com/docker/go-metrics v0.0.0-20170502235133-d466d4f6fd96
	github.com/docker/go-units v0.3.2-0.20170127094116-9e638d38cf69
	github.com/docker/libkv v0.2.2-0.20161109010621-1d8431073ae0
	github.com/docker/libnetwork v0.8.0-dev.2.0.20180314220121-1b91bc94094e
	github.com/docker/libtrust v0.0.0-20150526203908-9cbd2a1374f4
	github.com/docker/swarmkit v1.12.1-0.20180329182449-831df679a0b8
	github.com/fernet/fernet-go v0.0.0-20151007213151-1b2437bc582b // indirect
	github.com/fluent/fluent-logger-golang v1.3.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-check/check v0.0.0-20200227125254-8fa46927fb4f
	github.com/go-ini/ini v1.25.4 // indirect
	github.com/godbus/dbus v4.0.0+incompatible // indirect
	github.com/gogo/protobuf v1.2.1
	github.com/golang/gddo v0.0.0-20180130204405-9b12a26f3fbd
	github.com/google/certificate-transparency v0.0.0-20161025093837-d90e65c3a079 // indirect
	github.com/google/go-cmp v0.3.0
	github.com/googleapis/gax-go v0.0.0-20161107002406-da06d194a00e // indirect
	github.com/gorilla/context v0.0.0-20160226214623-1ea25387ff6f // indirect
	github.com/gorilla/mux v0.0.0-20160317213430-0eeaf8392f5b
	github.com/gotestyourself/gotestyourself v1.3.1-0.20180305173210-cf3a5ab914a2
	github.com/hashicorp/consul v1.7.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0
	github.com/hashicorp/go-memdb v1.0.3
	github.com/hashicorp/uuid v0.0.0-20160311170451-ebb0a03e909c // indirect
	github.com/imdario/mergo v0.3.6
	github.com/ishidawataru/sctp v0.0.0-20180213033435-07191f837fed // indirect
	github.com/kr/pty v1.1.1
	github.com/mattn/go-shellwords v1.0.3
	github.com/mistifyio/go-zfs v2.1.2-0.20160425201758-22c9b32c84eb+incompatible
	github.com/moby/buildkit v0.0.0-20170922161955-aaff9d591ef1
	github.com/mrunalp/fileutils v0.0.0-20171103030105-7d4729fb3618 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.0-rc5
	github.com/opencontainers/runtime-spec v1.0.1
	github.com/opencontainers/selinux v1.0.0-rc1.0.20171021164251-b29023b86e4a
	github.com/pborman/uuid v0.0.0-20160209185913-a97ce2ca70fa
	github.com/phayes/permbits v0.0.0-20190612203442-39d7c581d2ee // indirect
	github.com/philhofer/fwd v0.0.0-20160129035939-98c11a7a6ec8 // indirect
	github.com/pivotal-golang/clock v0.0.0-20151018222946-3fd3c1944c59 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3
	github.com/samuel/go-zookeeper v0.0.0-20150415181332-d0e0d8e11f31 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/seccomp/libseccomp-golang v0.0.0-20160531183505-32f571b70023
	github.com/sirupsen/logrus v1.2.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/cobra v0.0.7
	github.com/spf13/pflag v1.0.3
	github.com/stevvooe/ttrpc v0.0.0-20180205224448-d4528379866b // indirect
	github.com/syndtr/gocapability v0.0.0-20150716010906-2c00daeb6c3b
	github.com/tchap/go-patricia v2.2.6+incompatible
	github.com/tinylib/msgp v1.0.3-0.20171013044219-3b556c645408 // indirect
	github.com/tonistiigi/fsutil v0.0.0-20170929161712-dea3a0da73ae
	github.com/vbatts/tar-split v0.10.2
	github.com/vdemeester/shakers v0.1.0
	github.com/vishvananda/netlink v0.0.0-20171020171820-b2de5d10e38e
	github.com/vishvananda/netns v0.0.0-20150710222425-604eaf189ee8 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/net v0.0.0-20190923162816-aa69164e4478
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/api v0.21.0 // indirect
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.27.0
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0 // indirect
	labix.org/v2/mgo v0.0.0-20140701140051-000000000287 // indirect
)
