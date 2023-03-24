JOB ?= jobs/jobspec1/job.yml

uname_p := $(shell uname -p) # store the output of the command in a variable

build: build_pgquartz

build_pgquartz:
	sh ./set_version.sh
	go mod tidy -compat=1.17
	go build -o ./bin/pgquartz ./cmd/pgquartz
	ln ./bin/pgquartz ./bin/pgquartz.$(uname_p)

build_dlv:
	go get github.com/go-delve/delve/cmd/dlv@latest
	go build -o /bin/dlv.$(uname_p) github.com/go-delve/delve/cmd/dlv

build_image:
	docker build . --tag mannemsolutions/pgquartz

# Use the following on m1:
# alias make='/usr/bin/arch -arch arm64 /usr/bin/make'
debug:
	go build -gcflags "all=-N -l" -o ./bin/pgquartz.debug.$(uname_p) ./cmd/pgquartz
	~/go/bin/dlv --headless --listen=:2345 --api-version=2 --accept-multiclient exec ./bin/pgquartz.debug.$(uname_p) -- -c '$(JOB)'

debug_test:
	~/go/bin/dlv --headless --listen=:2345 --api-version=2 --accept-multiclient test ./pkg/git/

run:
	./bin/pgquartz.$(uname_p) -c '$(JOB)'

fmt:
	gofmt -w .
	goimports -w .
	gci write .

compose:
	./docker-compose-tests.sh

test: gotest sec lint

sec:
	gosec ./...
lint:
	golangci-lint run
gotest:
	go test -v ./...
