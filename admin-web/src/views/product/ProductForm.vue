<template>
  <div>
    <el-button @click="$router.back()" style="margin-bottom:16px">← 返回</el-button>
    <el-card>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="基本信息" name="info">
          <el-form :model="form" label-width="100px" style="max-width:600px">
            <el-form-item label="商品标题" required>
              <el-input v-model="form.title" />
            </el-form-item>
            <el-form-item label="副标题">
              <el-input v-model="form.subtitle" />
            </el-form-item>
            <el-form-item label="所属类目" required>
              <el-input-number v-model="form.category_id" :min="0" />
            </el-form-item>
            <el-form-item label="所属品牌" required>
              <el-input-number v-model="form.brand_id" :min="0" />
            </el-form-item>
            <el-form-item label="状态">
              <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="上架" inactive-text="下架" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="saving" @click="save">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane v-if="isEdit" label="SKU管理" name="skus">
          <el-button type="primary" size="small" style="margin-bottom:12px" @click="showSkuDialog">+ 新增SKU</el-button>
          <el-table :data="skus" border size="small">
            <el-table-column label="规格">
              <template #default="{ row }">{{ JSON.stringify(row.attrs) }}</template>
            </el-table-column>
            <el-table-column prop="price" label="价格" width="100" />
            <el-table-column prop="stock" label="库存" width="80" />
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button size="small" @click="editSku(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="skuDialogVisible" title="SKU信息" width="500px">
      <el-form :model="skuForm" label-width="80px">
        <el-form-item label="价格"><el-input-number v-model="skuForm.price" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="原价"><el-input-number v-model="skuForm.origin_price" :min="0" :precision="2" /></el-form-item>
        <el-form-item label="库存"><el-input-number v-model="skuForm.stock" :min="0" /></el-form-item>
        <el-form-item label="编码"><el-input v-model="skuForm.code" /></el-form-item>
        <el-form-item label="图片URL"><el-input v-model="skuForm.image" /></el-form-item>
        <el-form-item label="规格属性"><el-input v-model="skuForm.attrs" placeholder='{"颜色":"红色","尺寸":"XL"}' /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="skuDialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="savingSku" @click="saveSku">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { createSPU, updateSPU, listSPUs, listSKUs, createSKUs, updateSKU } from '../../api'
import { ElMessage } from 'element-plus'

const route = useRoute()
const isEdit = computed(() => !!route.params.id)
const spuId = computed(() => parseInt(route.params.id))
const activeTab = ref('info')
const saving = ref(false)
const savingSku = ref(false)
const skus = ref([])
const skuDialogVisible = ref(false)
const editingSku = ref(null)

const form = reactive({ title: '', subtitle: '', category_id: 0, brand_id: 0, status: 0, saleable_attr_names: [] })
const skuForm = reactive({ price: 0, origin_price: 0, stock: 0, code: '', image: '', attrs: '{}' })

async function fetchSpu() {
  if (!isEdit.value) return
  const data = await listSPUs({ page: 1, page_size: 1, keyword: '' })
  const spu = (data.spus || []).find(s => s.id === spuId.value)
  if (spu) Object.assign(form, spu)
}

async function fetchSkus() {
  if (!isEdit.value) return
  const data = await listSKUs(spuId.value)
  skus.value = data.skus || []
}

async function save() {
  saving.value = true
  try {
    if (isEdit.value) {
      await updateSPU(spuId.value, { title: form.title, subtitle: form.subtitle, category_id: form.category_id, brand_id: form.brand_id, status: form.status, saleable_attr_names: form.saleable_attr_names })
      ElMessage.success('更新成功')
    } else {
      const data = await createSPU({ title: form.title, subtitle: form.subtitle, category_id: form.category_id, brand_id: form.brand_id, saleable_attr_names: form.saleable_attr_names })
      ElMessage.success('创建成功')
      // Navigate to edit
    }
  } catch (e) { ElMessage.error('保存失败') }
  finally { saving.value = false }
}

function showSkuDialog() {
  editingSku.value = null
  Object.assign(skuForm, { price: 0, origin_price: 0, stock: 0, code: '', image: '', attrs: '{}' })
  skuDialogVisible.value = true
}

function editSku(row) {
  editingSku.value = row
  Object.assign(skuForm, { price: row.price, origin_price: row.origin_price, stock: row.stock, code: row.code, image: row.image, attrs: JSON.stringify(row.attrs) })
  skuDialogVisible.value = true
}

async function saveSku() {
  savingSku.value = true
  try {
    let attrs = {}
    try { attrs = JSON.parse(skuForm.attrs) } catch (e) { ElMessage.error('规格属性格式错误'); return }
    if (editingSku.value) {
      await updateSKU(editingSku.value.id, { ...skuForm, attrs })
      ElMessage.success('更新成功')
    } else {
      await createSKUs({ spu_id: spuId.value, skus: [{ ...skuForm, attrs }] })
      ElMessage.success('创建成功')
    }
    skuDialogVisible.value = false
    fetchSkus()
  } catch (e) { ElMessage.error('保存失败') }
  finally { savingSku.value = false }
}

import { computed } from 'vue'
onMounted(() => { fetchSpu(); fetchSkus() })
</script>
