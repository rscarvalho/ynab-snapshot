PROGRAM=ynab-snapshot

help:
	@echo Usage: $(PROGRAM) [-Date] [-Path] [-Token]

test:
	@go test

lint:
	@golint -set_exit_status