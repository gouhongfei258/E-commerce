<template>
  <div class="admin-layout">
    <div class="admin-sidebar">
      <div class="logo">管理后台</div>
      <el-menu
        :default-active="activeMenu"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        router
      >
        <el-menu-item index="/admin/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/admin/orders">
          <el-icon><Document /></el-icon>
          <span>订单管理</span>
        </el-menu-item>
        <el-sub-menu index="product">
          <template #title>
            <el-icon><Goods /></el-icon>
            <span>商品管理</span>
          </template>
          <el-menu-item index="/admin/products/spus">商品列表</el-menu-item>
          <el-menu-item index="/admin/products/categories">类目管理</el-menu-item>
          <el-menu-item index="/admin/products/brands">品牌管理</el-menu-item>
        </el-sub-menu>
        <el-menu-item index="/admin/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/payments">
          <el-icon><Money /></el-icon>
          <span>支付记录</span>
        </el-menu-item>
      </el-menu>
    </div>
    <div class="admin-main">
      <div class="admin-header">
        <span style="margin-right:12px">{{ authStore.user?.username }}</span>
        <el-button type="danger" size="small" @click="handleLogout">退出</el-button>
      </div>
      <div class="admin-content">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/admin/products/spus')) return '/admin/products/spus'
  if (path.startsWith('/admin/products/categories')) return '/admin/products/categories'
  if (path.startsWith('/admin/products/brands')) return '/admin/products/brands'
  return path
})

function handleLogout() {
  authStore.logout()
  router.push('/admin/login')
}
</script>
