<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  if (!username.value || !password.value) {
    error.value = 'Please enter username and password'
    return
  }
  error.value = ''
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    const redirect = route.query.redirect || '/'
    router.push(redirect)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card card">
      <h2>Login</h2>
      <p class="text-secondary mb-16">Welcome back! Please login to your account.</p>

      <div v-if="error" class="alert alert-error">{{ error }}</div>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label>Username</label>
          <input v-model="username" class="form-input" placeholder="Enter username" />
        </div>
        <div class="form-group">
          <label>Password</label>
          <input v-model="password" class="form-input" type="password" placeholder="Enter password" />
        </div>
        <button type="submit" class="btn btn-primary btn-lg btn-full" :disabled="loading">
          {{ loading ? 'Logging in...' : 'Login' }}
        </button>
      </form>

      <p class="text-center mt-16 text-secondary">
        Don't have an account?
        <router-link to="/register">Register</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 56px - 48px);
}

.auth-card {
  width: 100%;
  max-width: 420px;
  padding: 32px;
}

.auth-card h2 {
  font-size: 28px;
  margin-bottom: 4px;
}

.btn-full {
  width: 100%;
}
</style>
