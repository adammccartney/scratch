PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin


CC := gcc
CFLAGS := 

EXE := coro-sieve

default: $(EXE)

$(EXE): coro-sieve.c
	$(CC) $(CFLAGS) -o $@ $^

install: $(EXE)
	@echo "Installing $(EXE) to $(BINDIR)"
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(EXE) $(DESTDIR)$(BINDIR)/$(EXE)

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(EXE)

docker_build:
	docker build --tag coro-sieve:test -f Dockerfile .


clean:
	rm -f $(EXE)
