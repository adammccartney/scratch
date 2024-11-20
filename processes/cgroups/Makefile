PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

DOCKER_COMPOSE_DIR ?= /etc/docker/compose
SERVICE ?= coro-sieve

DOCKER_SOURCES = Makefile $(SOURCE) Dockerfile docker-compose.yml entrypoint.sh README.md

SYSTEMD_SERVICE_DIR ?= /etc/systemd/system

SOURCE = coro-sieve.c
EXE    = coro-sieve

CC := gcc
CFLAGS := 

default: $(EXE)

$(EXE): $(SOURCE)
	$(CC) $(CFLAGS) -o $@ $^

define install_docker_file
	install -o 1000 -g 1000 -m 644 $1 $(DOCKER_COMPOSE_DIR)/$(SERVICE)
endef

define install_docker_compose_service
	install -m 644 docker-compose@.service $(SYSTEMD_SERVICE_DIR)
endef

debug:
	$(foreach i,a b c d,$(call test_docker,$(i)))

install: $(EXE) $(DOCKER_SOURCES)
	@echo "Installing $(EXE) to $(BINDIR)"
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(EXE) $(DESTDIR)$(BINDIR)/$(EXE)
	install -d $(DOCKER_COMPOSE_DIR)/$(SERVICE)
	$(foreach _file,$(DOCKER_SOURCES),$(call install_docker_file,$(_file));)
	$(call install_docker_compose_service)

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(EXE)
	rm -rf $(DOCKER_COMPOSE_DIR)/$(SERVICE)
	rm -rf $(SYSTEMD_SERVICE_DIR)/docker-compose@.service

docker_build:
	docker build --tag coro-sieve:test -f Dockerfile .

clean:
	rm -f $(EXE)
