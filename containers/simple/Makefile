SHELL := /bin/sh
DOCKER := /usr/bin/docker
GO = $(shell which go)

TAG := mydate_simple:0.9
CONTAINER_NAME = mydate_simple
CONTEXT = $(shell pwd)/mydate

# Go scripts
CMD = $(shell pwd)/cmd

default: none

docker_run: docker_setup
	$(shell pwd)/docker_run.sh $(TAG) $(CONTAINER_NAME)

docker_setup:
	$(shell pwd)/docker_prerun.sh $(TAG) $(CONTEXT)

clean:
	$(DOCKER) rm -f $(CONTAINER_NAME)
	$(DOCKER) rmi $(TAG)

.PHONY: docker_run docker_setup clean
