<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getOrder, cancelOrder } from '../api/order'
import { getPaymentByOrder, createPayment, processPayment } from '../api/payment'

const route = useRoute()
const router = useRouter()

const order = ref(null)
const payment = ref(null)
const loading = ref(true)
const error = ref('')
const cancelling = ref(false)
const paying = ref(false)
const successMsg = ref('')

const statusLabels = {
  0: 'Pending Payment',
  1: 'Paid',
  2: 'Shipped',
  3: 'Completed',
  4: 'Cancelled',
}

onMounted(async () => {
  try {
    const res = await getOrder(route.params.id)
    order.value = res.data
    // Try to load payment info
    try {
      const payRes = await getPaymentByOrder(res.data.order_no)
      payment.value = payRes.data
    } catch {
      // no payment yet
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

async function handleCancel() {
  if (!confirm('Are you sure you want to cancel this order?')) return
  cancelling.value = true
  try {
    await cancelOrder(route.params.id)
    successMsg.value = 'Order cancelled'
    // Refresh
    const res = await getOrder(route.params.id)
    order.value = res.data
  } catch (e) {
    error.value = e.message
  } finally {
    cancelling.value = false
  }
}

async function handlePay() {
  paying.value = true
  error.value = ''
  try {
    if (!payment.value) {
      const payRes = await createPayment({ order_id: order.value.id })
      payment.value = payRes.data
    }
    await processPayment(payment.value.id)
    successMsg.value = 'Payment successful!'
    // Refresh
    const res = await getOrder(route.params.id)
    order.value = res.data
  } catch (e) {
    error.value = e.message
  } finally {
    paying.value = false
  }
}
</script>

<template>
  <div v-if="loading" class="loading">
    <div class="spinner"></div>
  </div>

  <div v-else-if="error && !order" class="alert alert-error">{{ error }}</div>

  <div v-else-if="order" class="order-detail">
    <button class="btn btn-sm btn-outline mb-16" @click="router.back()">&larr; Back</button>

    <div v-if="successMsg" class="alert alert-success">{{ successMsg }}</div>
    <div v-if="error" class="alert alert-error">{{ error }}</div>

    <!-- Order Header -->
    <div class="card">
      <div class="order-header">
        <div>
          <h3>Order #{{ order.order_no || order.id }}</h3>
          <span class="badge mt-8" :class="{
            'badge-warning': order.status === 0,
            'badge-success': order.status === 1 || order.status === 3,
            'badge-info': order.status === 2,
            'badge-danger': order.status === 4,
          }">{{ statusLabels[order.status] || 'Unknown' }}</span>
        </div>
        <div class="order-actions">
          <button
            v-if="order.status === 0"
            class="btn btn-primary"
            :disabled="paying"
            @click="handlePay"
          >
            {{ paying ? 'Processing...' : 'Pay Now' }}
          </button>
          <button
            v-if="order.status === 0"
            class="btn btn-outline"
            :disabled="cancelling"
            @click="handleCancel"
          >
            {{ cancelling ? 'Cancelling...' : 'Cancel Order' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Shipping Address -->
    <div class="card mt-16" v-if="order.address">
      <h4 class="mb-8">Shipping Address</h4>
      <p>{{ order.address.receiver_name }} &nbsp; {{ order.address.receiver_phone }}</p>
      <p class="text-secondary text-sm">{{ order.address.province }} {{ order.address.city }} {{ order.address.district }} {{ order.address.detail_address }}</p>
    </div>

    <!-- Order Items -->
    <div class="card mt-16">
      <h4 class="mb-16">Order Items</h4>
      <div v-for="item in order.items" :key="item.id" class="order-item">
        <div class="item-info">
          <span class="item-name">{{ item.spu_name || item.sku_name || `SKU #${item.sku_id}` }}</span>
          <span class="text-secondary text-sm" v-if="item.sku_name">{{ item.sku_name }}</span>
        </div>
        <div class="item-detail">
          <span class="item-price">&yen;{{ ((item.price || 0) / 100).toFixed(2) }}</span>
          <span class="item-qty">x{{ item.quantity }}</span>
          <span class="item-subtotal">&yen;{{ (((item.price || 0) * item.quantity) / 100).toFixed(2) }}</span>
        </div>
      </div>
    </div>

    <!-- Total -->
    <div class="card mt-16">
      <div class="total-row">
        <span>Total Amount</span>
        <span class="total-price">&yen;{{ ((order.total_amount || 0) / 100).toFixed(2) }}</span>
      </div>
    </div>

    <!-- Payment Info -->
    <div class="card mt-16" v-if="payment">
      <h4 class="mb-8">Payment Info</h4>
      <p class="text-secondary text-sm">Payment #{{ payment.payment_no || payment.id }}</p>
      <p class="text-secondary text-sm">
        Status:
        <span :class="{
          'badge badge-success': payment.status === 'paid' || payment.status === 1,
          'badge badge-warning': payment.status === 'pending' || payment.status === 0,
          'badge badge-danger': payment.status === 'failed' || payment.status === 2,
        }">{{ payment.status }}</span>
      </p>
    </div>
  </div>
</template>

<style scoped>
.order-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.order-actions {
  display: flex;
  gap: 8px;
}

.order-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border);
}

.order-item:last-child {
  border-bottom: none;
}

.item-info {
  display: flex;
  flex-direction: column;
}

.item-name {
  font-weight: 500;
}

.item-detail {
  display: flex;
  align-items: center;
  gap: 16px;
}

.item-price {
  color: var(--text-secondary);
  font-size: 14px;
}

.item-qty {
  color: var(--text-secondary);
  font-size: 14px;
}

.item-subtotal {
  font-weight: 600;
  color: var(--danger);
  min-width: 80px;
  text-align: right;
}

.total-row {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 24px;
  font-size: 16px;
}

.total-price {
  font-size: 22px;
  font-weight: 700;
  color: var(--danger);
}
</style>
