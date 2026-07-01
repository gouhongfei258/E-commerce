<script setup>
import { useAuthStore } from './stores/auth'
import { useCartStore } from './stores/cart'

const auth = useAuthStore()
const cart = useCartStore()
</script>

<template>
  <div v-if="auth.isLoggedIn" class="navbar">
    <div class="navbar-inner">
      <router-link to="/" class="navbar-brand">Mall</router-link>
      <nav class="navbar-nav">
        <router-link to="/">Products</router-link>
        <router-link to="/cart">Cart ({{ cart.totalCount }})</router-link>
        <router-link to="/orders">Orders</router-link>
        <router-link to="/addresses">Addresses</router-link>
      </nav>
      <div class="navbar-right">
        <span class="username">{{ auth.user?.username }}</span>
        <router-link to="/profile" class="btn btn-sm btn-outline">Profile</router-link>
        <button class="btn btn-sm btn-outline" @click="auth.logout">Logout</button>
      </div>
    </div>
  </div>
  <div class="main-content">
    <router-view />
  </div>
</template>
