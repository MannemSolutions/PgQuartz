build:
	./set_version.sh
	go mod tidy
	go build ./cmd/pgquartz

debug:
	go build -gcflags "all=-N -l" ./cmd/pgquartz
	~/go/bin/dlv --headless --listen=:2345 --api-version=2 --accept-multiclient exec ./pgquartz -- -c ./config.yaml

run:
	./pgquartz -c jobs/jobspec1/job.yml

fmt:
	gofmt -w .
	goimports -w .
	gci write .

compose:
	./docker-compose-tests.sh

test: sec lint

sec:
	gosec ./...
lint:
	golangci-lint run
