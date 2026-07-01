.PHONY: proto

# Generate protobuf code.
proto:
	protoc \
		--proto_path=. \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		api/order/v1/order.proto \
		api/product/v1/product.proto \
		api/cart/v1/cart.proto \
		api/user/v1/user.proto \
		api/payment/v1/payment.proto

# Tidy all modules.
tidy:
	cd api && go mod tidy
	cd service && go mod tidy
	cd bff && go mod tidy

# Build all modules.
build: proto
	cd service && go build ./...
	cd bff && go build ./...
