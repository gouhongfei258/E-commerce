<template>
  <div class="filter-bar">
    <el-form :inline="true" :model="filter">
      <el-form-item label="类目">
        <el-input-number v-model="filter.category_id" :min="0" placeholder="类目ID" />
      </el-form-item>
      <el-form-item label="品牌">
        <el-input-number v-model="filter.brand_id" :min="0" placeholder="品牌ID" />
      </el-form-item>
      <el-form-item label="关键词">
        <el-input v-model="filter.keyword" placeholder="商品名称" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="search">搜索</el-button>
        <el-button @click="$router.push('/admin/products/spus/new')">+ 新增商品</el-button>
      </el-form-item>
    </el-form>
  </div>
  <div class="data-table">
    <el-table :data="spus" v-loading="loading" stripe @row-click="(row) => $router.push(`/admin/products/spus/${row.id}/edit`)" style="cursor:pointer">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="title" label="商品名称" min-width="200" />
      <el-table-column prop="category_id" label="类目ID" width="80" />
      <el-table-column prop="brand_id" label="品牌ID" width="80" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '上架' : '下架' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="sale_count" label="销量" width="60" />
      <el-table-column prop="created_at" label="创建时间" width="160" />
    </el-table>
    <div class="pagination-wrap">
      <el-pagination v-model:current-page="filter.page" v-model:page-size="filter.page_size" :total="total" :page-sizes="[10,20,50]" layout="total,sizes,prev,pager,next" @change="fetchSPUs" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { listSPUs } from '../../api'
const spus = ref([]); const total = ref(0); const loading = ref(false)
const filter = reactive({ page: 1, page_size: 10, category_id: null, brand_id: null, keyword: '' })
async function fetchSPUs() {
  loading.value = true
  try {
    const params = { page: filter.page, page_size: filter.page_size, keyword: filter.keyword }
    if (filter.category_id) params.category_id = filter.category_id
    if (filter.brand_id) params.brand_id = filter.brand_id
    const data = await listSPUs(params)
    spus.value = data.spus || []
    total.value = data.total || 0
  } finally { loading.value = false }
}
function search() { filter.page = 1; fetchSPUs() }
onMounted(fetchSPUs)
</script>
