%if 0%{?fedora} || 0%{?rhel} >= 6
%global with_devel 1
# TODO: package new deps
%global with_bundled 0
%global with_debug 0
%global with_check 0
%global with_unit_test 0
%else
%global with_devel 0
%global with_bundled 0
%global with_debug 0
%global with_check 0
%global with_unit_test 0
%endif

%if 0%{?with_debug}
%global _dwz_low_mem_die_limit 0
%else
%global	debug_package	%{nil}
%endif

%global	provider github
%global	provider_tld com
%global project fusor
%global repo ansible-service-broker

%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path %{provider_prefix}

%define build_timestamp %(date +"%Y%m%d%H%M%%S")

Name: %{repo}
Version: 0
Release: 3.%{build_timestamp}%{?dist}
Summary: Ansible Service Broker
License: ASL 2.0
URL: https://%{provider_prefix}
Source0: https://%{provider_prefix}/archive/HEAD/%{repo}.tar.gz

# e.g. el6 has ppc64 arch without gcc-go, so EA tag is required
#ExclusiveArch: %%{?go_arches:%%{go_arches}}%%{!?go_arches:%%{ix86} x86_64 %{arm}}
ExclusiveArch: %{ix86} x86_64 %{arm} aarch64 ppc64le %{mips} s390x
# If go_compiler is not set to 1, there is no virtual provide. Use golang instead.
BuildRequires: %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}

BuildRequires: device-mapper-devel
BuildRequires: btrfs-progs-devel
%if ! 0%{?with_bundled}
BuildRequires: docker-devel
BuildRequires: kubernetes-devel
BuildRequires: runc-devel
BuildRequires: etcd-devel

BuildRequires: golang-github-gogo-protobuf-devel
BuildRequires: golang-github-ugorji-go-devel
BuildRequires: golang-github-vbatts-tar-split-devel

BuildRequires: golang(github.com/PuerkitoBio/purell)
BuildRequires: golang(github.com/PuerkitoBio/urlesc)
BuildRequires: golang(github.com/Azure/go-ansiterm)
BuildRequires: golang(github.com/containers/image)
BuildRequires: golang(github.com/containers/storage)
BuildRequires: golang(github.com/docker/distribution)
BuildRequires: golang(github.com/docker/go-connections)
BuildRequires: golang(github.com/docker/go-units)
BuildRequires: golang(github.com/docker/libtrust)
BuildRequires: golang(github.com/emicklei/go-restful)
BuildRequires: golang(github.com/fsouza/go-dockerclient)
BuildRequires: golang(github.com/ghodss/yaml)
BuildRequires: golang(github.com/go-openapi/jsonpointer)
BuildRequires: golang(github.com/go-openapi/jsonreference)
BuildRequires: golang(github.com/go-openapi/spec)
BuildRequires: golang(github.com/go-openapi/swag)
BuildRequires: golang(github.com/golang/glog)
BuildRequires: golang(github.com/google/gofuzz)
BuildRequires: golang(github.com/gorilla/context)
BuildRequires: golang(github.com/gorilla/mux)
BuildRequires: golang(github.com/hashicorp/go-cleanhttp)
BuildRequires: golang(github.com/imdario/mergo)
BuildRequires: golang(github.com/jessevdk/go-flags)
BuildRequires: golang(github.com/mailru/easyjson)
BuildRequires: golang(github.com/mattn/go-shellwords)
BuildRequires: golang(github.com/Microsoft/go-winio)
BuildRequires: golang(github.com/Microsoft/hcsshim)
BuildRequires: golang(github.com/mistifyio/go-zfs)
BuildRequires: golang(github.com/op/go-logging)
BuildRequires: golang(github.com/opencontainers/go-digest)
BuildRequires: golang(github.com/opencontainers/image-spec)
BuildRequires: golang(github.com/pborman/uuid)
BuildRequires: golang(github.com/pkg/errors)
BuildRequires: golang(github.com/Sirupsen/logrus)
BuildRequires: golang(github.com/spf13/pflag)
BuildRequires: golang(gopkg.in/yaml.v2)
BuildRequires: golang(k8s.io/apimachinery)
BuildRequires: golang(k8s.io/client-go)
%endif

%description
%{summary}

%if 0%{?with_devel}
%package devel
Summary: %{summary}
BuildArch: noarch

