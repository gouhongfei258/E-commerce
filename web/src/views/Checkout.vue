<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { listAddresses } from '../api/address'
import { createOrder } from '../api/order'

const router = useRouter()
const cart = useCartStore()

const addresses = ref([])
const selectedAddressId = ref('')
const loading = ref(true)
const submitting = ref(false)
const error = ref('')

onMounted(async () => {
  await cart.fetchItems()
  try {
    const res = await listAddresses()
    addresses.value = res.data || []
    const defaultAddr = addresses.value.find((a) => a.is_default) || addresses.value[0]
    if (defaultAddr) {
      selectedAddressId.value = defaultAddr.id
    }
  } catch (e) {
    // non-critical
  } finally {
    loading.value = false
  }
})

async function handleCheckout() {
  if (!selectedAddressId.value) {
    error.value = 'Please select a shipping address'
    return
  }
  if (cart.items.length === 0) {
    error.value = 'Your cart is empty'
    return
  }
  error.value = ''
  submitting.value = true
  try {
    const res = await createOrder({ address_id: selectedAddressId.value })
    router.push(`/orders/${res.data.id}`)
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div>
    <div class="page-header">
      <h2>Checkout</h2>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <div v-else class="checkout-layout">
      <div class="checkout-main">
        <!-- Address Selection -->
        <div class="card">
          <h3 class="mb-16">Shipping Address</h3>
          <div v-if="addresses.length === 0" class="empty-state">
            <p>No addresses found. Please add one first.</p>
            <router-link to="/addresses" class="btn btn-primary">Manage Addresses</router-link>
          </div>
          <div v-else class="address-list">
            <div
              v-for="addr in addresses"
              :key="addr.id"
              class="address-item"
              :class="{ active: selectedAddressId === addr.id }"
              @click="selectedAddressId = addr.id"
            >
              <div class="address-radio">
                <div class="radio-circle" :class="{ checked: selectedAddressId === addr.id }"></div>
              </div>
              <div class="address-detail">
                <div class="address-name">
                  {{ addr.receiver_name }} &nbsp; {{ addr.receiver_phone }}
                  <span v-if="addr.is_default" class="badge badge-info ml-8">Default</span>
                </div>
                <div class="address-text">{{ addr.province }} {{ addr.city }} {{ addr.district }} {{ addr.detail_address }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Order Items -->
        <div class="card mt-16">
          <h3 class="mb-16">Order Items ({{ cart.totalCount }})</h3>
          <div v-for="item in cart.items" :key="item.id" class="order-item">
            <div class="item-info">
              <span class="item-name">{{ item.spu_name || `SKU #${item.sku_id}` }}</span>
              <span class="text-secondary text-sm" v-if="item.sku_name">{{ item.sku_name }}</span>
            </div>
            <div class="item-qty">x{{ item.quantity }}</div>
            <div class="item-price">&yen;{{ (((item.price || 0) * item.quantity) / 100).toFixed(2) }}</div>
          </div>
        </div>
      </div>

      <!-- Summary -->
      <div class="checkout-side">
        <div class="card">
          <h3 class="mb-16">Order Summary</h3>
          <div class="summary-row">
            <span>Items</span>
            <span>{{ cart.totalCount }}</span>
          </div>
          <div class="summary-row">
            <span>Subtotal</span>
            <span>&yen;{{ (cart.totalPrice / 100).toFixed(2) }}</span>
          </div>
          <div class="summary-row summary-total">
            <span>Total</span>
            <span class="total-price">&yen;{{ (cart.totalPrice / 100).toFixed(2) }}</span>
          </div>

          <div v-if="error" class="alert alert-error mt-16">{{ error }}</div>

          <button
            class="btn btn-primary btn-lg btn-full mt-16"
            :disabled="submitting || cart.items.length === 0"
            @click="handleCheckout"
          >
            {{ submitting ? 'Placing Order...' : 'Place Order' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.checkout-layout {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 24px;
  align-items: start;
}

.address-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.address-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  cursor: pointer;
  transition: all 0.2s;
}

.address-item:hover {
  border-color: var(--primary);
}

.address-item.active {
  border-color: var(--primary);
  background: #F0F7FF;
}

.radio-circle {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border);
  border-radius: 50%;
  margin-top: 2px;
  transition: all 0.2s;
}

.radio-circle.checked {
  border-color: var(--primary);
  background: var(--primary);
  box-shadow: inset 0 0 0 3px white;
}

.address-name {
  font-weight: 500;
  font-size: 14px;
}

.address-text {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.order-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid var(--border);
  gap: 12px;
}

.order-item:last-child {
  border-bottom: none;
}

.item-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.item-name {
  font-weight: 500;
  font-size: 14px;
}

.item-qty {
  color: var(--text-secondary);
  font-size: 14px;
}

.item-price {
  font-weight: 600;
  color: var(--danger);
  font-size: 14px;
  min-width: 80px;
  text-align: right;
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

.ml-8 {
  margin-left: 8px;
}

@media (max-width: 768px) {
  .checkout-layout {
    grid-template-columns: 1fr;
  }
}
</style>
