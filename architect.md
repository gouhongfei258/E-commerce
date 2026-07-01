# Mall E-Commerce Microservice — Architecture & Development Guide

## Project Root

```
/Users/storm/githubproject/myidea/
```

## Compilation Status

All Go modules pass `go build ./...` and `go vet ./...`. Vue frontend builds with `npm run build`.

- `service` — Order gRPC `:9000`, Product gRPC `:9001`, Cart gRPC `:9002`, User gRPC `:9003`, Payment gRPC `:9004`
- `bff` — Gin HTTP `:8080`
- `web` — Vue 3 SPA (Vite dev server `:5173`, production build → `web/dist/`)

---

## 1. Overall Architecture

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

### Layer Responsibilities

| Layer | Tech | Responsibilities |
|-------|------|-----------------|
| **Frontend (SPA)** | Vue 3 + Vite | UI rendering, client-side routing, state management (Pinia), JWT storage |
| **BFF (API Gateway)** | Gin | Route dispatch, JWT auth (incl. generation on login/register), request validation, unified HTTP JSON response, gRPC client calls |
| **Kratos Service** | Kratos v2 + GORM | DDD layers: `service` (gRPC), `biz` (domain), `data` (persistence) |

### Cross-cutting

- **Trace propagation**: OpenTelemetry trace context is injected via Gin middleware (see `bff/middleware/tracer.go`). When the gRPC client uses `otelgrpc`, the TraceID propagates transparently through gRPC metadata.
- **User identity propagation**: BFF extracts `user_id` from JWT, injects it into gRPC outgoing metadata as `x-user-id`. Kratos service can extract it from incoming metadata.
- **Error conversion**: Kratos errors (`errors.NotFound`, `errors.Forbidden`, `errors.BadRequest`) are serialized as gRPC status errors. The BFF's `respondError()` converts them to the unified JSON format: `{"code": <gRPC_code>, "msg": <message>, "data": null}`.
- **Context timeout**: BFF middleware sets a 10s per-request timeout on `context.Context`. The gRPC layer propagates this to the service.
- **Frontend API flow**: Vue SPA stores JWT in `localStorage`. Axios interceptor attaches `Authorization: Bearer <token>` to all requests. Router guard redirects to `/login` on 401.

---

## 2. Project Directory Structure

```
myidea/
├── go.work                              # Go workspace (api + bff + service)
├── Makefile                             # proto generate + mod tidy + build
│
├── api/                                 # Shared protocol layer
│   ├── go.mod
│   ├── order/v1/
│   │   ├── order.proto                  # Order service contract
│   │   ├── order.pb.go
│   │   └── order_grpc.pb.go
│   ├── product/v1/
│   │   ├── product.proto                # Product service contract
│   │   ├── product.pb.go
│   │   └── product_grpc.pb.go
│   ├── cart/v1/
│   │   ├── cart.proto                   # Cart service contract
│   │   ├── cart.pb.go
│   │   └── cart_grpc.pb.go
│   └── user/v1/
│       ├── user.proto                   # User service contract
│       ├── user.pb.go
│       └── user_grpc.pb.go
│   └── payment/v1/
│       ├── payment.proto                # Payment service contract
│       ├── payment.pb.go
│       └── payment_grpc.pb.go
│
├── service/                             # Kratos microservices (mono-repo)
│   ├── go.mod
│   ├── cmd/
│   │   ├── order/main.go               # Order gRPC server (:9000)
│   │   ├── product/main.go             # Product gRPC server (:9001)
│   │   ├── cart/main.go                # Cart gRPC server (:9002)
│   │   ├── user/main.go                # User gRPC server (:9003)
│   │   └── payment/main.go             # Payment gRPC server (:9004)
│   └── internal/
│       ├── conf/conf.go                  # Config (env vars)
│       ├── biz/                          # Business logic layer
│       │   ├── errors.go                # Domain error codes (all services + payment)
│       │   ├── order.go                 # Order domain model + usecase
│       │   ├── category.go              # Category domain model + usecase
│       │   ├── brand.go                 # Brand domain model + usecase
│       │   ├── spu.go                   # SPU domain model + usecase
│       │   ├── sku.go                   # SKU domain model + usecase (stock)
│       │   ├── cart.go                  # CartItem domain model + usecase
│       │   ├── user.go                  # User domain model + usecase
│       │   ├── address.go               # Address domain model + usecase
│       │   └── payment.go               # Payment domain model + Provider interface + usecase
│       ├── data/                        # Data/persistence layer
│       │   ├── data.go                  # GORM connection pool + auto-migrate
│       │   ├── order.go                 # OrderRepository GORM impl
│       │   ├── category.go              # CategoryRepository GORM impl
│       │   ├── brand.go                 # BrandRepository GORM impl
│       │   ├── spu.go                   # SPURepository GORM impl
│       │   ├── sku.go                   # SKURepository GORM impl (stock journal)
│       │   ├── cart.go                  # CartRepository GORM impl
│       │   ├── user.go                  # UserRepository GORM impl
│       │   ├── address.go               # AddressRepository GORM impl
│       │   ├── payment.go               # PaymentRepository GORM impl + MockPaymentProvider
│       │   └── product_client.go        # gRPC ProductStockClient (Order → Product stock RPCs)
│       └── service/                     # gRPC transport layer
│           ├── order.go                 # Order gRPC server impl
│           ├── product.go               # Product gRPC server impl
│           ├── cart.go                  # Cart gRPC server impl
│           ├── user.go                  # User gRPC server impl
│           └── payment.go               # Payment gRPC server impl
│
├── bff/                                 # Gin API Gateway
│   ├── go.mod
│   ├── main.go                          # Entry: Gin engine + route registration
│   ├── config/config.go                 # Config (env vars)
│   ├── middleware/
│   │   ├── auth.go                      # JWT bearer token validation
│   │   ├── cors.go                      # CORS headers + request timeout
│   │   └── tracer.go                    # OpenTelemetry trace injection
│   └── handler/
│       ├── order.go                     # Order HTTP handlers + gRPC client
│       ├── product.go                   # Product HTTP handlers + gRPC client
│       ├── cart.go                      # Cart HTTP handlers + gRPC client (incl. checkout)
│       ├── user.go                      # User HTTP handlers + gRPC client (auth + address)
│       └── payment.go                   # Payment HTTP handlers + gRPC client
│
└── web/                                 # Vue 3 SPA Frontend
    ├── index.html
    ├── package.json
    ├── vite.config.js                   # Vite config, API proxy to :8080
    └── src/
        ├── main.js                      # App entry, Pinia + Router setup
        ├── App.vue                      # Root component with conditional navbar
        ├── api/                         # API client layer
        │   ├── http.js                  # Axios instance, JWT interceptor, 401 redirect
        │   ├── auth.js                  # login(), register()
        │   ├── product.js               # categories, brands, spus, skus
        │   ├── cart.js                  # cart CRUD + checkout
        │   ├── order.js                 # order CRUD + cancel
        │   ├── payment.js               # payment operations
        │   └── address.js               # address CRUD
        ├── stores/                      # Pinia state management
        │   ├── auth.js                  # Auth state (token, user, login/logout)
        │   └── cart.js                  # Cart state (items, totals, CRUD)
        ├── router/
        │   └── index.js                 # 10 routes with auth guards
        ├── views/                       # Page components
        │   ├── Login.vue                # Login form
        │   ├── Register.vue             # Registration form
        │   ├── ProductList.vue          # SPU grid with filters + pagination
        │   ├── ProductDetail.vue        # SPU detail with SKU selection + add-to-cart
        │   ├── Cart.vue                 # Cart items, qty adjust, checkout
        │   ├── Checkout.vue             # Address selection, order placement
        │   ├── OrderList.vue            # Order history with status filter
        │   ├── OrderDetail.vue          # Order detail + payment + cancel
        │   ├── AddressList.vue          # Address CRUD with inline form
        │   └── Profile.vue              # User info display
        └── assets/
            └── style.css                # Global CSS variables & base styles
```

