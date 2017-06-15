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
%global project openshift
%global repo ansible-service-broker

%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path %{provider_prefix}

%if 0%{?copr}
%define build_timestamp .%(date +"%Y%m%d%H%M%%S")
%else
%define build_timestamp %{nil}
%endif

%define selinux_variants targeted
%define moduletype apps
%define modulename ansible-service-broker

Name: %{repo}
Version: 0
Release: 3%{build_timestamp}%{?dist}
Summary: Ansible Service Broker
License: ASL 2.0
URL: https://%{provider_prefix}
Source0: https://%{provider_prefix}/archive/HEAD/%{repo}.tar.gz

# e.g. el6 has ppc64 arch without gcc-go, so EA tag is required
#ExclusiveArch: %%{?go_arches:%%{go_arches}}%%{!?go_arches:%%{ix86} x86_64 %{arm}}
ExclusiveArch: %{ix86} x86_64 %{arm} aarch64 ppc64le %{mips} s390x
# If go_compiler is not set to 1, there is no virtual provide. Use golang instead.
BuildRequires: %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}

Requires(pre): shadow-utils
Requires: %{name}-selinux

BuildRequires: device-mapper-devel
BuildRequires: btrfs-progs-devel
%if ! 0%{?with_bundled}
%endif

%description
%{summary}

%package container-scripts
Summary: scripts required for running ansible-service-broker in a container
BuildArch: noarch

%description container-scripts
containers scripts for ansible-service-broker

%package selinux
Summary: selinux policy module for %{name}
BuildRequires: checkpolicy, selinux-policy-devel, hardlink, policycoreutils
BuildRequires: /usr/bin/pod2man
Requires: selinux-policy >= %{selinux_policy_ver}
Requires(post): /usr/sbin/semodule, /sbin/restorecon, /usr/sbin/setsebool, /usr/sbin/selinuxenabled, /usr/sbin/semanage
Requires(post): policycoreutils-python
Requires(post): selinux-policy-targeted
Requires(postun): /usr/sbin/semodule, /sbin/restorecon
BuildArch: noarch

%description selinux
selinux policy module for %{name}

%post selinux
for selinuxvariant in %{selinux_variants}
do
  /usr/sbin/semodule -s ${selinuxvariant} -i \
    %{_datadir}/selinux/${selinuxvariant}/%{modulename}.pp.bz2 > /dev/null
done

%postun selinux
if [ $1 -eq 0 ] ; then
  for selinuxvariant in %{selinux_variants}
  do
    /usr/sbin/semodule -s ${selinuxvariant} -r %{modulename} > /dev/null
  done
fi

%pre
getent group ansibleservicebroker || groupadd -r ansibleservicebroker
getent passwd ansibleservicebroker || \
  useradd -r -g ansibleservicebroker -d /var/lib/ansibleservicebroker -s /sbin/nologin \
  ansibleservicebroker
exit 0

%post
%systemd_post %{name}.service

%postun
%systemd_postun

%if 0%{?with_devel}
%package devel
Summary: %{summary}
BuildArch: noarch

Requires: %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}
Requires: device-mapper-devel
Requires: btrfs-progs-devel

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
ln -sf vendor src
mkdir -p src/github.com/openshift/ansible-service-broker
cp -r pkg src/github.com/openshift/ansible-service-broker

%build
export GOPATH=$(pwd):%{gopath}
export LDFLAGS='-s -w'
BUILDTAGS="seccomp selinux"
%if ! 0%{?gobuild:1}
%define gobuild() go build -ldflags "${LDFLAGS:-} -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n')" -a -v -x %{**};
%endif

%gobuild -tags "$BUILDTAGS" ./cmd/broker

#Build selinux modules
# create selinux-friendly version from VR and replace it inplace
perl -i -pe 'BEGIN { $VER = join ".", grep /^\d+$/, split /\./, "%{version}.%{release}"; } s!\@\@VERSION\@\@!$VER!g;' extras/%{modulename}.te

