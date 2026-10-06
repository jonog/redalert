BINARY=redalert

VERSION=0.2.4
COMMIT=$(shell git rev-parse HEAD)

LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT}"

GO_VERSION=1.27.1

install-deps:
	go mod download

build:
	go build ${LDFLAGS} -o ${BINARY} .

embed-static: build-ui
	# The dashboard files are embedded by web/web.go with //go:embed.

build-ui:
	cd ui && npm ci && NODE_OPTIONS=--openssl-legacy-provider NODE_ENV=production ./node_modules/.bin/webpack -p
	mkdir -p web/assets
	cp ui/dist/assets/app.bundle.js web/assets/
	cp ui/index.html web/assets

run-dev-ui:
	cd ui && npm ci && NODE_OPTIONS=--openssl-legacy-provider ./node_modules/.bin/webpack-dev-server

build-proto:
	protoc -I servicepb/ servicepb/service.proto --go_out=plugins=grpc:servicepb

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

test-deps:
	docker pull sickp/alpine-sshd@sha256:0f5a58ba5bfc5549a910264f32c337903967bb377d596c91c03611f15b4699ad
	docker pull postgres:9.5

test: test-unit

test-unit:
	go test ./...

test-integration:
	DOCKER_API_VERSION=$${DOCKER_API_VERSION:-1.24} POSTGRES_IMAGE=postgres:9.5 REDALERT_INTEGRATION=1 go test -v ./checks -run 'Test(DockerStats_Check|Postgres_Check|RemoteCommand_Check(_MetadataExitStatus)?)$$' -count=1

build-docker-image-local:
	docker run --rm \
		-v "$(shell pwd):/src" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		golang:$(GO_VERSION) \
		sh -c 'cd /src && CGO_ENABLED=0 go build -buildvcs=false -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)" -o /src/redalert .'
	docker build -t jonog/redalert .

build-docker-image-remote: build-docker-image-local
	docker tag jonog/redalert jonog/redalert:v${VERSION}
	docker push jonog/redalert

.PHONY: install-deps embed-static build-ui build-proto clean test-deps test test-unit test-integration build build-docker-image build-docker-image-remote