Requires: %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}
Requires: device-mapper-devel
Requires: btrfs-progs-devel
Requires: docker-devel
Requires: kubernetes-devel
Requires: runc-devel
Requires: etcd-devel
Requires: golang-github-gogo-protobuf-devel
Requires: golang-github-ugorji-go-devel
Requires: golang-github-vbatts-tar-split-devel
Requires: golang(github.com/PuerkitoBio/purell)
Requires: golang(github.com/PuerkitoBio/urlesc)
Requires: golang(github.com/Azure/go-ansiterm)
Requires: golang(github.com/containers/image)
Requires: golang(github.com/containers/storage)
Requires: golang(github.com/docker/distribution)
Requires: golang(github.com/docker/go-connections)
Requires: golang(github.com/docker/go-units)
Requires: golang(github.com/docker/libtrust)
Requires: golang(github.com/emicklei/go-restful)
Requires: golang(github.com/fsouza/go-dockerclient)
Requires: golang(github.com/ghodss/yaml)
Requires: golang(github.com/go-openapi/jsonpointer)
Requires: golang(github.com/go-openapi/jsonreference)
Requires: golang(github.com/go-openapi/spec)
Requires: golang(github.com/go-openapi/swag)
Requires: golang(github.com/golang/glog)
Requires: golang(github.com/google/gofuzz)
Requires: golang(github.com/gorilla/context)
Requires: golang(github.com/gorilla/mux)
Requires: golang(github.com/hashicorp/go-cleanhttp)
Requires: golang(github.com/imdario/mergo)
Requires: golang(github.com/jessevdk/go-flags)
Requires: golang(github.com/mailru/easyjson)
Requires: golang(github.com/mattn/go-shellwords)
Requires: golang(github.com/Microsoft/go-winio)
Requires: golang(github.com/Microsoft/hcsshim)
Requires: golang(github.com/mistifyio/go-zfs)
Requires: golang(github.com/op/go-logging)
Requires: golang(github.com/opencontainers/go-digest)
Requires: golang(github.com/opencontainers/image-spec)
Requires: golang(github.com/pborman/uuid)
Requires: golang(github.com/pkg/errors)
Requires: golang(github.com/Sirupsen/logrus)
Requires: golang(github.com/spf13/pflag)
Requires: golang(gopkg.in/yaml.v2)
Requires: golang(k8s.io/apimachinery)
Requires: golang(k8s.io/client-go)

%description devel
devel for %{name}
%{import_path} prefix.
%endif

%if 0%{?with_unit_test} && 0%{?with_devel}
%package unit-test
Summary: Unit tests for %{name} package
# If go_compiler is not set to 1, there is no virtual provide. Use golang instead.
BuildRequires: %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}

%if 0%{?with_check}
#Here comes all BuildRequires: PACKAGE the unit tests
#in %%check section need for running
%endif

# test subpackage tests code from devel subpackage
Requires: %{name}-devel = %{version}-%{release}

%description unit-test
unit-test for %{name}
%endif

%prep
%setup -q -n %{repo}-%{version}
mkdir -p src/github.com/fusor/ansible-service-broker
cp -r pkg src/github.com/fusor/ansible-service-broker

%build
export GOPATH=$(pwd):%{gopath}

BUILDTAGS="seccomp selinux"
%if ! 0%{?gobuild:1}
%define gobuild() go build -ldflags "${LDFLAGS:-} -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n')" -a -v -x %{**};

%endif

%gobuild -tags "$BUILDTAGS" ./cmd/broker
rm -rf src

