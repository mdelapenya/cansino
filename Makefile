.PHONY: build
build:
	docker-compose -f ./docker/docker-compose.yml build

.PHONY: destroy
destroy:
	docker-compose -f docker/docker-compose.yml down

.PHONY: start
start:
	docker-compose -f ./docker/docker-compose.yml up