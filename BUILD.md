# Compilation Guide

## Project Structure

```
myidea/
├── go.work          # Go workspace — no go.mod at root
├── Makefile         # proto + tidy + build
├── api/             # Shared protocol (proto + generated Go stubs)
├── service/         # Kratos backend — 5 gRPC servers
├── bff/             # Gin API gateway
└── web/             # Vue 3 SPA frontend
```

This is a **Go workspace** (`go.work`). There is **no `go.mod`** at the project root. Each Go sub-module has its own `go.mod`. The frontend (`web/`) uses Node.js/npm.

---

## Prerequisites

| Tool | Purpose |
|------|---------|
| Go 1.25+ | Build & run backend |
| Node.js 20+ | Build & run frontend |
| MySQL 8.0+ | Data storage |
| `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc` | Proto codegen (only when changing `.proto`) |

Install proto tools:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="$HOME/go/bin:$PATH"
```

---

## Quick Start (Backend)

### 1. Generate protobuf stubs

```bash
cd /Users/storm/githubproject/myidea
export PATH="$HOME/go/bin:$PATH"
make proto
```

Regenerates `.pb.go` and `_grpc.pb.go` for all proto files in `api/`.

### 2. Build & vet

```bash
# ✅ Correct — build each module from its own directory
cd service && go build ./... && go vet ./...
cd bff    && go build ./... && go vet ./...

# ✅ Also correct — from workspace root, specify the module path
go build ./service/...
go build ./bff/...

# ❌ Will NOT work — go.work root does not support `go build ./...`
go build ./...
```

### 3. Full build (proto + service + bff)

```bash
cd /Users/storm/githubproject/myidea
export PATH="$HOME/go/bin:$PATH"
make build
```

---

## Quick Start (Frontend)

### 1. Install dependencies

```bash
cd /Users/storm/githubproject/myidea/web
npm install
```

### 2. Development server

```bash
npm run dev
```

Starts Vite dev server on `:5173`. API requests to `/api/*` are proxied to BFF at `http://localhost:8080`.

### 3. Production build

```bash
npm run build
```

Output: `web/dist/`.

---

## Run

### 1. Start MySQL

```bash
docker run -d \
  --name mall-mysql \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=mall_order \
  -p 3306:3306 \
  mysql:8.0
```

### 2. Start Services (5 independent gRPC servers)

Each service runs independently in its own process:

```bash
cd /Users/storm/githubproject/myidea/service

# 可选：覆盖默认 DSN
export DB_DSN="root:password@tcp(127.0.0.1:3306)/mall_order?charset=utf8mb4&parseTime=True&loc=Local"

# 分别启动各个服务（每个终端启动一个）
go run ./cmd/order/     # :9000
go run ./cmd/product/   # :9001
go run ./cmd/cart/      # :9002
go run ./cmd/user/      # :9003
go run ./cmd/payment/   # :9004
```

每个服务独立进程运行，各自监听对应端口。`AutoMigrate` 会自动创建所有表。

### 3. Start BFF (HTTP gateway)

```bash
cd /Users/storm/githubproject/myidea/bff

# 可选：设置 JWT 密钥
export JWT_SECRET="my-secret-key"

go run .
```

BFF 监听 `:8080`。

### 4. Start Frontend (development)

```bash
cd /Users/storm/githubproject/myidea/web
npm run dev
```

Vite dev server 监听 `:5173`。在浏览器打开 http://localhost:5173。

---

## Quick API Test

### Health check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Register + Login

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"pass123","email":"test@example.com"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"pass123"}' | \
  python3 -c "import sys,json;print(json.load(sys.stdin)['data']['token'])")

echo $TOKEN
```

### Full checkout flow (requires MySQL with real data)

```bash
# Add item to cart
curl -X POST http://localhost:8080/api/v1/cart/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"sku_id":1,"product_id":1,"spu_id":1,"title":"Test Phone","price":5999,"quantity":2}'

# Checkout (LockStock → Order → Payment)
curl -X POST http://localhost:8080/api/v1/cart/checkout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"address":{"receiver_name":"John","receiver_phone":"13800138000","detail_address":"Beijing"},"remark":"fast"}'
```

---

## Common Commands

### Tidy dependencies

```bash
cd api      && go mod tidy
cd service  && go mod tidy
cd bff      && go mod tidy
```

### Proto generation (detailed)

```bash
protoc \
  --proto_path=. \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  api/cart/v1/cart.proto
```

---

## Troubleshooting

### "pattern ./...: does not contain modules listed in go.work"

**Cause**: Running `go build ./...` from the `go.work` root directory.

**Fix**: Build per-module instead:
```bash
go build ./service/...
go build ./bff/...
```

### "protoc-gen-go: program not found"

**Cause**: `protoc-gen-go` or `protoc-gen-go-grpc` not in `PATH`.

**Fix**: Add Go binaries to PATH:
```bash
export PATH="$HOME/go/bin:$PATH"
make proto
```

### "undefined: pbCart.CartServiceClient" or similar

**Cause**: Proto stubs were not generated, or are stale after proto changes.

**Fix**: Regenerate stubs and rebuild:
```bash
make proto
cd service && go build ./...
```

### "order.gRPC.ProductAddr" / "gRPC.CartAddr" field missing

**Cause**: `conf/conf.go` or `bff/config/config.go` was updated but not rebuilt.

**Fix**: After modifying config structs, always rebuild from scratch:
```bash
cd service && go build ./... && go vet ./...
cd bff    && go build ./... && go vet ./...
```

### `go mod tidy` removes needed dependencies

**Cause**: Go sums are out of sync after adding new proto-generated files.

**Fix**: Run tidy for all modules in order:
```bash
cd api      && go mod tidy
cd service  && go mod tidy
cd bff      && go mod tidy
```

---

## Port Reference

| Service | Protocol | Port |
|---------|----------|------|
| Order gRPC | gRPC | `:9000` |
| Product gRPC | gRPC | `:9001` |
| Cart gRPC | gRPC | `:9002` |
| User gRPC | gRPC | `:9003` |
| Payment gRPC | gRPC | `:9004` |
| BFF HTTP | HTTP | `:8080` |
| Vite Dev Server | HTTP | `:5173` |
