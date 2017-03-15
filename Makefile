BINARY=redalert

VERSION=0.2.1
BUILD=`git rev-parse HEAD`

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

build: build-static
	go build ${LDFLAGS} -o ${BINARY}

build-static:
	go get github.com/GeertJohan/go.rice
	go get github.com/GeertJohan/go.rice/rice
	cd web && rice embed-go && cd ..

build-proto:
	protoc -I servicepb/ servicepb/service.proto --go_out=plugins=grpc:servicepb

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

test-deps:
	docker pull sickp/alpine-sshd
	docker pull postgres

test:
	go test -v -race $(shell glide novendor)

build-docker-image-local: build-static
	docker run --rm \
		-v "$(shell pwd):/src" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		centurylink/golang-builder \
		jonog/redalert

build-docker-image-remote: build-docker-image-local
	docker tag jonog/redalert jonog/redalert:v${VERSION}
	docker push jonog/redalert

.PHONY: build-static build-proto clean test-deps test build-docker-image build-docker-image-remote
