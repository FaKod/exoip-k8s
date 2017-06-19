all: exoip-k8s

TAG = latest
PREFIX = innoq/exoip-k8s

exoip-k8s:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o exoip-k8s main.go

container: exoip-k8s
	docker build --pull -t $(PREFIX):$(TAG) .

push: container
	docker push $(PREFIX):$(TAG)

clean:
	rm -f exoip-k8s
