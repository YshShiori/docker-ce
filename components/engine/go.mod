module github.com/docker/docker

go 1.13

require (
	cloud.google.com/go v0.26.0
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/BurntSushi/toml v0.3.1
	github.com/Graylog2/go-gelf v0.0.0-20171211094031-414364622654
	github.com/Microsoft/go-winio v0.4.6
	github.com/Microsoft/hcsshim v0.6.8
	github.com/Microsoft/opengcs v0.3.6
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5
	github.com/RackSec/srslog v0.0.0-20161121191927-456df3a81436
	github.com/armon/go-metrics v0.0.0-20150106224455-eb0af217e5e9
	github.com/armon/go-radix v0.0.0-20150105235045-e39d623f12e8
	github.com/aws/aws-sdk-go v1.12.66
	github.com/beorn7/perks v1.0.0
	github.com/bfirsh/funker-go v0.0.0-20161231111542-eaa0a2e06f30
	github.com/boltdb/bolt v1.3.1-0.20160913165339-fff57c100f4d
	github.com/bsphere/le_go v0.0.0-20170215134836-7a984a84b549
	github.com/cloudflare/cfssl v0.0.0-20160825002822-7fb22c8cba7e
	github.com/containerd/cgroups v0.0.0-20180309154421-fe281dd26576
	github.com/containerd/console v0.0.0-20180220200639-2748ece16665
	github.com/containerd/containerd v1.0.1-0.20171227210547-3fa104f843ec
	github.com/containerd/continuity v0.0.0-20180216233310-d8fb8589b0e8
	github.com/containerd/fifo v0.0.0-20170714183902-fbfb6a11ec67
	github.com/containerd/go-runc v0.0.0-20180125000618-4f6e87ae043f
	github.com/containerd/typeurl v0.0.0-20170912152501-f6943554a7e7
	github.com/coreos/etcd v3.3.10+incompatible
	github.com/coreos/go-semver v0.2.0
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/deckarep/golang-set v0.0.0-20141123011944-ef32fa3046d9
	github.com/dmcgowan/go-tar v0.0.0-20171110022031-01eea4a853c4
	github.com/docker/distribution v2.6.0-rc.1.0.20170726174610-edc3ab29cdff+incompatible
	github.com/docker/go-connections v0.3.1-0.20180212134524-7beb39f0b969
	github.com/docker/go-events v0.0.0-20170721190031-9461782956ad
	github.com/docker/go-metrics v0.0.0-20170502235133-d466d4f6fd96
	github.com/docker/go-units v0.3.2-0.20170127094116-9e638d38cf69
	github.com/docker/libkv v0.2.2-0.20161109010621-1d8431073ae0
	github.com/docker/libnetwork v0.8.0-dev.2.0.20180314220121-1b91bc94094e
	github.com/docker/libtrust v0.0.0-20150526203908-9cbd2a1374f4
	github.com/docker/swarmkit v1.12.1-0.20180329182449-831df679a0b8
	github.com/fernet/fernet-go v0.0.0-20151007213151-1b2437bc582b
	github.com/fluent/fluent-logger-golang v1.3.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-check/check v0.0.0-20190902080502-41f04d3bba15
	github.com/go-ini/ini v1.25.4
	github.com/godbus/dbus v4.0.0+incompatible
	github.com/gogo/protobuf v1.2.1
	github.com/golang/gddo v0.0.0-20180130204405-9b12a26f3fbd
	github.com/golang/protobuf v1.3.1
	github.com/google/certificate-transparency v0.0.0-20161025093837-d90e65c3a079
	github.com/google/go-cmp v0.2.0
	github.com/googleapis/gax-go v0.0.0-20161107002406-da06d194a00e
	github.com/gorilla/context v0.0.0-20160226214623-1ea25387ff6f
	github.com/gorilla/mux v0.0.0-20160317213430-0eeaf8392f5b
	github.com/gotestyourself/gotestyourself v1.3.1-0.20180305173210-cf3a5ab914a2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/consul v0.5.2
	github.com/hashicorp/go-immutable-radix v0.0.0-20160222071000-8e8ed81f8f0b
	github.com/hashicorp/go-memdb v0.0.0-20161216180745-cb9a474f84cc
	github.com/hashicorp/go-msgpack v0.0.0-20140221154404-71c2886f5a67
	github.com/hashicorp/go-multierror v0.0.0-20150127051936-fcdddc395df1
	github.com/hashicorp/go-sockaddr v0.0.0-20170324171533-acd314c5781e
	github.com/hashicorp/golang-lru v0.0.0-20160207214719-a0d98a5f2880
	github.com/hashicorp/memberlist v0.1.1-0.20171201184301-3d8438da9589
	github.com/hashicorp/serf v0.7.1-0.20160317193612-598c54895cc5
	github.com/imdario/mergo v0.0.0-20151231081848-bc0f15622cd2
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/ishidawataru/sctp v0.0.0-20180213033435-07191f837fed
	github.com/jmespath/go-jmespath v0.0.0-20160202185014-0b12d6b521d8
	github.com/kr/pty v1.1.1
	github.com/mattn/go-shellwords v1.0.3
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/miekg/dns v0.0.0-20151202105714-75e6e86cc601
	github.com/mistifyio/go-zfs v2.1.2-0.20160425201758-22c9b32c84eb+incompatible
	github.com/moby/buildkit v0.0.0-20170922161955-aaff9d591ef1
	github.com/mrunalp/fileutils v0.0.0-20171103030105-7d4729fb3618 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.0-rc5
	github.com/opencontainers/runtime-spec v1.0.1
	github.com/opencontainers/selinux v1.0.0-rc1.0.20171021164251-b29023b86e4a
	github.com/pborman/uuid v0.0.0-20160209185913-a97ce2ca70fa
	github.com/philhofer/fwd v0.0.0-20160129035939-98c11a7a6ec8
	github.com/pivotal-golang/clock v0.0.0-20151018222946-3fd3c1944c59
	github.com/pkg/errors v0.8.1-0.20161002052512-839d9e913e06
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v0.9.3
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.4.0
	github.com/prometheus/procfs v0.0.0-20190507164030-5867b95ac084
	github.com/samuel/go-zookeeper v0.0.0-20150415181332-d0e0d8e11f31
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529
	github.com/seccomp/libseccomp-golang v0.0.0-20160531183505-32f571b70023
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.3
	github.com/stevvooe/ttrpc v0.0.0-20180205224448-d4528379866b
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/syndtr/gocapability v0.0.0-20150716010906-2c00daeb6c3b
	github.com/tchap/go-patricia v2.2.6+incompatible
	github.com/tinylib/msgp v1.0.3-0.20171013044219-3b556c645408
	github.com/tonistiigi/fsutil v0.0.0-20170929161712-dea3a0da73ae
	github.com/ugorji/go v1.1.4
	github.com/vbatts/tar-split v0.10.2
	github.com/vdemeester/shakers v0.1.0
	github.com/vishvananda/netlink v0.0.0-20171020171820-b2de5d10e38e
	github.com/vishvananda/netns v0.0.0-20150710222425-604eaf189ee8
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
	golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a
	golang.org/x/text v0.3.0
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/api v0.0.0-20161128233437-3cc2e591b550
	google.golang.org/genproto v0.0.0-20180817151627-c66870c02cf8
	google.golang.org/grpc v1.21.0
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
)
