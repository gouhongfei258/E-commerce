<template>
  <div class="data-table">
    <el-button type="primary" style="margin-bottom:12px" @click="showDialog(null)">+ 新增类目</el-button>
    <el-tree :data="tree" node-key="id" default-expand-all :props="{ label: 'name', children: 'children' }">
      <template #default="{ node, data }">
        <span style="flex:1">{{ data.name }}</span>
        <span style="margin-left:12px;color:#909399;font-size:12px">排序:{{ data.sort_order }}</span>
        <el-button size="small" style="margin-left:12px" @click="showDialog(data)">+ 子类目</el-button>
        <el-button size="small" @click="editCategory(data)">编辑</el-button>
        <el-button size="small" type="danger" @click="del(data.id)">删除</el-button>
      </template>
    </el-tree>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑类目' : '新增类目'" width="420px">
      <el-form :model="cForm" label-width="80px">
        <el-form-item label="名称"><el-input v-model="cForm.name" /></el-form-item>
        <el-form-item label="图标"><el-input v-model="cForm.icon" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="cForm.sort_order" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" @click="saveCategory">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getCategoryTree, createCategory, updateCategory, deleteCategory } from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const tree = ref([])
const dialogVisible = ref(false)
const editing = ref(false)
const parentId = ref(0)
const cForm = reactive({ name: '', icon: '', sort_order: 0 })

async function fetchTree() {
  const data = await getCategoryTree()
  tree.value = data.categories || []
}

function showDialog(parent) {
  editing.value = false
  parentId.value = parent ? parent.id : 0
  cForm.name = ''; cForm.icon = ''; cForm.sort_order = 0
  dialogVisible.value = true
}

function editCategory(node) {
  editing.value = true
  parentId.value = node.parent_id || 0
  cForm.name = node.name; cForm.icon = node.icon; cForm.sort_order = node.sort_order
  dialogVisible.value = true
  cForm._id = node.id
}

async function saveCategory() {
  if (editing.value) {
    await updateCategory(cForm._id, { name: cForm.name, icon: cForm.icon, sort_order: cForm.sort_order })
  } else {
    await createCategory({ name: cForm.name, icon: cForm.icon, sort_order: cForm.sort_order, parent_id: parentId.value })
  }
  dialogVisible.value = false
  ElMessage.success('保存成功')
  fetchTree()
}

async function del(id) {
  try {
    await ElMessageBox.confirm('确认删除？', '提示', { type: 'warning' })
    await deleteCategory(id)
    ElMessage.success('删除成功')
    fetchTree()
  } catch (e) { /* cancel */ }
}

onMounted(fetchTree)
</script>