%install
install -d -p %{buildroot}%{_bindir}
install -p -m 755 broker %{buildroot}%{_bindir}/asbd
install -p -m 755 build/%{name} %{buildroot}%{_bindir}/%{name}
sed -i 's,/usr/local/%{name}/bin,/usr/libexec/%{name},g' %{buildroot}%{_bindir}/%{name}
install -d -p %{buildroot}%{_docdir}/%{name}
install -d -p %{buildroot}%{_sysconfdir}/%{name}
install -p -m 755 etc/ex.dev.config.yaml %{buildroot}%{_docdir}/%{name}/ex.dev.config.yaml
install -p -m 755 etc/ex.dockerimg.config.yaml %{buildroot}%{_sysconfdir}/%{name}/config.yaml
install -p -m 755 etc/ex.dockerimg.config.yaml %{buildroot}%{_docdir}/%{name}/ex.dockerimg.config.yaml
install -p -m 755 etc/ex.prod.config.yaml %{buildroot}%{_docdir}/%{name}/ex.prod.config.yaml
install -d -p %{buildroot}%{_libexecdir}/%{name}
cp -r scripts/* %{buildroot}%{_libexecdir}/%{name}
install -d -p %{buildroot}%{_unitdir}
install -p extras/%{name}.service  %{buildroot}%{_unitdir}/%{name}.service

# source codes for building projects
%if 0%{?with_devel}
install -d -p %{buildroot}/%{gopath}/src/%{import_path}/
# find all *.go but no *_test.go files and generate devel.file-list
for file in $(find . -iname "*.go" \! -iname "*_test.go" | grep -v "^./Godeps") ; do
    echo "%%dir %%{gopath}/src/%%{import_path}/$(dirname $file)" >> devel.file-list
    install -d -p %{buildroot}/%{gopath}/src/%{import_path}/$(dirname $file)
    cp -pav $file %{buildroot}/%{gopath}/src/%{import_path}/$file
    echo "%%{gopath}/src/%%{import_path}/$file" >> devel.file-list
done
for file in $(find . -iname "*.proto" | grep -v "^./Godeps") ; do
    echo "%%dir %%{gopath}/src/%%{import_path}/$(dirname $file)" >> devel.file-list
    install -d -p %{buildroot}/%{gopath}/src/%{import_path}/$(dirname $file)
    cp -pav $file %{buildroot}/%{gopath}/src/%{import_path}/$file
    echo "%%{gopath}/src/%%{import_path}/$file" >> devel.file-list
done
%endif

# testing files for this project
%if 0%{?with_unit_test} && 0%{?with_devel}
install -d -p %{buildroot}/%{gopath}/src/%{import_path}/
# find all *_test.go files and generate unit-test.file-list
for file in $(find . -iname "*_test.go" | grep -v "^./Godeps"); do
    echo "%%dir %%{gopath}/src/%%{import_path}/$(dirname $file)" >> devel.file-list
    install -d -p %{buildroot}/%{gopath}/src/%{import_path}/$(dirname $file)
    cp -pav $file %{buildroot}/%{gopath}/src/%{import_path}/$file
    echo "%%{gopath}/src/%%{import_path}/$file" >> unit-test.file-list
done
%endif

%if 0%{?with_devel}
sort -u -o devel.file-list devel.file-list
%endif

%check
%if 0%{?with_check} && 0%{?with_unit_test} && 0%{?with_devel}
%if ! 0%{?with_bundled}
export GOPATH=%{buildroot}/%{gopath}:%{gopath}
%else
export GOPATH=%{buildroot}/%{gopath}:$(pwd)/Godeps/_workspace:%{gopath}
%endif

%if ! 0%{?gotest:1}
%global gotest go test
%endif

# FAIL: TestFactoryNewTmpfs (0.00s), factory_linux_test.go:59: operation not permitted
#%%gotest %%{import_path}/libcontainer
%gotest %{import_path}/libcontainer/cgroups
# --- FAIL: TestInvalidCgroupPath (0.00s)
#	apply_raw_test.go:16: couldn't get cgroup root: mountpoint for cgroup not found
#	apply_raw_test.go:25: couldn't get cgroup data: mountpoint for cgroup not found
#%%gotest %%{import_path}/libcontainer/cgroups/fs
%gotest %{import_path}/libcontainer/configs
%gotest %{import_path}/libcontainer/devices
# undefined reference to `nsexec'
#%%gotest %%{import_path}/libcontainer/integration
%gotest %{import_path}/libcontainer/label
# Unable to create tstEth link: operation not permitted
#%%gotest %%{import_path}/libcontainer/netlink
# undefined reference to `nsexec'
#%%gotest %%{import_path}/libcontainer/nsenter
%gotest %{import_path}/libcontainer/selinux
%gotest %{import_path}/libcontainer/stacktrace
#constant 2147483648 overflows int
#%%gotest %%{import_path}/libcontainer/user
#%%gotest %%{import_path}/libcontainer/utils
#%%gotest %%{import_path}/libcontainer/xattr
%endif

#define license tag if not already defined
%{!?_licensedir:%global license %doc}

%post
%systemd_post %{name}.service

%postun
%systemd_postun

%files
%license LICENSE
%{_bindir}/asbd
%{_bindir}/%{name}
%{_docdir}/%{name}
%dir %{_sysconfdir}/%{name}
%config %{_sysconfdir}/%{name}/config.yaml
%{_unitdir}/%{name}.service
%{_libexecdir}/%{name}

%if 0%{?with_devel}
%files devel -f devel.file-list
%license LICENSE
%dir %{gopath}/src/%{provider}.%{provider_tld}/%{project}
%dir %{gopath}/src/%{import_path}
%endif

%if 0%{?with_unit_test} && 0%{?with_devel}
%files unit-test -f unit-test.file-list
%license LICENSE
%endif

%changelog

