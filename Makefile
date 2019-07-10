PROGRAM=ynab-snapshot

test:
	go test -v ./ ./client

help:
	@echo Usage: $(PROGRAM) [-Date] [-Path] [-Token]

lint:
	@golint -set_exit_status

install:
	@go install github.com/rscarvalho/ynab-snapshot