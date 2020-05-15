# build:
# 	go build

# run: build
# 	./alex -c config.json

build:
	docker-compose build

run:
	docker-compose up