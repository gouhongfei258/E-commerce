<script setup>
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
</script>

<template>
  <div>
    <div class="page-header">
      <h2>Profile</h2>
    </div>

    <div class="card profile-card">
      <div class="avatar">
        {{ (auth.user?.username || 'U').charAt(0).toUpperCase() }}
      </div>
      <div class="profile-info">
        <div class="info-row">
          <span class="info-label">Username</span>
          <span class="info-value">{{ auth.user?.username || '-' }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">Email</span>
          <span class="info-value">{{ auth.user?.email || '-' }}</span>
        </div>
        <div class="info-row" v-if="auth.user?.user_id">
          <span class="info-label">User ID</span>
          <span class="info-value">{{ auth.user?.user_id }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-card {
  max-width: 600px;
}

.avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--primary);
  color: white;
  font-size: 32px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
}

.profile-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-row {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: var(--bg);
  border-radius: var(--radius);
}

.info-label {
  font-size: 14px;
  color: var(--text-secondary);
  min-width: 100px;
}

.info-value {
  font-size: 15px;
  font-weight: 500;
}
</style>