---

## 3. Vue Frontend Architecture

### 3.1 Tech Stack

| Component | Library | Purpose |
|-----------|---------|---------|
| Framework | Vue 3 (Composition API, `<script setup>`) | UI components |
| Build Tool | Vite | Dev server + production bundling |
| Routing | Vue Router 4 | SPA routing with navigation guards |
| State Mgmt | Pinia 2 | Auth state, cart state |
| HTTP Client | Axios | API calls with JWT interceptor |
| Styling | Plain CSS + CSS variables | Lightweight, no UI framework |

### 3.2 Auth Flow

1. User registers or logs in → `POST /api/v1/auth/login` (or `/register`)
2. Response: `{ code: 0, msg: "ok", data: { user: {...}, token: "..." } }`
3. Token stored in `localStorage` + Pinia auth store
4. Axios request interceptor automatically attaches `Authorization: Bearer <token>`
5. Axios response interceptor catches 401 → clears auth → redirects to `/login`
6. Router `beforeEach` guard redirects unauthenticated users to `/login`

### 3.3 Routes

| Path | View | Auth Required | Description |
|------|------|:---:|---|
| `/login` | Login | No | User login |
| `/register` | Register | No | User registration |
| `/` | ProductList | Yes | Browse SPUs with category/brand/keyword filters |
| `/products/:id` | ProductDetail | Yes | SPU detail with SKU selection |
| `/cart` | Cart | Yes | Cart management with qty adjust |
| `/checkout` | Checkout | Yes | Address selection + order placement |
| `/orders` | OrderList | Yes | Order history with status tabs |
| `/orders/:id` | OrderDetail | Yes | Order detail + payment + cancel |
| `/addresses` | AddressList | Yes | Address CRUD |
| `/profile` | Profile | Yes | User info display |

### 3.4 Key Design Decisions

- **No heavy UI framework**: Plain CSS with CSS variables (soft blue `#4A90D9`, clean typography, card-based layout). Keeps bundle small.
- **Lazy-loaded routes**: Each view is code-split via dynamic `import()` in the router, reducing initial load time.
- **Cart state persistence**: Cart store fetches fresh data from BFF on mount. No local caching — cart state is server-authoritative.
- **Response envelope unwrapping**: Axios interceptor unwraps `{ code, msg, data }` — components receive `data` directly and only handle the `code !== 0` error case.
- **Pagination**: Product list and order list use offset-based pagination (`page` / `page_size` params) with Previous/Next controls.

### 3.5 API Client Layer

Each BFF module has a corresponding API file in `src/api/`:

