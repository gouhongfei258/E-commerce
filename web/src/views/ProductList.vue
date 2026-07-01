<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { listSPUs, getCategories, listBrands } from '../api/product'

const router = useRouter()

const spus = ref([])
const categories = ref([])
const brands = ref([])
const loading = ref(false)
const error = ref('')

const filters = ref({
  category_id: '',
  brand_id: '',
  keyword: '',
  page: 1,
  page_size: 12,
})

const total = ref(0)

async function fetchCategories() {
  try {
    const res = await getCategories()
    categories.value = res.data || []
  } catch (e) {
    // non-critical
  }
}

async function fetchBrands() {
  try {
    const res = await listBrands({ page: 1, page_size: 100 })
    brands.value = res.data?.list || []
  } catch (e) {
    // non-critical
  }
}

async function fetchSPUs() {
  loading.value = true
  error.value = ''
  try {
    const params = { ...filters.value }
    Object.keys(params).forEach((k) => {
      if (!params[k] && params[k] !== 0) delete params[k]
    })
    const res = await listSPUs(params)
    spus.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function onSearch() {
  filters.value.page = 1
  fetchSPUs()
}

function goDetail(id) {
  router.push(`/products/${id}`)
}

watch(() => filters.value.category_id, () => {
  filters.value.page = 1
  fetchSPUs()
})

watch(() => filters.value.brand_id, () => {
  filters.value.page = 1
  fetchSPUs()
})

const totalPages = ref(0)
watch(total, (t) => {
  totalPages.value = Math.max(1, Math.ceil(t / filters.value.page_size))
})

function goPage(p) {
  if (p < 1 || p > totalPages.value) return
  filters.value.page = p
  fetchSPUs()
}

onMounted(() => {
  fetchCategories()
  fetchBrands()
  fetchSPUs()
})
</script>

<template>
  <div>
    <div class="page-header">
      <h2>Products</h2>
    </div>

    <!-- Filters -->
    <div class="card mb-16">
      <div class="filters">
        <div class="filter-group">
          <label>Category</label>
          <select v-model="filters.category_id" class="form-select">
            <option value="">All Categories</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="filter-group">
          <label>Brand</label>
          <select v-model="filters.brand_id" class="form-select">
            <option value="">All Brands</option>
            <option v-for="b in brands" :key="b.id" :value="b.id">{{ b.name }}</option>
          </select>
        </div>
        <div class="filter-group">
          <label>Keyword</label>
          <input v-model="filters.keyword" class="form-input" placeholder="Search products..."
            @keyup.enter="onSearch" />
        </div>
        <div class="filter-group filter-action">
          <label>&nbsp;</label>
          <button class="btn btn-primary" @click="onSearch">Search</button>
        </div>
      </div>
    </div>

    <!-- Error -->
    <div v-if="error" class="alert alert-error">{{ error }}</div>

    <!-- Loading -->
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <!-- Product Grid -->
    <div v-else-if="spus.length === 0" class="empty-state">
      <p>No products found</p>
      <button class="btn btn-outline" @click="filters = { page: 1, page_size: 12 }; fetchSPUs()">Clear Filters</button>
    </div>

    <div v-else class="product-grid">
      <div v-for="spu in spus" :key="spu.id" class="product-card card" @click="goDetail(spu.id)">
        <div class="product-image">
          <div class="image-placeholder">{{ spu.name.charAt(0) }}</div>
        </div>
        <div class="product-info">
          <h3 class="product-name">{{ spu.name }}</h3>
          <p v-if="spu.description" class="product-desc">{{ spu.description }}</p>
          <p class="product-brand" v-if="spu.brand_name">{{ spu.brand_name }}</p>
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
.filters {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  align-items: flex-end;
}

.filter-group {
  flex: 1;
  min-width: 180px;
}

.filter-group label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.filter-action {
  flex: 0 0 auto;
  min-width: auto;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.product-card {
  cursor: pointer;
  padding: 0;
  overflow: hidden;
  transition: transform 0.2s;
}

.product-card:hover {
  transform: translateY(-2px);
}

.product-image {
  height: 180px;
  background: linear-gradient(135deg, #E3F2FD, #BBDEFB);
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-placeholder {
  font-size: 48px;
  font-weight: 700;
  color: var(--primary);
  opacity: 0.6;
}

.product-info {
  padding: 16px;
}

.product-name {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.product-desc {
  font-size: 13px;
  color: var(--text-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.product-brand {
  font-size: 12px;
  color: var(--primary);
  margin-top: 8px;
}
</style>
