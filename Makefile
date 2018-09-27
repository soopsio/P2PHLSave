VERSION := "0.2"
all:
	make docker-build
	make docker-push
	make docker-run
build:
	go build -o gosignaler .
run:
	./gosignaler
docker-build:
	make docker-build-signaler
	make docker-build-tracker
docker-build-signaler:
	docker build -t gosignaler:$(VERSION) .
docker-build-tracker:
	docker build -f tracker.Dockerfile -t gotracker:$(VERSION) .
docker-push:
	docker tag gosignaler:$(VERSION) etng/p2p_hlsaver_signaler:$(VERSION)
	docker tag gosignaler:$(VERSION) etng/p2p_hlsaver_signaler:latest
	docker push etng/p2p_hlsaver_signaler

	docker tag gotracker:$(VERSION) etng/p2p_hlsaver_tracker:$(VERSION)
	docker tag gotracker:$(VERSION) etng/p2p_hlsaver_tracker:latest
	docker push etng/p2p_hlsaver_tracker
docker-run:
	docker ps -qa -f name=gosignaler |xargs -r docker rm --force
	docker run -d  -p 8181:8181 --name gosignaler etng/p2p_hlsaver_signaler:$(VERSION)

	docker ps -qa -f name=gotracker |xargs -r docker rm --force
	docker run -d  -p 8787:8787 --name gotracker etng/p2p_hlsaver_tracker:$(VERSION)
deps:
	go get github.com/gorilla/mux
	go get github.com/go-redis/redis
	go get github.com/gorilla/websocket
tracker:
	go run tracker.go