| Module | File | Functions |
|--------|------|-----------|
| Auth | `auth.js` | `login()`, `register()` |
| Product | `product.js` | `getCategories()`, `listBrands()`, `listSPUs()`, `getSPU()`, `listSKUs()` |
| Cart | `cart.js` | `listCartItems()`, `addCartItem()`, `updateCartItem()`, `removeCartItem()`, `checkout()` |
| Order | `order.js` | `createOrder()`, `listOrders()`, `getOrder()`, `cancelOrder()` |
| Payment | `payment.js` | `createPayment()`, `processPayment()`, `getPayment()`, `getPaymentByOrder()` |
| Address | `address.js` | `listAddresses()`, `createAddress()`, `updateAddress()`, `deleteAddress()`, `setDefaultAddress()` |

All functions use the shared Axios instance from `http.js`, which handles JWT injection and response unwrapping.

---

## 4. Backend Architecture (Legacy Documentation)

### 4.1 Order Service — Protobuf Contract (`api/order/v1/order.proto`)

```protobuf
service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
  rpc UpdateOrderStatus(UpdateOrderStatusRequest) returns (UpdateOrderStatusResponse);
  rpc CancelOrder(CancelOrderRequest) returns (CancelOrderResponse);
}
```

Messages: `OrderProto`, `OrderItemProto` (with `sku_id`), `ShippingAddressProto`

### 4.2 Order Service — Kratos Implementation

#### biz layer (`service/internal/biz/`)

