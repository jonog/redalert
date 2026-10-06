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
	cd web && go run github.com/GeertJohan/go.rice/rice@v1.1.0 embed-go

build-ui:
	cd ui && npm install && NODE_ENV=production ./node_modules/.bin/webpack -p && cd ..
	mkdir -p web/assets
	cp ui/dist/assets/app.bundle.js web/assets/
	cp ui/index.html web/assets

run-dev-ui:
	cd ui && npm install && ./node_modules/.bin/webpack-dev-server

build-proto:
	protoc -I servicepb/ servicepb/service.proto --go_out=plugins=grpc:servicepb

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

test-deps:
	docker pull sickp/alpine-sshd
	docker pull postgres

test: test-unit

test-unit:
	go test ./...

test-integration:
	REDALERT_INTEGRATION=1 go test ./checks -run 'Test(DockerStats_Check|Postgres_Check|RemoteCommand_Check(_MetadataExitStatus)?)$$' -count=1

build-docker-image-local: embed-static
	docker run --rm \
		-v "$(shell pwd):/src" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		centurylink/golang-builder \
		jonog/redalert

build-docker-image-remote: build-docker-image-local
	docker tag jonog/redalert jonog/redalert:v${VERSION}
	docker push jonog/redalert

.PHONY: install-deps embed-static build-ui build-proto clean test-deps test test-unit test-integration build build-docker-image build-docker-image-remote
