BINARY  := zarazaex
PORT    := 8801
LDFLAGS := -s -w

.PHONY: build run clean docker docker-run

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) server.go

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

docker:
	docker build -t $(BINARY) .

docker-run: docker
	docker run --rm -p $(PORT):$(PORT) $(BINARY)
