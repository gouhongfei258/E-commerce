# Mall E-Commerce Microservice

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A production-grade e-commerce platform built with Go + Vue 3, following microservice architecture. Features a **Kratos** gRPC service layer, a **Gin** BFF (Backend For Frontend) API gateway, a **Vue 3** SPA frontend, gRPC inter-service communication, JWT authentication, OpenTelemetry distributed tracing, and GORM-based persistence.

## Architecture

```
┌──────────────────┐       HTTP/JSON       ┌────────────────────────────────┐
│   Vue 3 SPA      │  ──────────────────→  │        Gin BFF (:8080)         │
│   Vite :5173     │  ←─────────────────   │  JWT auth · CORS · Tracing     │
└──────────────────┘                       └──────────────┬─────────────────┘
                                                           │ gRPC
                                                           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                        Kratos Service Layer                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Order   │  │ Product  │  │   Cart   │  │   User   │  │ Payment  │  │
│  │ :9000    │  │ :9001    │  │ :9002    │  │ :9003    │  │ :9004    │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
│       └──────────────┴─────────────┴─────────────┴─────────────┘       │
│                                   │                                     │
│                              ┌─────┴──────┐                             │
│                              │   MySQL    │                             │
│                              └────────────┘                             │
└──────────────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
myidea/
├── go.work                  # Go workspace (api + bff + service)
├── Makefile                 # proto generate, mod tidy, build
├── api/                     # Shared protocol layer (protobuf + generated stubs)
│   ├── order/v1/           # Order service contract
│   ├── product/v1/         # Product service contract
│   ├── cart/v1/            # Cart service contract
│   ├── user/v1/            # User service contract
│   └── payment/v1/         # Payment service contract
├── service/                 # Kratos microservice (mono-repo)
│   ├── cmd/               # Entry points — one per gRPC service
│   │   ├── order/         #   Order service (:9000)
│   │   ├── product/       #   Product service (:9001)
│   │   ├── cart/          #   Cart service (:9002)
│   │   ├── user/          #   User service (:9003)
│   │   └── payment/       #   Payment service (:9004)
│   └── internal/
│       ├── conf/           # Configuration (environment variables)
│       ├── biz/            # Business logic layer (domain models + usecases)
│       ├── data/           # Data/persistence layer (GORM)
│       └── service/        # gRPC service implementations
├── bff/                    # BFF API gateway (Gin)
│   ├── main.go             # Entry point with route registration
│   ├── config/             # Configuration
│   ├── handler/            # HTTP handlers (order, product, cart, user, payment)
│   └── middleware/         # Auth (JWT), CORS, tracing middleware
└── web/                    # Vue 3 SPA frontend
    ├── index.html
    ├── package.json
    ├── vite.config.js
    └── src/
        ├── main.js         # App entry
        ├── App.vue         # Root component with navbar
        ├── api/            # Axios API client + per-module API functions
        ├── stores/         # Pinia stores (auth, cart)
        ├── router/         # Vue Router with auth guards
        ├── views/          # 10 page components
        └── assets/         # Global CSS
```

## Features

- **User** — Register, login, JWT authentication, address management
- **Product** — SPU/SKU catalog, categories, brands, inventory with stock locking
- **Cart** — Add/remove items, stock validation, checkout flow
- **Order** — Order creation, status state machine, cancellation
- **Payment** — Mock payment processing, provider abstraction for 3rd-party SDKs
- **Frontend** — Vue 3 SPA with full CRUD for all business modules
- **Cross-cutting** — OpenTelemetry tracing, unified error handling, request timeout

## Tech Stack

| Component               | Technology                            |
|-------------------------|---------------------------------------|
| Language (Backend)      | Go 1.25+                              |
| Service Framework       | Kratos v2                             |
| API Gateway (BFF)       | Gin                                   |
| RPC Protocol            | gRPC + Protobuf                       |
| ORM                     | GORM                                  |
| Database                | MySQL 8.0+                            |
| Auth                    | JWT (golang-jwt/v5)                   |
| Tracing                 | OpenTelemetry                         |
| Workspace               | Go workspace (3 modules)              |
| Frontend                | Vue 3 + Composition API               |
| Build Tool              | Vite                                  |
| State Management        | Pinia                                 |
| Routing                 | Vue Router                            |
| HTTP Client             | Axios                                 |

## Prerequisites

| Tool                         | Purpose                   |
|------------------------------|---------------------------|
| Go 1.25+                     | Build & run backend       |
| Node.js 20+                  | Build & run frontend      |
| MySQL 8.0+                   | Data storage              |
| `protoc` + protoc-gen plugins| Proto code generation     |

## Quick Start

### 1. Start MySQL

