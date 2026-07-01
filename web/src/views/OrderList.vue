<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listOrders } from '../api/order'

const router = useRouter()
const orders = ref([])
const loading = ref(true)
const error = ref('')
const filters = ref({
  status: '',
  page: 1,
  page_size: 10,
})
const total = ref(0)
const totalPages = ref(0)

const statusLabels = {
  0: 'Pending',
  1: 'Paid',
  2: 'Shipped',
  3: 'Completed',
  4: 'Cancelled',
}

async function fetchOrders() {
  loading.value = true
  error.value = ''
  try {
    const params = { ...filters.value }
    Object.keys(params).forEach((k) => {
      if (params[k] === '' || params[k] === undefined) delete params[k]
    })
    const res = await listOrders(params)
    orders.value = res.data?.list || []
    total.value = res.data?.total || 0
    totalPages.value = Math.max(1, Math.ceil(total.value / filters.value.page_size))
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function goDetail(id) {
  router.push(`/orders/${id}`)
}

function onFilterChange() {
  filters.value.page = 1
  fetchOrders()
}

function goPage(p) {
  if (p < 1 || p > totalPages.value) return
  filters.value.page = p
  fetchOrders()
}

onMounted(fetchOrders)
</script>

<template>
  <div>
    <div class="page-header">
      <h2>My Orders</h2>
    </div>

    <!-- Status Filter -->
    <div class="card mb-16">
      <div class="status-tabs">
        <button
          class="status-tab"
          :class="{ active: filters.status === '' }"
          @click="filters.status = ''; onFilterChange()"
        >All</button>
        <button
          v-for="(label, key) in statusLabels"
          :key="key"
          class="status-tab"
          :class="{ active: String(filters.status) === String(key) }"
          @click="filters.status = String(key); onFilterChange()"
        >{{ label }}</button>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <div v-else-if="error" class="alert alert-error">{{ error }}</div>

    <div v-else-if="orders.length === 0" class="empty-state">
      <p>No orders found</p>
      <router-link to="/" class="btn btn-primary">Start Shopping</router-link>
    </div>

    <div v-else class="order-list">
      <div v-for="order in orders" :key="order.id" class="order-card card" @click="goDetail(order.id)">
        <div class="order-header">
          <div>
            <span class="order-no">Order #{{ order.order_no || order.id }}</span>
            <span class="badge ml-8" :class="{
              'badge-warning': order.status === 0,
              'badge-success': order.status === 1,
              'badge-info': order.status === 2,
              'badge-success': order.status === 3,
              'badge-danger': order.status === 4,
            }">{{ statusLabels[order.status] || 'Unknown' }}</span>
          </div>
          <span class="text-secondary text-sm">{{ order.created_at || '' }}</span>
        </div>
        <div class="order-body">
          <div class="order-items-preview">
            <span v-if="order.items && order.items.length > 0">
              {{ order.items.map(i => i.spu_name || i.sku_name).join(', ') }}
            </span>
            <span v-else class="text-secondary">{{ order.item_count || 0 }} item(s)</span>
          </div>
          <div class="order-total">&yen;{{ ((order.total_amount || 0) / 100).toFixed(2) }}</div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button class="btn btn-sm btn-outline" :disabled="filters.page <= 1" @click="goPage(filters.page - 1)">
        Previous
      </button>
      <span class="page-info">{{ filters.page }} / {{ totalPages }}</span>
      <button class="btn btn-sm btn-outline" :disabled="filters.page >= totalPages" @click="goPage(filters.page + 1)">
        Next
      </button>
    </div>
  </div>
</template>

<style scoped>
.status-tabs {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.status-tab {
  padding: 6px 16px;
  border: 1px solid var(--border);
  background: transparent;
  border-radius: 20px;
  font-size: 13px;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.status-tab:hover {
  border-color: var(--primary);
  color: var(--primary);
}

.status-tab.active {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
}

.order-card {
  cursor: pointer;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border);
}

.order-no {
  font-weight: 600;
  font-size: 14px;
}

.order-body {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.order-items-preview {
  font-size: 14px;
  color: var(--text-secondary);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 16px;
}

.order-total {
  font-size: 18px;
  font-weight: 600;
  color: var(--danger);
}

.ml-8 {
  margin-left: 8px;
}
</style>
