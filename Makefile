BINDIR:=./bin
BINARY:=bootstrap
ZIPFILE:=$(BINARY).zip
CMD:=./cmd
REPORT:=./report

.PHONY: deps clean build deploy test vet fmt
deps:
	go get -u ./...

clean:
	rm -rf $(BINDIR)

test:
	go test -cover ./...

cover:
	mkdir -p $(REPORT)
	go test ./... -coverprofile $(REPORT)/cover.out
	go tool cover -html=$(REPORT)/cover.out -o $(REPORT)/index.html
	cd $(REPORT) && python3 -m http.server 8000
	

vet:
	go vet ./...

fmt:
	go fmt ./...
