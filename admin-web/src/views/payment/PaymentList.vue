<template>
  <div class="filter-bar">
    <el-input v-model="filter.order_no" placeholder="搜索订单号" style="width:220px" clearable @keyup.enter="search" />
    <el-select v-model="filter.status" clearable placeholder="状态" style="width:130px;margin-left:8px">
      <el-option label="待支付" :value="1" />
      <el-option label="成功" :value="2" />
      <el-option label="失败" :value="3" />
      <el-option label="退款中" :value="4" />
      <el-option label="已退款" :value="5" />
    </el-select>
    <el-button type="primary" style="margin-left:8px" @click="search">搜索</el-button>
  </div>
  <div class="data-table">
    <el-table :data="payments" v-loading="loading" stripe>
      <el-table-column prop="order_no" label="订单号" width="180" />
      <el-table-column prop="user_id" label="用户ID" width="80" />
      <el-table-column label="金额" width="100">
        <template #default="{ row }">¥{{ row.total_amount }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="支付方式" width="100">
        <template #default="{ row }">{{ methodText(row.method) }}</template>
      </el-table-column>
      <el-table-column prop="provider_trade_no" label="交易号" min-width="140" />
      <el-table-column prop="created_at" label="创建时间" width="160" />
    </el-table>
    <div class="pagination-wrap">
      <el-pagination v-model:current-page="filter.page" v-model:page-size="filter.page_size" :total="total" :page-sizes="[10,20,50]" layout="total,sizes,prev,pager,next" @change="fetchPayments" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { listPayments } from '../../api'
const payments = ref([]); const total = ref(0); const loading = ref(false)
const filter = reactive({ page: 1, page_size: 10, status: null, order_no: '' })

function statusText(s) {
  const m = { 1: '待支付', 2: '成功', 3: '失败', 4: '退款中', 5: '已退款' }
  return m[s] || '未知'
}
function statusType(s) {
  const m = { 1: 'warning', 2: 'success', 3: 'danger', 4: 'warning', 5: 'info' }
  return m[s] || 'info'
}
function methodText(m) {
  const t = { 1: 'Mock', 2: '支付宝', 3: '微信支付' }
  return t[m] || '未知'
}

async function fetchPayments() {
  loading.value = true
  try {
    const params = { page: filter.page, page_size: filter.page_size }
    if (filter.status) params.status = filter.status
    if (filter.order_no) params.order_no = filter.order_no
    const data = await listPayments(params)
    payments.value = data.payments || []
    total.value = data.total || 0
  } finally { loading.value = false }
}
function search() { filter.page = 1; fetchPayments() }
onMounted(fetchPayments)
</script>
