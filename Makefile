build:
	go build -o gosignaler .
run:
	./gosignaler
docker-build:
	docker build -t gosignaler:0.1 .
docker-run:
	# docker run -it --rm  --name gosignaler gosignaler:0.1
	docker rm -f gosignaler || 1
	docker run -d  -p 8181:8081 --name gosignaler gosignaler:0.1
deps:
	go get github.com/gorilla/mux
	go get github.com/go-redis/redis
	go get github.com/gorilla/websocket
tracker:
	go run tracker.go
