APP = pophttpd

LINUX_ARCH = x86_64
LINUX_UNAME = $(LINUX_ARCH)-linux-gnu

DESTINATIONS = buster.local:bin/ ripley.local:bin/ tesla.local:bin/
LINUX_DESTINATIONS = nibbler.local:bin/

all: deploy

deploy: build build_linux
	for dest in $(DESTINATIONS); \
	do \
		scp $(APP) $$dest; \
	done
	for dest in $(LINUX_DESTINATIONS); \
	do \
		scp $(APP).$(LINUX_UNAME) $$dest/$(APP); \
	done

install:
	go install

build:
	go build

build_linux:
	GOOS=linux ARCH=$(LINUX_ARCH) go build -o $(APP).$(LINUX_UNAME)

.PHONY: all deploy install build build_linux
