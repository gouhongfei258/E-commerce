<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const email = ref('')
const error = ref('')
const loading = ref(false)

async function handleRegister() {
  if (!username.value || !password.value || !email.value) {
    error.value = 'Please fill in all fields'
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }
  error.value = ''
  loading.value = true
  try {
    await auth.register(username.value, password.value, email.value)
    router.push('/')
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
      <h2>Register</h2>
      <p class="text-secondary mb-16">Create a new account</p>

      <div v-if="error" class="alert alert-error">{{ error }}</div>

      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label>Username</label>
          <input v-model="username" class="form-input" placeholder="Enter username" />
        </div>
        <div class="form-group">
          <label>Email</label>
          <input v-model="email" class="form-input" type="email" placeholder="Enter email" />
        </div>
        <div class="form-group">
          <label>Password</label>
          <input v-model="password" class="form-input" type="password" placeholder="Enter password" />
        </div>
        <div class="form-group">
          <label>Confirm Password</label>
          <input v-model="confirmPassword" class="form-input" type="password" placeholder="Confirm password" />
        </div>
        <button type="submit" class="btn btn-primary btn-lg btn-full" :disabled="loading">
          {{ loading ? 'Registering...' : 'Register' }}
        </button>
      </form>

      <p class="text-center mt-16 text-secondary">
        Already have an account?
        <router-link to="/login">Login</router-link>
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
