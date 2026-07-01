<template>
  <div class="filter-bar">
    <el-input v-model="keyword" placeholder="搜索用户名/手机号" style="width:260px" clearable @keyup.enter="search" />
    <el-button type="primary" style="margin-left:8px" @click="search">搜索</el-button>
  </div>
  <div class="data-table">
    <el-table :data="users" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="username" label="用户名" width="150" />
      <el-table-column prop="phone" label="手机号" width="140" />
      <el-table-column prop="email" label="邮箱" min-width="180" />
      <el-table-column label="角色" width="80">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'primary'" size="small">{{ row.role || 'user' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="注册时间" width="160" />
    </el-table>
    <div class="pagination-wrap">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50]" layout="total,sizes,prev,pager,next" @change="fetchUsers" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { listUsers } from '../../api'
const users = ref([]); const total = ref(0); const loading = ref(false)
const page = ref(1); const pageSize = ref(10); const keyword = ref('')

async function fetchUsers() {
  loading.value = true
  try {
    const data = await listUsers({ page: page.value, page_size: pageSize.value, keyword: keyword.value })
    users.value = data.users || []
    total.value = data.total || 0
  } finally { loading.value = false }
}
function search() { page.value = 1; fetchUsers() }
onMounted(fetchUsers)
</script>
