<template>
  <div>
    <div class="filter-bar">
      <el-form :inline="true" :model="filter" size="default">
        <el-form-item label="状态">
          <el-select v-model="filter.status" clearable placeholder="全部" style="width:130px">
            <el-option label="待支付" :value="0" />
            <el-option label="已支付" :value="1" />
            <el-option label="已发货" :value="2" />
            <el-option label="已签收" :value="3" />
            <el-option label="已取消" :value="4" />
            <el-option label="退款中" :value="5" />
            <el-option label="已退款" :value="6" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filter.keyword" placeholder="订单号/收货人" clearable style="width:200px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">搜索</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="data-table">
      <el-table :data="orders" v-loading="loading" stripe>
        <el-table-column prop="order_no" label="订单号" width="180" />
        <el-table-column prop="user_id" label="用户ID" width="80" />
        <el-table-column prop="total_amount" label="金额" width="100">
          <template #default="{ row }">¥{{ row.total_amount }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="address.receiver_name" label="收货人" width="100" />
        <el-table-column prop="created_at" label="下单时间" width="180" />
        <el-table-column label="操作" min-width="180">
          <template #default="{ row }">
            <el-button size="small" @click="$router.push(`/admin/orders/${row.id}`)">详情</el-button>
            <el-button v-if="row.status === 0" size="small" type="success" @click="ship(row)">发货</el-button>
            <el-button v-if="row.status === 1" size="small" type="warning" @click="refund(row)">退款</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="filter.page"
          v-model:page-size="filter.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @current-change="fetchOrders"
          @size-change="fetchOrders"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { listOrders, shipOrder, refundOrder } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const orders = ref([])
const total = ref(0)
const loading = ref(false)
const filter = reactive({ page: 1, page_size: 10, status: null, keyword: '' })

function statusText(s) {
  const m = { 0: '待支付', 1: '已支付', 2: '已发货', 3: '已签收', 4: '已取消', 5: '退款中', 6: '已退款' }
  return m[s] || '未知'
}
function statusType(s) {
  const m = { 0: 'warning', 1: 'primary', 2: 'success', 3: 'success', 4: 'info', 5: 'danger', 6: 'info' }
  return m[s] || 'info'
}

async function fetchOrders() {
  loading.value = true
  try {
    const params = { page: filter.page, page_size: filter.page_size }
    if (filter.status != null) params.status = filter.status
    if (filter.keyword) params.keyword = filter.keyword
    const data = await listOrders(params)
    orders.value = data.orders || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

function search() { filter.page = 1; fetchOrders() }
function reset() { filter.status = null; filter.keyword = ''; search() }

async function ship(row) {
  try {
    await ElMessageBox.confirm(`确认将订单 ${row.order_no} 发货？`, '提示', { type: 'info' })
    await shipOrder(row.id)
    ElMessage.success('发货成功')
    fetchOrders()
  } catch (e) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

async function refund(row) {
  try {
    await ElMessageBox.confirm(`确认对订单 ${row.order_no} 发起退款？`, '提示', { type: 'warning' })
    await refundOrder(row.id)
    ElMessage.success('已发起退款')
    fetchOrders()
  } catch (e) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

onMounted(() => fetchOrders())
</script>