- **`order.go`** — Domain model (aggregate root) with behavior:
  - `OrderStatus` enum with `CanTransitionTo()` state machine:
    ```
    Pending → Paid, Cancelled
    Paid → Shipped, Refunding
    Shipped → Delivered
    Delivered → Refunding
    Refunding → Refunded
    Cancelled/Refunded → (terminal)
    ```
  - `ShippingAddress`, `OrderItem` — value objects (`OrderItem` includes `SKUID` for stock locking)
  - `Order` aggregate with `CalculateTotal()`
  - `OrderRepository` interface (persistence contract)
  - `ProductStockClient` interface (calls Product Service's `LockStock`/`ConfirmDeductStock`/`UnlockStock` via gRPC)
  - `OrderBiz` usecase struct: `CreateOrder`, `GetOrder`, `ListOrders`, `UpdateOrderStatus`, `CancelOrder`
  - **`CreateOrder`** — locks stock for each SKU via `ProductStockClient` before saving; unlocks all on failure (rollback)
  - **`UpdateOrderStatus`** — confirms stock deduction (`ConfirmDeductStock`) when transitioning to `Paid`
  - **`CancelOrder`** — unlocks stock (`UnlockStock`) for each item when cancelling a pending order

- **`errors.go`** — 6 domain errors using Kratos' error factory:
  ```go
  ErrOrderNotFound       = errors.NotFound("ORDER_NOT_FOUND", "order not found")
  ErrOrderStatusInvalid  = errors.Forbidden("ORDER_STATUS_INVALID", "...")
  ErrOrderCannotCancel   = errors.Forbidden("ORDER_CANNOT_CANCEL", "...")
  ErrOrderItemEmpty      = errors.BadRequest("ORDER_ITEM_EMPTY", "...")
  ErrOrderAddressRequired = errors.BadRequest("ORDER_ADDRESS_REQUIRED", "...")
  ErrOrderUnauthorized   = errors.Forbidden("ORDER_UNAUTHORIZED", "...")
  ```

#### data layer (`service/internal/data/`)

- **`data.go`** — `Data` struct holding `*gorm.DB`. `NewData()`:
  - Opens MySQL via GORM with connection pooling (MaxIdleConns=10, MaxOpenConns=100, ConnMaxLifetime=30m)
  - Runs `AutoMigrate` for all GORM models
  - Exposes `DB(ctx)` method for context propagation

- **`order.go`** — `orderRepo` implementing `biz.OrderRepository`:
  - GORM model: `GORMOrder` (table `orders`), `GORMOrderItem` (table `order_items`, FK cascade delete, includes `sku_id` column)
  - Transaction for `Save()` (order + items atomic)
  - `Preload("Items")` for read queries
  - Error mapping: `gorm.ErrRecordNotFound` → `biz.ErrOrderNotFound`
  - Paginated `List()` with count + offset/limit
  - `toGORM()` / `toDomain()` mapping helpers

- **`product_client.go`** — gRPC `ProductStockClient` implementation:
  - Dials Product Service (`:9001`) at startup
  - `LockStock` — calls `ProductService.LockStock` RPC for a single SKU
  - `ConfirmDeductStock` — calls `ProductService.ConfirmDeductStock` RPC
  - `UnlockStock` — calls `ProductService.UnlockStock` RPC
  - Idempotency is handled by Product Service's stock journal unique index

#### service layer (`service/internal/service/`)

- **`order.go`** — `OrderService` implementing `pb.OrderServiceServer`:
  - Proto → domain conversion in request handlers
  - Domain → proto conversion via `orderToProto()`
  - Error passthrough (Kratos errors travel through gRPC automatically)

#### cmd/order/main.go

- Manual DI chain: `conf → data → biz → service → grpc.Server`
- `OrderBiz` receives a `ProductStockClient` implementation that dials Product Service (`:9001`) via gRPC
- Graceful shutdown via `signal.Notify`

### 4.3 Order Service — BFF Implementation

#### handler/order.go

- `OrderHandler` holds `pb.OrderServiceClient` gRPC client
- `NewOrderHandler()` dials gRPC with 5s timeout, insecure credentials
- 5 HTTP handler methods with Gin's `ShouldBindJSON` validation
- `injectUserID()` propagates `user_id` via gRPC metadata.Headers
- `respond()` — unified JSON envelope: `{"code","msg","data"}`
- `respondError()` — `status.FromError()` → `grpcCodeToHTTP()` mapping
- `OrderServiceClient()` — exposes gRPC client for Cart checkout orchestration

#### middleware/

- **`auth.go`** — `JWTAuth(secret)` Gin middleware:
  - Extracts Bearer token from Authorization header
  - Validates HMAC-SHA256 signature
  - Injects `user_id` claim into `gin.Context`
  - Returns 401 JSON on any failure

- **`cors.go`** — CORS headers + `Timeout()` middleware wrapping `context.WithTimeout`

- **`tracer.go`** — `Trace()` middleware:
  - Extracts incoming trace context from HTTP headers
  - Creates a root span per request
  - Injects trace context back into outgoing headers
  - Sets 10s timeout context

### 4.4 BFF API Endpoints (Order)

All `/api/v1/*` routes require `Authorization: Bearer <JWT>` with `user_id` claim (except auth routes and health check).

| Method | Path | Auth | Description |
|--------|------|:---:|-------------|
| POST | `/api/v1/orders` | JWT | Create order (items + address) |
| GET | `/api/v1/orders` | JWT | List orders (`?status=&page=&page_size=`) |
| GET | `/api/v1/orders/:id` | JWT | Get single order |
| PUT | `/api/v1/orders/:id/status` | JWT | Update order status |
| POST | `/api/v1/orders/:id/cancel` | JWT | Cancel order |
| GET | `/health` | No | Health check |

### 4.5 Configuration (env vars)

| Variable | Default | Module |
|----------|---------|--------|
| `SERVICE_NAME` | `order-service` | service |
| `GRPC_ADDR` | `:9000` | service |
| `PRODUCT_GRPC_ADDR` | `:9001` | service |
| `CART_GRPC_ADDR` | `:9002` | service |
| `USER_GRPC_ADDR` | `:9003` | service |
| `PAYMENT_GRPC_ADDR` | `:9004` | service |
| `DB_DSN` | `root:password@tcp(127.0.0.1:3306)/mall_order?...` | service |
| `BFF_ADDR` | `:8080` | bff |
| `ORDER_SERVICE_ADDR` | `127.0.0.1:9000` | bff |
| `PRODUCT_SERVICE_ADDR` | `127.0.0.1:9001` | bff |
| `CART_SERVICE_ADDR` | `127.0.0.1:9002` | bff |
| `USER_SERVICE_ADDR` | `127.0.0.1:9003` | bff |
| `PAYMENT_SERVICE_ADDR` | `127.0.0.1:9004` | bff |
| `JWT_SECRET` | `change-me-in-production` | bff |

---

### 4.6 Product Service

#### Protobuf Contract (`api/product/v1/product.proto`)

```protobuf
service ProductService {
  // Category
  rpc CreateCategory(CreateCategoryRequest) returns (CategoryResponse);
  rpc UpdateCategory(UpdateCategoryRequest) returns (CategoryResponse);
  rpc DeleteCategory(DeleteCategoryRequest) returns (Empty);
  rpc GetCategoryTree(GetCategoryTreeRequest) returns (GetCategoryTreeResponse);

  // Brand
  rpc CreateBrand(CreateBrandRequest) returns (BrandResponse);
  rpc ListBrands(ListBrandsRequest) returns (ListBrandsResponse);

  // SPU
  rpc CreateSPU(CreateSPURequest) returns (SPUResponse);
  rpc UpdateSPU(UpdateSPURequest) returns (SPUResponse);
  rpc GetSPU(GetSPURequest) returns (SPUResponse);
  rpc ListSPUs(ListSPUsRequest) returns (ListSPUsResponse);

  // SKU
  rpc BatchCreateSKU(BatchCreateSKURequest) returns (BatchCreateSKUResponse);
  rpc UpdateSKU(UpdateSKURequest) returns (SKUResponse);
  rpc ListSKUs(ListSKUsRequest) returns (ListSKUsResponse);

  // Stock (核心，供订单服务调用)
  rpc LockStock(LockStockRequest) returns (LockStockResponse);
  rpc ConfirmDeductStock(ConfirmDeductRequest) returns (Empty);
  rpc UnlockStock(UnlockStockRequest) returns (Empty);
}
```

#### biz layer (`service/internal/biz/`)

- **`category.go`** — `Category` with `IsRoot()`, tree builder, 3-level depth validation
- **`brand.go`** — `Brand` with unique `Name`
- **`spu.go`** — `SPU` with `SPUStatus` state machine (Offline↔Online↔SoldOut)
- **`sku.go`** — `SKU` with `AvailableStock()`, optimistic-lock stock operations, idempotent journal
- **`errors.go`** — 10 product errors alongside order errors
- **`CategoryBiz`** — `CreateCategory`, `UpdateCategory`, `DeleteCategory` (with children check), `GetCategoryTree`
- **`BrandBiz`** — `CreateBrand`, `ListBrands` (paginated + keyword search)
- **`SPUBiz`** — `CreateSPU`, `UpdateSPU`, `GetSPU`, `ListSPUs` (with filters)
- **`SKUBiz`** — `BatchCreateSKU`, `UpdateSKU`, `ListSKUs`, `LockStock`/`ConfirmDeductStock`/`UnlockStock`

#### data layer (`service/internal/data/`)

- **`data.go`** — AutoMigrate includes all 5 product GORM models
- **`category.go`** — `GORMCategory` (table `categories`), tree query by `parent_id`, children count
- **`brand.go`** — `GORMBrand` (table `brands`, unique `name`), paginated list
- **`spu.go`** — `GORMSPU` (table `spus`), JSON serialization for `saleable_attr_names`, multi-filter list
- **`sku.go`** — `GORMSKU` (table `skus`, JSON `attrs`, optimistic lock `version`)
  - `GORMStockJournal` (table `stock_journals`, composite unique index `uk_order_sku_type`)
  - `LockStock`: `UPDATE ... SET stock=stock-N, locked_stock=locked_stock+N, version=version+1 WHERE id=? AND version=? AND stock>=N`
  - `ConfirmDeduct`: `UPDATE ... SET locked_stock=locked_stock-N WHERE id=? AND locked_stock>=N`
  - `UnlockStock`: `UPDATE ... SET stock=stock+N, locked_stock=locked_stock-N, version=version+1 WHERE id=? AND locked_stock>=N`
  - `CreateJournal`: check-then-insert with unique index as idempotency safety net

#### service layer (`service/internal/service/`)

- **`product.go`** — `ProductService` implementing `pb.ProductServiceServer`
  - Pure proto↔domain mapping, no business logic
  - `categoryToProto` / `brandToProto` / `spuToProto` / `skuToProto` helpers

#### BFF handler (`bff/handler/product.go`)

- `ProductHandler` with `pb.ProductServiceClient`
- HTTP endpoints for categories, brands, SPUs, SKUs

#### BFF API Endpoints

| Method | Path | Auth | Description |
|--------|------|:---:|-------------|
| GET | `/api/v1/categories` | JWT | Category tree |
| POST | `/api/v1/categories` | JWT | Create category |
| PUT | `/api/v1/categories/:id` | JWT | Update category |
| DELETE | `/api/v1/categories/:id` | JWT | Delete category (no children) |
| GET | `/api/v1/brands` | JWT | List brands (`?keyword=&page=&page_size=`) |
| POST | `/api/v1/brands` | JWT | Create brand |
| GET | `/api/v1/spus` | JWT | List SPUs (`?category_id=&brand_id=&keyword=&page=&page_size=`) |
| GET | `/api/v1/spus/:id` | JWT | SPU detail (with SKUs) |
| POST | `/api/v1/spus` | JWT | Create SPU |
| PUT | `/api/v1/spus/:id` | JWT | Update SPU |
| PUT | `/api/v1/spus/:id/status` | JWT | Online/offline toggle |
| GET | `/api/v1/skus` | JWT | List SKUs (`?spu_id=`) |
| POST | `/api/v1/skus/batch` | JWT | Batch create SKUs |
| PUT | `/api/v1/skus/:id` | JWT | Update SKU |

#### Service Dependencies

Product Service is an **independent** Kratos microservice. It does NOT call Order Service. The dependency direction is reversed:

```
Order Service → (gRPC) Product Service
   LockStock(ctx, sku_id, quantity, order_no)
   ConfirmDeductStock(ctx, sku_id, quantity, order_no)
   UnlockStock(ctx, sku_id, quantity, order_no)
```

---

### 4.7 Cart Service

#### Protobuf Contract (`api/cart/v1/cart.proto`)

```protobuf
service CartService {
  rpc AddItem(AddItemRequest) returns (CartItemProto);
  rpc UpdateQuantity(UpdateQuantityRequest) returns (CartItemProto);
  rpc RemoveItem(RemoveItemRequest) returns (Empty);
  rpc ListItems(ListItemsRequest) returns (ListItemsResponse);
  rpc ClearItems(ClearItemsRequest) returns (Empty);
}
```

Error codes: `CART_ITEM_NOT_FOUND`, `CART_QUANTITY_INVALID`

#### biz layer (`service/internal/biz/cart.go`)

- **Domain model**: `CartItem` with `SubTotal()`, `Attrs` (map[string]string for SKU attribute snapshot)
- **CartRepository interface**: `AddItem`, `UpdateQuantity`, `RemoveItem`, `ListItems`, `ClearItems`
- **CartBiz usecase**: validates quantity > 0, delegates to repo

#### data layer (`service/internal/data/cart.go`)

- **`GORMCartItem`** (table `cart_items`) with composite unique index `(user_id, sku_id)`
- **AddItem with upsert**: same user + same SKU → increments quantity; otherwise inserts new row
- **Attrs stored as JSON** in MySQL, serialized via `json.Marshal` / `json.Unmarshal`
- **Delete scoped to user**: every mutation validates `user_id` ownership

#### service layer (`service/internal/service/cart.go`)

- `CartService` implementing `pb.CartServiceServer`
- Pure proto↔domain mapping via `itemToProto()`
- Cart gRPC server registered on `:9002`

#### BFF handler (`bff/handler/cart.go`)

- `CartHandler` holds `pb.CartServiceClient` + `pb.OrderServiceClient` + `pb.PaymentServiceClient` (for checkout)
- **Checkout** orchestration flow:
  ```
  Cart.ListItems → Order.CreateOrder (with LockStock) → Cart.ClearItems
    → Payment.CreatePayment → Payment.ProcessPayment (mock)
      ├── Success → Order.UpdateOrderStatus(Paid) → ConfirmDeductStock
      └── Failed  → Order.CancelOrder → UnlockStock
  ```
- Standard `respond()` / `respondError()` helpers (shared from `order.go`)

#### BFF API Endpoints

| Method | Path | Auth | Description |
|--------|------|:---:|-------------|
| GET | `/api/v1/cart` | JWT | List cart items |
| POST | `/api/v1/cart/items` | JWT | Add item (upserts by SKU) |
| PUT | `/api/v1/cart/items/:id` | JWT | Update quantity |
| DELETE | `/api/v1/cart/items/:id` | JWT | Remove item |
| POST | `/api/v1/cart/checkout` | JWT | Cart → Order → Payment (full checkout with stock locking + mock payment) |

---

### 4.8 User Service

#### Protobuf Contract (`api/user/v1/user.proto`)

```protobuf
service UserService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc GetUser(GetUserRequest) returns (UserProto);
  rpc CreateAddress(CreateAddressRequest) returns (AddressProto);
  rpc UpdateAddress(UpdateAddressRequest) returns (AddressProto);
  rpc DeleteAddress(DeleteAddressRequest) returns (Empty);
  rpc ListAddresses(ListAddressesRequest) returns (ListAddressesResponse);
  rpc SetDefaultAddress(SetDefaultAddressRequest) returns (AddressProto);
}
```

Messages: `UserProto`, `AddressProto`, `Empty`

Error codes: `USER_NOT_FOUND`, `USER_DUPLICATE`, `USER_PASSWORD_WRONG`, `ADDRESS_NOT_FOUND`, `ADDRESS_LIMIT`

#### biz layer (`service/internal/biz/`)

- **`user.go`** — Domain model with behavior:
  - `User` aggregate with `Username` (unique), bcrypt `Password`, `Phone`, `Email`
  - `HashPassword()` / `CheckPassword()` — bcrypt hashing and verification
  - `UserRepository` interface: `Create`, `FindByUsername`, `FindByID`
  - `UserBiz` usecase struct:
    - `Register(ctx, username, password, phone, email)` — validates uniqueness → bcrypt hash → create
    - `Login(ctx, username, password)` — find by username → bcrypt verify
    - `GetUser(ctx, id)` — find by ID

- **`address.go`** — Address domain model:
  - `Address` value object with `ReceiverName`, `ReceiverPhone`, `Province`, `City`, `District`, `DetailAddress`, `IsDefault`
  - `AddressRepository` interface: `Create`, `Update`, `Delete`, `ListByUserID`, `SetDefault`, `CountByUserID`
  - `AddressBiz` usecase struct:
    - `Create` — enforces max 10 addresses per user, handles default uniqueness
    - `Update` / `Delete` — scoped to user, returns `ErrAddressNotFound` on mismatch
    - `SetDefault` — unsets previous default atomically
    - `ListByUserID` — defaults first, then by creation time desc

- **`errors.go`** — 5 user/address domain errors:
  ```go
  ErrUserNotFound      = errors.NotFound("USER_NOT_FOUND", "user not found")
  ErrUserDuplicate     = errors.Forbidden("USER_DUPLICATE", "username already exists")
  ErrUserPasswordWrong = errors.Forbidden("USER_PASSWORD_WRONG", "invalid password")
  ErrAddressNotFound   = errors.NotFound("ADDRESS_NOT_FOUND", "address not found")
  ErrAddressLimit      = errors.Forbidden("ADDRESS_LIMIT", "maximum 10 addresses per user")
  ```

#### data layer (`service/internal/data/`)

- **`user.go`** — `GORMUser` (table `users`):
  - `username` unique index, standard timestamp fields
  - `userRepo` implementing `biz.UserRepository`
  - Not-found returns `nil, nil` (caller distinguishes empty vs error)

- **`address.go`** — `GORMAddress` (table `addresses`):
  - `user_id` index for efficient user-scoped queries
  - `addressRepo` implementing `biz.AddressRepository`
  - `Update`/`Delete` check `RowsAffected` → `ErrAddressNotFound` on miss
  - `SetDefault` clears all defaults first, then sets the target if id > 0
  - `ListByUserID` ordered by `is_default DESC, created_at DESC`

#### service layer (`service/internal/service/user.go`)

- `UserService` implementing `pb.UserServiceServer`
- Pure proto↔domain mapping, no business logic
- `userToProto()` / `addressToProto()` helpers

#### cmd/user/main.go

- User gRPC server registered on `:9003` with reflection
- Full DI chain: `userRepo → userBiz`, `addressRepo → addressBiz` → `userSvc`
- Graceful shutdown includes user gRPC server

#### BFF handler (`bff/handler/user.go`)

- `UserHandler` holds `pb.UserServiceClient` + `jwtSecret` + `jwtExpiry`
- **`Register`** — calls `UserService.Register` → generates JWT with `user_id` claim → returns `{user, token}`
- **`Login`** — calls `UserService.Login` → generates JWT → returns `{user, token}`
- **Address CRUD** — injects `user_id` from JWT context, calls corresponding gRPC methods
- JWT generation uses the same `golang-jwt` library as middleware

#### BFF API Endpoints

| Method | Path | Auth | Description |
|--------|------|:---:|-------------|
| POST | `/api/v1/auth/register` | No | Register, returns JWT |
| POST | `/api/v1/auth/login` | No | Login, returns JWT |
| GET | `/api/v1/addresses` | JWT | List user's addresses |
| POST | `/api/v1/addresses` | JWT | Add new address (max 10) |
| PUT | `/api/v1/addresses/:id` | JWT | Update address |
| DELETE | `/api/v1/addresses/:id` | JWT | Delete address |
| PUT | `/api/v1/addresses/:id/default` | JWT | Set as default address |

---

### 4.9 Payment Service

#### Protobuf Contract (`api/payment/v1/payment.proto`)

```protobuf
service PaymentService {
  rpc CreatePayment(CreatePaymentRequest) returns (CreatePaymentResponse);
  rpc ProcessPayment(ProcessPaymentRequest) returns (ProcessPaymentResponse);
  rpc GetPayment(GetPaymentRequest) returns (GetPaymentResponse);
  rpc GetPaymentByOrder(GetPaymentByOrderRequest) returns (GetPaymentByOrderResponse);
  rpc NotifyPayment(NotifyPaymentRequest) returns (NotifyPaymentResponse);
}
```

Payment status: `PENDING(1)`, `SUCCESS(2)`, `FAILED(3)`, `REFUNDING(4)`, `REFUNDED(5)`
Payment method: `MOCK(1)`, `ALIPAY(2)`, `WECHAT_PAY(3)` — ALIPAY/WECHAT_PAY are reserved for third-party SDK integration.

#### Third-party SDK Extension Design

**Core abstraction** — `PaymentProvider` interface (`biz/payment.go`):

```go
type PaymentProvider interface {
    Pay(ctx, req) → (resp, error)
    VerifySignature(ctx, raw) → (bool, error)
    Name() string
}
```

**Registration mechanism** — `DefaultPaymentProviderFactory`:
- `factory.Register(PaymentMethodAlipay, &alipayProvider{})` — register any third-party provider
- `factory.GetProvider(method)` — biz layer retrieves provider by method, fully decoupled

**Built-in mock** — `MockPaymentProvider` (`data/payment.go`):
- 80% success / 20% failure simulation for development and testing

#### biz layer (`service/internal/biz/payment.go`)

- **`PaymentStatus`** enum with `CanTransitionTo()` state machine:
  ```
  Pending → Success, Failed
  Success → Refunding
  Refunding → Refunded
  Failed/Refunded → (terminal)
  ```
- **`PaymentMethod`** enum: `Mock`, `Alipay`, `WechatPay`
- **`Payment`** aggregate root: ID, UserID, OrderNo, TotalAmount, PaidAmount, Status, Method, ProviderTradeNo, ProviderRawResponse, FailReason
- **`PaymentProvider`** interface — the extension point for third-party SDKs (Alipay, WeChat Pay, etc.)
- **`PaymentProviderFactory`** interface + **`DefaultPaymentProviderFactory`** — registration and retrieval of providers
- **`PaymentRepository`** interface: `Save`, `GetByID`, `GetByOrderNo`, `UpdateStatus`
- **`PaymentBiz`** usecase:
  - `CreatePayment` — validates method, creates Pending payment record
  - `ProcessPayment` — retrieves provider via factory → calls `Pay()` → maps result to status update
  - `GetPayment` / `GetPaymentByOrder` — read queries
  - `NotifyPayment` — handles async callbacks from third-party providers with signature verification

Domain errors: `PAYMENT_NOT_FOUND`, `PAYMENT_ALREADY_PROCESSED`, `PAYMENT_STATUS_INVALID`, `PAYMENT_AMOUNT_MISMATCH`, `PAYMENT_PROVIDER_FAIL`, `PAYMENT_INVALID_METHOD`

#### data layer (`service/internal/data/payment.go`)

- **`GORMPayment`** (table `payments`): stores payment records with provider response data
- **`paymentRepo`** implementing `biz.PaymentRepository`: standard GORM CRUD with error mapping
- **`MockPaymentProvider`**: built-in mock implementing `biz.PaymentProvider` for dev/test

#### service layer (`service/internal/service/payment.go`)

- `PaymentService` implementing `pb.PaymentServiceServer`
- Pure proto↔domain mapping via `paymentToProto()`
- Payment gRPC server registered on `:9004`

#### BFF handler (`bff/handler/payment.go`)

- `PaymentHandler` with `pb.PaymentServiceClient`
- 5 HTTP handler methods with Gin's `ShouldBindJSON` validation
- Uses shared `respond()`/`respondError()` helpers (from `order.go`)
- Notify endpoint is unauthenticated (simulates third-party webhook callback)

#### BFF API Endpoints

| Method | Path | Auth | Description |
|--------|------|:---:|-------------|
| POST | `/api/v1/payments` | JWT | Create payment record |
| POST | `/api/v1/payments/:id/process` | JWT | Process payment via configured provider |
| GET | `/api/v1/payments/:id` | JWT | Get payment by ID |
| GET | `/api/v1/payments/by-order/:orderNo` | JWT | Get payment by order number |
| POST | `/api/v1/payments/:id/notify` | No | Third-party async callback (webhook) |

#### How to add a new payment provider (e.g. Alipay)

1. Create `service/internal/data/alipay_provider.go`:

```go
package data

import "github.com/storm/myidea/service/internal/biz"

type alipayProvider struct {
    appID      string
    privateKey string
    publicKey  string
}

func NewAlipayProvider(appID, privateKey, publicKey string) biz.PaymentProvider {
    return &alipayProvider{appID: appID, privateKey: privateKey, publicKey: publicKey}
}

func (p *alipayProvider) Pay(ctx context.Context, req *biz.PaymentProviderRequest) (*biz.PaymentProviderResponse, error) {
    // Call Alipay SDK: alipay.trade.page.pay
    // Return provider_trade_no and raw response
}

func (p *alipayProvider) VerifySignature(ctx context.Context, rawResponse string) (bool, error) {
    // Verify Alipay async notification signature
}

func (p *alipayProvider) Name() string { return "alipay" }
```

2. Register in `cmd/payment/main.go`:

```go
alipayProvider := data.NewAlipayProvider(os.Getenv("ALIPAY_APP_ID"), os.Getenv("ALIPAY_PRIVATE_KEY"), os.Getenv("ALIPAY_PUBLIC_KEY"))
providerFactory.Register(biz.PaymentMethodAlipay, alipayProvider)
```

3. BFF calls with `method: 2` to route through Alipay.

---

## 5. Development Conventions (for AI)

### 5.1 File Creation Order (Backend)

For each new service module, create files in this strict order:

1. `api/{module}/v1/{module}.proto` — define protobuf contract
2. Run `make proto` to generate Go stubs
3. `service/internal/biz/errors.go` — domain errors first
4. `service/internal/biz/{entity}.go` — domain model + repository interface + usecase
5. `service/internal/data/{entity}.go` — GORM implementation
6. `service/internal/service/{module}.go` — gRPC service implementation
7. Update `service/internal/data/data.go` — add `AutoMigrate` for new GORM models
8. `bff/handler/{module}.go` — HTTP handlers + gRPC client
9. Update `bff/main.go` — register new routes + init new handler

### 5.2 Coding Rules

- **Don't write proto-generated Go code by hand.** Always use `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc`. Proto files are the single source of truth.
- **Every `biz` entity must have behavior methods** (not anemic). At minimum: a status machine, a calculated field, or a validation method.
- **Repository interfaces go in `biz/`** (dependency inversion principle). GORM structs stay in `data/`.
- **`service/` only does proto↔domain mapping.** No business logic, no database calls.
- **BFF handler does HTTP parsing, gRPC client calls, and JSON response.** No domain logic.
- **Use Kratos `errors.New(code, reason, message)` for all domain error codes.** Don't use `fmt.Errorf` for client-facing errors.
- **Always propagate `context.Context`.** Every biz method, every data method takes `ctx` as first parameter.
- **go.work must list all modules.** After creating a new module, run `go work use ./<module>`.
- **Build from per-module directories** (not from go.work root with `./...`).

### 5.3 Protobuf Generation

```bash
cd /Users/storm/githubproject/myidea
export PATH="$HOME/go/bin:$PATH"
make proto
```

The Makefile runs:
```bash
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/{module}/v1/{module}.proto
```

### 5.4 Verifying Compilation

```bash
# Build each Go module separately (go.work root does NOT support go build ./...)
cd /Users/storm/githubproject/myidea/service && go build ./... && go vet ./...
cd /Users/storm/githubproject/myidea/bff && go build ./... && go vet ./...

# Or from the workspace root with explicit patterns:
cd /Users/storm/githubproject/myidea && go build ./service/... && go build ./bff/...

# Build frontend
cd /Users/storm/githubproject/myidea/web && npm run build
```

---

## 6. Roadmap

1. ✅ **Order Service** — Order CRUD, status state machine
2. ✅ **Product Service** — Category, Brand, SPU, SKU, Stock
3. ✅ **Cart Service** — Shopping cart CRUD, checkout orchestration
4. ✅ **User Service** — Register/login, address management
5. ✅ **Payment Service** — Mock payment, provider abstraction for 3rd-party SDKs
6. ✅ **Integration** — Full order creation workflow: LockStock → Order → Payment → ConfirmDeduct/UnlockStock
7. ✅ **Vue Frontend** — Complete SPA with 10 views, auth flow, cart management, order lifecycle

---

## 7. Next: Production Hardening

### 7.1 接入真实第三方支付

1. 实现 `AlipayProvider`（对接支付宝 SDK `alipay.trade.page.pay` / 异步通知验签）
2. 实现 `WechatPayProvider`（对接微信支付 SDK）
3. 通过 `providerFactory.Register()` 注册到系统中
4. 在 BFF 层区分支付渠道路由

### 7.2 测试与 CI

1. **单元测试**: 每个 `biz` usecase 使用 `go mock` 或 `testify` 编写测试
2. **集成测试**: 启动 MySQL + gRPC server，测试完整 RPC 流程
3. **CI**: 配置 GitHub Actions（`golangci-lint` → `go test` → `go build` → `npm run build`）
4. **Docker Compose**: 编排 MySQL + 5 个 gRPC server + BFF + Frontend

### 7.3 生产加固

1. **TLS**: gRPC 和 HTTP 都启用 TLS（`credentials.NewServerTLSFromFile`）
2. **配置管理**: 从 env vars 迁移到 YAML 配置文件 + viper
3. **速率限制**: BFF 层添加 `rate.Limiter` 中间件
4. **请求 ID**: 每个请求生成唯一 `request_id` 用于跟踪
5. **OpenTelemetry**: 集成完整的 trace + metric 导出（Jaeger / Prometheus）
6. **前端 PWA**: 添加 Service Worker, 离线缓存, 性能优化

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