%if 0%{?rhel} >= 6
    distver=rhel%{rhel}
%endif
%if 0%{?fedora} >= 18
    distver=fedora%{fedora}
%endif

for selinuxvariant in %{selinux_variants}
do
    pushd extras
    make NAME=${selinuxvariant} -f /usr/share/selinux/devel/Makefile DISTRO=${distver}
    bzip2 -9 %{modulename}.pp
    mv %{modulename}.pp.bz2 %{modulename}.ppbz2.${selinuxvariant}
    make NAME=${selinuxvariant} -f /usr/share/selinux/devel/Makefile clean DISTRO=${distver}
    popd
done


rm -rf src

%install
install -d -p %{buildroot}%{_bindir}
install -p -m 755 broker %{buildroot}%{_bindir}/asbd
install -p -m 755 build/entrypoint.sh %{buildroot}%{_bindir}/entrypoint.sh
install -p -m 755 build/%{name} %{buildroot}%{_bindir}/%{name}
sed -i 's,/usr/local/%{name}/bin,/usr/libexec/%{name},g' %{buildroot}%{_bindir}/%{name}
install -d -p %{buildroot}%{_docdir}/%{name}
install -d -p %{buildroot}%{_sysconfdir}/%{name}
install -p -m 755 etc/ex.dev.config.yaml %{buildroot}%{_docdir}/%{name}/ex.dev.config.yaml
install -p -m 755 etc/ex.dockerimg.config.yaml %{buildroot}%{_docdir}/%{name}/ex.dockerimg.config.yaml
install -p -m 755 etc/ex.prod.config.yaml %{buildroot}%{_docdir}/%{name}/ex.prod.config.yaml
install -p -m 644 build/config.yaml %{buildroot}%{_sysconfdir}/%{name}/config.yaml
install -d -p %{buildroot}%{_libexecdir}/%{name}
cp -r scripts/* %{buildroot}%{_libexecdir}/%{name}
install -d -p %{buildroot}%{_unitdir}
install -p extras/%{name}.service  %{buildroot}%{_unitdir}/%{name}.service
install -d -p %{buildroot}%{_var}/log/%{name}
touch %{buildroot}%{_var}/log/%{name}/asb.log

# install selinux policy modules
for selinuxvariant in %{selinux_variants}
  do
    install -d %{buildroot}%{_datadir}/selinux/${selinuxvariant}
    install -p -m 644 extras/%{modulename}.ppbz2.${selinuxvariant} \
        %{buildroot}%{_datadir}/selinux/${selinuxvariant}/%{modulename}.pp.bz2
  done

# install interfaces
install -d %{buildroot}%{_datadir}/selinux/devel/include/%{moduletype}
install -p -m 644 extras/%{modulename}.if %{buildroot}%{_datadir}/selinux/devel/include/%{moduletype}/%{modulename}.if

# hardlink identical policy module packages together
/usr/sbin/hardlink -cv %{buildroot}%{_datadir}/selinux

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

%files
%license LICENSE
%{_bindir}/asbd
%{_bindir}/%{name}
%{_docdir}/%{name}
%attr(750, ansibleservicebroker, ansibleservicebroker) %dir %{_sysconfdir}/%{name}
%attr(640, ansibleservicebroker, ansibleservicebroker) %config %{_sysconfdir}/%{name}/config.yaml
%{_unitdir}/%{name}.service
%{_libexecdir}/%{name}
%attr(750, ansibleservicebroker, ansibleservicebroker) %dir %{_var}/log/%{name}
%attr(640, ansibleservicebroker, ansibleservicebroker) %{_var}/log/%{name}/asb.log

%files container-scripts
%{_bindir}/entrypoint.sh

%files selinux
%attr(0600,root,root) %{_datadir}/selinux/*/%{modulename}.pp.bz2
%{_datadir}/selinux/devel/include/%{moduletype}/%{modulename}.if

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