```bash
docker run -d \
  --name mall-mysql \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=mall_order \
  -p 3306:3306 \
  mysql:8.0
```

### 2. Start gRPC Services

```bash
cd service
export DB_DSN="root:password@tcp(127.0.0.1:3306)/mall_order?charset=utf8mb4&parseTime=True&loc=Local"

# Start each service in a separate terminal (or use & for background):
go run ./cmd/order/     # :9000
go run ./cmd/product/   # :9001
go run ./cmd/cart/      # :9002
go run ./cmd/user/      # :9003
go run ./cmd/payment/   # :9004
```

5 independent gRPC servers run on `:9000`–`:9004`. `AutoMigrate` creates all tables automatically.

### 3. Start BFF Gateway

```bash
cd bff
export JWT_SECRET="my-secret-key"
go run .
```

BFF listens on `:8080`.

### 4. Start Frontend

```bash
cd web
npm install
npm run dev
```

Frontend dev server listens on `:5173`, proxying `/api` requests to the BFF at `:8080`.

Open http://localhost:5173 in your browser.

## API Endpoints

### Auth (No JWT)

| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/health` | Health check |
| `POST` | `/api/v1/auth/register` | User registration |
| `POST` | `/api/v1/auth/login` | User login |
| `POST` | `/api/v1/payments/:id/notify` | Payment webhook (3rd-party callback) |

### Product (JWT Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/api/v1/categories` | Category tree |
| `POST` | `/api/v1/categories` | Create category |
| `PUT`  | `/api/v1/categories/:id` | Update category |
| `DELETE` | `/api/v1/categories/:id` | Delete category (no children) |
| `GET`  | `/api/v1/brands` | List brands (`?keyword=&page=&page_size=`) |
| `POST` | `/api/v1/brands` | Create brand |
| `GET`  | `/api/v1/spus` | List SPUs (`?category_id=&brand_id=&keyword=&page=&page_size=`) |
| `GET`  | `/api/v1/spus/:id` | SPU detail (with SKUs) |
| `POST` | `/api/v1/spus` | Create SPU |
| `PUT`  | `/api/v1/spus/:id` | Update SPU |
| `PUT`  | `/api/v1/spus/:id/status` | Online/offline toggle |
| `GET`  | `/api/v1/skus` | List SKUs (`?spu_id=`) |
| `POST` | `/api/v1/skus/batch` | Batch create SKUs |
| `PUT`  | `/api/v1/skus/:id` | Update SKU |

### Cart (JWT Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/api/v1/cart` | List cart items |
| `POST` | `/api/v1/cart/items` | Add item (upsert by SKU) |
| `PUT`  | `/api/v1/cart/items/:id` | Update quantity |
| `DELETE` | `/api/v1/cart/items/:id` | Remove item |
| `POST` | `/api/v1/cart/checkout` | Checkout (Cart → Order → Payment) |

### Order (JWT Required)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/orders` | Create order |
| `GET`  | `/api/v1/orders` | List orders (`?status=&page=&page_size=`) |
| `GET`  | `/api/v1/orders/:id` | Get order detail |
| `PUT`  | `/api/v1/orders/:id/status` | Update order status |
| `POST` | `/api/v1/orders/:id/cancel` | Cancel order |

### Address (JWT Required)

| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/api/v1/addresses` | List addresses |
| `POST` | `/api/v1/addresses` | Create address (max 10) |
| `PUT`  | `/api/v1/addresses/:id` | Update address |
| `DELETE` | `/api/v1/addresses/:id` | Delete address |
| `PUT`  | `/api/v1/addresses/:id/default` | Set as default |

### Payment (JWT Required)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/payments` | Create payment |
| `POST` | `/api/v1/payments/:id/process` | Process payment |
| `GET`  | `/api/v1/payments/:id` | Get payment by ID |
| `GET`  | `/api/v1/payments/by-order/:orderNo` | Get payment by order |

## Port Reference

| Service        | Protocol | Port  |
|----------------|----------|-------|
| Order gRPC     | gRPC     | :9000 |
| Product gRPC   | gRPC     | :9001 |
| Cart gRPC      | gRPC     | :9002 |
| User gRPC      | gRPC     | :9003 |
| Payment gRPC   | gRPC     | :9004 |
| BFF HTTP       | HTTP     | :8080 |
| Vite Dev Server| HTTP     | :5173 |

## Development

### Generate protobuf stubs

```bash
make proto
```

### Build all backend modules

```bash
make build
```

### Tidy dependencies

```bash
make tidy
```

### Build individual modules

```bash
go build ./service/...
go build ./bff/...
```

### Build frontend for production

```bash
cd web
npm run build
```

Output is in `web/dist/`.

## License

MIT
