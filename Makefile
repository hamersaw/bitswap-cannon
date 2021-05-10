all: binaries

binaries: cmd/bitswap-cannon

FORCE:

cmd/%: FORCE
	@echo "$@"
	@go build -o "./bin/$$(basename $@)" "./$@"

clean:
	@echo "$@"
	@rm -rf ./bin/bitswap-cannon
