# build:
# 	go build

# run: build
# 	./alex -c config.json

local_run:
	go build && ./alex -c config-local.json

build:
	docker-compose build

run:
	docker-compose up