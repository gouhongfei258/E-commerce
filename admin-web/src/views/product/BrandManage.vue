<template>
  <div class="data-table">
    <el-button type="primary" style="margin-bottom:12px" @click="showDialog(null)">+ 新增品牌</el-button>
    <el-table :data="brands" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="品牌名称" min-width="150" />
      <el-table-column prop="logo" label="Logo" width="200">
        <template #default="{ row }"><span style="color:#409EFF">{{ row.logo }}</span></template>
      </el-table-column>
      <el-table-column prop="sort_order" label="排序" width="80" />
      <el-table-column prop="created_at" label="创建时间" width="160" />
    </el-table>
    <div class="pagination-wrap">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" @change="fetchBrands" layout="total,prev,pager,next" />
    </div>

    <el-dialog v-model="dialogVisible" title="新增品牌" width="400px">
      <el-form :model="bForm" label-width="80px">
        <el-form-item label="品牌名称"><el-input v-model="bForm.name" /></el-form-item>
        <el-form-item label="Logo URL"><el-input v-model="bForm.logo" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="bForm.sort_order" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" @click="create">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { listBrands, createBrand } from '../../api'
import { ElMessage } from 'element-plus'

const brands = ref([]); const total = ref(0); const loading = ref(false)
const page = ref(1); const pageSize = ref(10)
const dialogVisible = ref(false)
const bForm = reactive({ name: '', logo: '', sort_order: 0 })

async function fetchBrands() {
  loading.value = true
  try {
    const data = await listBrands({ page: page.value, page_size: pageSize.value })
    brands.value = data.brands || []
    total.value = data.total || 0
  } finally { loading.value = false }
}

function showDialog() { bForm.name = ''; bForm.logo = ''; bForm.sort_order = 0; dialogVisible.value = true }

async function create() {
  if (!bForm.name) { ElMessage.error('请输入品牌名称'); return }
  await createBrand({ name: bForm.name, logo: bForm.logo, sort_order: bForm.sort_order })
  dialogVisible.value = false
  ElMessage.success('创建成功')
  fetchBrands()
}

onMounted(fetchBrands)
</script>
