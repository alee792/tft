.PHONY: bin
bin:
	go build ./cmd/tft

.PHONY: proto
proto: 
	protoc -I=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf --gogo_out=.\
		proto/tft.proto
