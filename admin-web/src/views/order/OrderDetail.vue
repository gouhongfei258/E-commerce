<template>
  <div>
    <el-button @click="$router.back()" style="margin-bottom:16px">← 返回</el-button>
    <el-card v-loading="loading">
      <template v-if="order">
        <el-descriptions title="订单基本信息" :column="2" border>
          <el-descriptions-item label="订单号">{{ order.order_no }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType(order.status)">{{ statusText(order.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="用户ID">{{ order.user_id }}</el-descriptions-item>
          <el-descriptions-item label="总金额">¥{{ order.total_amount }}</el-descriptions-item>
          <el-descriptions-item label="已支付">¥{{ order.paid_amount }}</el-descriptions-item>
          <el-descriptions-item label="支付方式">{{ order.payment_method }}</el-descriptions-item>
          <el-descriptions-item label="下单时间">{{ order.created_at }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ order.remark || '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-descriptions title="收货地址" :column="3" border style="margin-top:16px">
          <el-descriptions-item label="收货人">{{ order.address?.receiver_name }}</el-descriptions-item>
          <el-descriptions-item label="电话">{{ order.address?.receiver_phone }}</el-descriptions-item>
          <el-descriptions-item label="地址">
            {{ order.address?.province }}{{ order.address?.city }}{{ order.address?.district }}{{ order.address?.detail_address }}
          </el-descriptions-item>
        </el-descriptions>

        <div style="margin-top:16px">
          <h4 style="margin-bottom:8px">订单项</h4>
          <el-table :data="order.items" border size="small">
            <el-table-column prop="product_name" label="商品" />
            <el-table-column prop="price" label="单价" width="100">
              <template #default="{ row }">¥{{ row.price }}</template>
            </el-table-column>
            <el-table-column prop="quantity" label="数量" width="80" />
            <el-table-column label="小计" width="100">
              <template #default="{ row }">¥{{ (row.price * row.quantity).toFixed(2) }}</template>
            </el-table-column>
          </el-table>
        </div>

        <div style="margin-top:20px">
          <el-button v-if="order.status === 0" type="success" @click="ship">发货</el-button>
          <el-button v-if="order.status === 1" type="warning" @click="refund">退款</el-button>
        </div>
      </template>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getOrder, shipOrder, refundOrder } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const order = ref(null)
const loading = ref(false)

function statusText(s) {
  const m = { 0: '待支付', 1: '已支付', 2: '已发货', 3: '已签收', 4: '已取消', 5: '退款中', 6: '已退款' }
  return m[s] || '未知'
}
function statusType(s) {
  const m = { 0: 'warning', 1: 'primary', 2: 'success', 3: 'success', 4: 'info', 5: 'danger', 6: 'info' }
  return m[s] || 'info'
}

async function ship() {
  try {
    await ElMessageBox.confirm('确认发货？', '提示', { type: 'info' })
    await shipOrder(order.value.id)
    ElMessage.success('发货成功')
    loadOrder()
  } catch (e) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

async function refund() {
  try {
    await ElMessageBox.confirm('确认发起退款？', '提示', { type: 'warning' })
    await refundOrder(order.value.id)
    ElMessage.success('已发起退款')
    loadOrder()
  } catch (e) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

async function loadOrder() {
  loading.value = true
  try {
    const data = await getOrder(route.params.id)
    order.value = data.order
  } finally {
    loading.value = false
  }
}

onMounted(() => loadOrder())
</script>
