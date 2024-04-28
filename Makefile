test:
	CGO_ENABLED=0 go test -count=1 ./... || exit 1

test-race:
	CGO_ENABLED=1 go test -race -count=1 ./... || exit 1