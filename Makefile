TAG=v0.0.1
BINARY=requestl
NAME=ektowett/$(BINARY)
IMAGE=$(NAME):$(TAG)
LATEST=$(NAME):latest

VERSION?=?
GIT_COMMIT := $(shell git rev-list -1 HEAD)
SHA1 := $(shell git rev-parse HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
NOW := $(shell date +'%Y-%m-%d_%T')
LDFLAGS := -ldflags "-X github.com/etowett/requestl/build.Sha1Ver=$(SHA1) -X github.com/etowett/requestl/build.Time=$(NOW) -X github.com/etowett/requestl/build.GitCommit=$(GIT_COMMIT) -X github.com/etowett/requestl/build.GitBranch=$(BRANCH) -X github.com/etowett/requestl/build.Version=$(VERSION)"

.PHONY: all build

build:
	CGO_ENABLED=0 GOOS=linux go build $(LDFLAGS) -a -installsuffix cgo -o $(BINARY) cmd/requestl/main.go
	docker build -t $(IMAGE) . -f Dockerfile-local
	docker tag $(IMAGE) $(LATEST)
	rm $(BINARY)

push:
	docker push $(IMAGE)
	docker push $(LATEST)

up:
	docker-compose up -d

rm: stop
	docker-compose rm

stop:
	docker-compose stop

logs:
	docker-compose logs -f

ps:
	docker-compose ps

compile:
	go build -v -o /tmp/requestl cmd/requestl/main.go && rm /tmp/requestl
