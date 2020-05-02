TAG=v0.0.1
BINARY=requestl
NAME=ektowett/$(BINARY)
IMAGE=$(NAME):$(TAG)
LATEST=$(NAME):latest

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY) -v
	docker build -t $(IMAGE) .
	docker tag $(IMAGE) $(LATEST)
	rm $(BINARY)

application:
	make -C ../smsl-django/ up

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
	go build -v -o /tmp/requestl . && rm /tmp/requestl
