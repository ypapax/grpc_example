.PHONY: server client proto

server:
	go run ./server/

client:
	go run ./client/

proto:
	protoc \
		--go_out=. --go_opt=module=github.com/ypapax/grpc_example \
		--go-grpc_out=. --go-grpc_opt=module=github.com/ypapax/grpc_example \
		proto/order.proto
