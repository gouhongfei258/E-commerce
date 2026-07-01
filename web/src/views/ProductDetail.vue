<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getSPU } from '../api/product'
import { useCartStore } from '../stores/cart'

const route = useRoute()
const router = useRouter()
const cart = useCartStore()

const spu = ref(null)
const skus = ref([])
const selectedSku = ref(null)
const quantity = ref(1)
const loading = ref(true)
const error = ref('')
const adding = ref(false)
const addedMsg = ref('')

onMounted(async () => {
  try {
    const res = await getSPU(route.params.id)
    spu.value = res.data
    skus.value = res.data.skus || []
    if (skus.value.length > 0) {
      selectedSku.value = skus.value[0]
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

function selectSku(sku) {
  selectedSku.value = sku
  quantity.value = 1
}

async function addToCart() {
  if (!selectedSku.value) return
  adding.value = true
  addedMsg.value = ''
  try {
    await cart.addItem(selectedSku.value.id, quantity.value)
    addedMsg.value = 'Added to cart!'
    setTimeout(() => { addedMsg.value = '' }, 2000)
  } catch (e) {
    error.value = e.message
  } finally {
    adding.value = false
  }
}
</script>

<template>
  <div v-if="loading" class="loading">
    <div class="spinner"></div>
  </div>

  <div v-else-if="error" class="alert alert-error">{{ error }}</div>

  <div v-else-if="spu" class="product-detail">
    <button class="btn btn-sm btn-outline mb-16" @click="router.back()">&larr; Back</button>

    <div class="detail-layout">
      <div class="detail-image">
        <div class="main-image">{{ spu.name.charAt(0) }}</div>
      </div>
      <div class="detail-info card">
        <h1>{{ spu.name }}</h1>
        <p class="text-secondary mt-8" v-if="spu.description">{{ spu.description }}</p>
        <p class="text-secondary text-sm mt-8" v-if="spu.brand_name">Brand: {{ spu.brand_name }}</p>
        <p class="text-secondary text-sm" v-if="spu.category_name">Category: {{ spu.category_name }}</p>

        <!-- SKU Selection -->
        <div class="mt-24" v-if="skus.length > 0">
          <h3 class="mb-8">Select SKU</h3>
          <div class="sku-list">
            <div
              v-for="sku in skus"
              :key="sku.id"
              class="sku-item"
              :class="{ active: selectedSku?.id === sku.id }"
              @click="selectSku(sku)"
            >
              <div class="sku-name">{{ sku.name || sku.attrs || `SKU #${sku.id}` }}</div>
              <div class="sku-price">&yen;{{ (sku.price / 100).toFixed(2) }}</div>
              <div class="sku-stock" v-if="sku.stock !== undefined">
                Stock: {{ sku.stock }}
              </div>
            </div>
          </div>
        </div>

        <!-- Selected SKU info -->
        <div v-if="selectedSku" class="mt-16 selected-info">
          <span class="text-lg font-bold">&yen;{{ (selectedSku.price / 100).toFixed(2) }}</span>
          <span v-if="selectedSku.stock !== undefined" class="text-secondary text-sm ml-8">
            ({{ selectedSku.stock }} in stock)
          </span>
        </div>

        <!-- Quantity -->
        <div class="form-group mt-16">
          <label>Quantity</label>
          <div class="qty-control">
            <button class="btn btn-sm btn-outline" :disabled="quantity <= 1" @click="quantity--">-</button>
            <input v-model.number="quantity" class="form-input qty-input" type="number" min="1" />
            <button class="btn btn-sm btn-outline" @click="quantity++">+</button>
          </div>
        </div>

        <!-- Actions -->
        <div class="mt-24">
          <div v-if="addedMsg" class="alert alert-success">{{ addedMsg }}</div>
          <button
            class="btn btn-primary btn-lg"
            :disabled="!selectedSku || adding"
            @click="addToCart"
          >
            {{ adding ? 'Adding...' : 'Add to Cart' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.detail-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.detail-image {
  height: 400px;
  background: linear-gradient(135deg, #E3F2FD, #BBDEFB);
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
}

.main-image {
  font-size: 80px;
  font-weight: 700;
  color: var(--primary);
  opacity: 0.4;
}

.detail-info h1 {
  font-size: 24px;
  font-weight: 700;
}

.sku-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.sku-item {
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 10px 14px;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 120px;
}

.sku-item:hover {
  border-color: var(--primary);
}

.sku-item.active {
  border-color: var(--primary);
  background: #E3F2FD;
}

.sku-name {
  font-size: 13px;
  font-weight: 500;
}

.sku-price {
  font-size: 14px;
  color: var(--danger);
  font-weight: 600;
  margin-top: 4px;
}

.sku-stock {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.qty-control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.qty-input {
  width: 80px;
  text-align: center;
}

.ml-8 {
  margin-left: 8px;
}

@media (max-width: 768px) {
  .detail-layout {
    grid-template-columns: 1fr;
  }
}
</style>
