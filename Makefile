all: test build
build:
	go build

test: godeps
	go test -race -cover

.PHONY: godeps
godeps:
	(cd $(GOPATH)/src/github.com/mholt/caddy 2>/dev/null              && git checkout -q master 2>/dev/null || true)
	go get -u github.com/mholt/caddy
	(cd $(GOPATH)/src/github.com/mholt/caddy              && git checkout -q v0.10.11)
	# github.com/flynn/go-shlex is required by mholt/caddy at the moment
	go get -u github.com/flynn/go-shlex
