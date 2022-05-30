.PHONY: build install clean

build:
	CGO_ENABLED=0 go build -mod=vendor -o bin/annotations-exporter github.com/alex123012/annotations-exporter/cmd/annotations-exporter
install:
	CGO_ENABLED=0 go install github.com/alex123012/annotations-exporter/cmd/annotations-exporter
clean:
	rm -f $$GOPATH/bin/annotations-exporter
	rm -f bin/*
