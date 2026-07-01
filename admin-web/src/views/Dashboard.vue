<template>
  <div>
    <div class="stat-cards">
      <div class="stat-card">
        <div class="label">总订单数</div>
        <div class="value">{{ stats.total_orders }}</div>
      </div>
      <div class="stat-card">
        <div class="label">总营收</div>
        <div class="value">¥{{ stats.total_revenue.toFixed(2) }}</div>
      </div>
      <div class="stat-card">
        <div class="label">用户数</div>
        <div class="value">{{ stats.total_users }}</div>
      </div>
      <div class="stat-card">
        <div class="label">商品数</div>
        <div class="value">{{ stats.total_products }}</div>
      </div>
    </div>

    <div class="chart-row">
      <div class="chart-box">
        <div class="chart-title">最近订单</div>
        <el-table :data="recentOrders" size="small">
          <el-table-column prop="order_no" label="订单号" width="180" />
          <el-table-column prop="user_id" label="用户ID" width="80" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="金额" width="120">
            <template #default="{ row }">¥{{ row.total_amount }}</template>
          </el-table-column>
          <el-table-column prop="created_at" label="时间" width="180" />
        </el-table>
      </div>
      <div class="chart-box">
        <div class="chart-title">订单状态分布</div>
        <v-chart :option="pieOption" style="height:280px" autoresize />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { getDashboard } from '../api'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([PieChart, TitleComponent, TooltipComponent, LegendComponent, CanvasRenderer])

const stats = ref({ total_orders: 0, total_revenue: 0, total_users: 0, total_products: 0 })
const recentOrders = ref([])

function statusText(s) {
  const m = { 0: '待支付', 1: '已支付', 2: '已发货', 3: '已签收', 4: '已取消', 5: '退款中', 6: '已退款' }
  return m[s] || '未知'
}
function statusType(s) {
  const m = { 0: 'warning', 1: 'primary', 2: 'success', 3: 'success', 4: 'info', 5: 'danger', 6: 'info' }
  return m[s] || 'info'
}

const pieOption = computed(() => ({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0 },
  series: [{
    type: 'pie',
    radius: ['40%', '70%'],
    data: [
      { value: 0, name: '待支付' },
      { value: 0, name: '已支付' },
      { value: 0, name: '已发货' },
      { value: 0, name: '已签收' },
      { value: 0, name: '已取消' },
      { value: 0, name: '退款中/已退款' }
    ]
  }]
}))

onMounted(async () => {
  try {
    const data = await getDashboard()
    stats.value = { total_orders: data.total_orders, total_revenue: data.total_revenue, total_users: data.total_users, total_products: data.total_products }
    recentOrders.value = data.recent_orders || []
  } catch (e) { /* fail silently */ }
})
</script>
