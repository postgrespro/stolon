PROJDIR=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))

# change to project dir so we can express all as relative paths
$(shell cd $(PROJDIR))

REPO_PATH=github.com/sorintlab/stolon

VERSION ?= $(shell scripts/git-version.sh)

LD_FLAGS="-w -X $(REPO_PATH)/cmd.Version=$(VERSION)"

$(shell mkdir -p bin )

ifneq (, $(shell which dpkg))
dpkg_arch := $(shell dpkg --print-architecture)
dpkg_dir := stolon-shardman-$(VERSION).$(dpkg_arch)
define CONTROLFILE
Package: stolon-shardman
Version: $(VERSION)
Architecture: $(dpkg_arch)
Maintainer: Postgrespro Build <build@postgrespro.ru>
Description: stolon is a cloud native PostgreSQL manager for PostgreSQL high availability. It's cloud native because it'll let you keep an high available PostgreSQL inside your containers (kubernetes integration) but also on every other kind of infrastructure (cloud IaaS, old style infrastructures etc...)
endef
endif
export CONTROLFILE


.PHONY: all
all: build

.PHONY: build
build: sentinel keeper proxy stolonctl

.PHONY: test
test: build
	./test

.PHONY: sentinel keeper proxy stolonctl docker

keeper:
	GO111MODULE=on go build -ldflags $(LD_FLAGS) -o $(PROJDIR)/bin/stolon-keeper $(REPO_PATH)/cmd/keeper

sentinel:
	CGO_ENABLED=0 GO111MODULE=on go build -ldflags $(LD_FLAGS) -o $(PROJDIR)/bin/stolon-sentinel $(REPO_PATH)/cmd/sentinel

proxy:
	CGO_ENABLED=0 GO111MODULE=on go build -ldflags $(LD_FLAGS) -o $(PROJDIR)/bin/stolon-proxy $(REPO_PATH)/cmd/proxy

stolonctl:
	CGO_ENABLED=0 GO111MODULE=on go build -ldflags $(LD_FLAGS) -o $(PROJDIR)/bin/stolonctl $(REPO_PATH)/cmd/stolonctl

.PHONY: docker
docker:
	if [ -z $${PGVERSION} ]; then echo 'PGVERSION is undefined'; exit 1; fi; \
	if [ -z $${TAG} ]; then echo 'TAG is undefined'; exit 1; fi; \
	docker build --build-arg PGVERSION=${PGVERSION} -t ${TAG} -f examples/kubernetes/image/docker/Dockerfile .

deb:
ifdef dpkg_arch
	rm -rf $(dpkg_dir)
	mkdir -p $(dpkg_dir)/opt/pgpro/shardman/bin
	cp -p bin/* $(dpkg_dir)/opt/pgpro/shardman/bin
	mkdir -p $(dpkg_dir)/DEBIAN
	echo "$$CONTROLFILE" > $(dpkg_dir)/DEBIAN/control
	dpkg-deb --build --root-owner-group $(dpkg_dir)
	rm -rf $(dpkg_dir)
else
	@echo dpkg not found!
endif
