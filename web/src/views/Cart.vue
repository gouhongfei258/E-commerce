<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'

const router = useRouter()
const cart = useCartStore()

onMounted(() => {
  cart.fetchItems()
})

async function handleUpdate(item, delta) {
  const qty = item.quantity + delta
  if (qty < 1) return
  try {
    await cart.updateQuantity(item.id, qty)
  } catch (e) {
    // error handled by store
  }
}

async function handleRemove(id) {
  try {
    await cart.removeItem(id)
  } catch (e) {
    // error handled by store
  }
}

function goCheckout() {
  router.push('/checkout')
}

function goProduct(skuId) {
  // Navigate to product detail — we may not have spu_id in cart items
  router.push('/')
}
</script>

<template>
  <div>
    <div class="page-header">
      <h2>Shopping Cart</h2>
    </div>

    <div v-if="cart.loading" class="loading">
      <div class="spinner"></div>
    </div>

    <div v-else-if="cart.items.length === 0" class="empty-state">
      <p>Your cart is empty</p>
      <router-link to="/" class="btn btn-primary">Browse Products</router-link>
    </div>

    <div v-else class="cart-layout">
      <div class="cart-items">
        <div v-for="item in cart.items" :key="item.id" class="cart-item card">
          <div class="item-info">
            <h4>{{ item.spu_name || item.sku_name || `SKU #${item.sku_id}` }}</h4>
            <p class="text-secondary text-sm" v-if="item.sku_name">SKU: {{ item.sku_name }}</p>
            <p class="item-price">&yen;{{ ((item.price || 0) / 100).toFixed(2) }}</p>
          </div>
          <div class="item-actions">
            <div class="qty-control">
              <button class="btn btn-sm btn-outline" :disabled="item.quantity <= 1"
                @click="handleUpdate(item, -1)">-</button>
              <span class="qty-value">{{ item.quantity }}</span>
              <button class="btn btn-sm btn-outline" @click="handleUpdate(item, 1)">+</button>
            </div>
            <span class="item-subtotal">
              &yen;{{ (((item.price || 0) * item.quantity) / 100).toFixed(2) }}
            </span>
            <button class="btn btn-sm btn-danger" @click="handleRemove(item.id)">Remove</button>
          </div>
        </div>
      </div>

      <div class="cart-summary card">
        <h3 class="mb-16">Order Summary</h3>
        <div class="summary-row">
          <span>Total Items</span>
          <span>{{ cart.totalCount }}</span>
        </div>
        <div class="summary-row summary-total">
          <span>Total</span>
          <span class="total-price">&yen;{{ (cart.totalPrice / 100).toFixed(2) }}</span>
        </div>
        <button class="btn btn-primary btn-lg btn-full mt-16" @click="goCheckout">
          Proceed to Checkout
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cart-layout {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 24px;
  align-items: start;
}

.cart-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.item-info h4 {
  font-size: 16px;
  margin-bottom: 4px;
}

.item-price {
  color: var(--danger);
  font-weight: 600;
  margin-top: 4px;
}

.item-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.qty-control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.qty-value {
  font-size: 16px;
  font-weight: 600;
  min-width: 24px;
  text-align: center;
}

.item-subtotal {
  font-size: 16px;
  font-weight: 600;
  color: var(--danger);
  min-width: 100px;
  text-align: right;
}

.cart-summary {
  position: sticky;
  top: 80px;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.summary-total {
  border-top: 1px solid var(--border);
  margin-top: 8px;
  padding-top: 12px;
}

.summary-total span {
  font-size: 18px;
  font-weight: 600;
  color: var(--text);
}

.total-price {
  color: var(--danger) !important;
}

.btn-full {
  width: 100%;
}

@media (max-width: 768px) {
  .cart-layout {
    grid-template-columns: 1fr;
  }

  .cart-item {
    flex-direction: column;
    align-items: flex-start;
  }

  .item-actions {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
