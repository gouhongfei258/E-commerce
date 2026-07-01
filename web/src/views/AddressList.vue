<script setup>
import { ref, onMounted } from 'vue'
import { listAddresses, createAddress, updateAddress, deleteAddress, setDefaultAddress } from '../api/address'

const addresses = ref([])
const loading = ref(true)
const error = ref('')
const showForm = ref(false)
const editingId = ref(null)

const form = ref({
  receiver_name: '',
  receiver_phone: '',
  province: '',
  city: '',
  district: '',
  detail_address: '',
  is_default: false,
})

const submitting = ref(false)

onMounted(fetchAddresses)

async function fetchAddresses() {
  loading.value = true
  try {
    const res = await listAddresses()
    addresses.value = res.data || []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  form.value = { receiver_name: '', receiver_phone: '', province: '', city: '', district: '', detail_address: '', is_default: false }
  showForm.value = true
}

function openEdit(addr) {
  editingId.value = addr.id
  form.value = {
    receiver_name: addr.receiver_name,
    receiver_phone: addr.receiver_phone,
    province: addr.province,
    city: addr.city,
    district: addr.district,
    detail_address: addr.detail_address,
    is_default: addr.is_default,
  }
  showForm.value = true
}

function cancelForm() {
  showForm.value = false
  editingId.value = null
}

async function handleSubmit() {
  if (!form.value.receiver_name || !form.value.receiver_phone || !form.value.detail_address) {
    error.value = 'Please fill in required fields'
    return
  }
  submitting.value = true
  error.value = ''
  try {
    if (editingId.value) {
      await updateAddress(editingId.value, form.value)
    } else {
      await createAddress(form.value)
    }
    showForm.value = false
    editingId.value = null
    await fetchAddresses()
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id) {
  if (!confirm('Are you sure you want to delete this address?')) return
  try {
    await deleteAddress(id)
    await fetchAddresses()
  } catch (e) {
    error.value = e.message
  }
}

async function handleSetDefault(id) {
  try {
    await setDefaultAddress(id)
    await fetchAddresses()
  } catch (e) {
    error.value = e.message
  }
}
</script>

<template>
  <div>
    <div class="page-header">
      <h2>Addresses</h2>
      <button class="btn btn-primary" @click="openCreate">Add Address</button>
    </div>

    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="showForm" class="card mb-16">
      <h3 class="mb-16">{{ editingId ? 'Edit Address' : 'New Address' }}</h3>
      <form @submit.prevent="handleSubmit">
        <div class="grid-2">
          <div class="form-group">
            <label>Receiver Name *</label>
            <input v-model="form.receiver_name" class="form-input" placeholder="Name" />
          </div>
          <div class="form-group">
            <label>Phone *</label>
            <input v-model="form.receiver_phone" class="form-input" placeholder="Phone" />
          </div>
        </div>
        <div class="grid-3">
          <div class="form-group">
            <label>Province</label>
            <input v-model="form.province" class="form-input" placeholder="Province" />
          </div>
          <div class="form-group">
            <label>City</label>
            <input v-model="form.city" class="form-input" placeholder="City" />
          </div>
          <div class="form-group">
            <label>District</label>
            <input v-model="form.district" class="form-input" placeholder="District" />
          </div>
        </div>
        <div class="form-group">
          <label>Detail Address *</label>
          <input v-model="form.detail_address" class="form-input" placeholder="Street, building, room" />
        </div>
        <div class="form-group">
          <label class="checkbox-label">
            <input type="checkbox" v-model="form.is_default" />
            Set as default address
          </label>
        </div>
        <div class="flex gap-8">
          <button type="submit" class="btn btn-primary" :disabled="submitting">
            {{ submitting ? 'Saving...' : 'Save' }}
          </button>
          <button type="button" class="btn btn-outline" @click="cancelForm">Cancel</button>
        </div>
      </form>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <div v-else-if="addresses.length === 0 && !showForm" class="empty-state">
      <p>No addresses yet</p>
      <button class="btn btn-primary" @click="openCreate">Add Address</button>
    </div>

    <div v-else class="address-list">
      <div v-for="addr in addresses" :key="addr.id" class="card address-card">
        <div class="address-main">
          <div class="address-info">
            <div class="address-name">
              <strong>{{ addr.receiver_name }}</strong> &nbsp; {{ addr.receiver_phone }}
              <span v-if="addr.is_default" class="badge badge-info ml-8">Default</span>
            </div>
            <p class="text-secondary text-sm mt-4">
              {{ addr.province }} {{ addr.city }} {{ addr.district }} {{ addr.detail_address }}
            </p>
          </div>
          <div class="address-actions">
            <button v-if="!addr.is_default" class="btn btn-sm btn-outline" @click="handleSetDefault(addr.id)">
              Set Default
            </button>
            <button class="btn btn-sm btn-outline" @click="openEdit(addr)">Edit</button>
            <button class="btn btn-sm btn-danger" @click="handleDelete(addr.id)">Delete</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.address-card {
  transition: none;
}

.address-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.address-name {
  font-size: 15px;
}

.address-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 400 !important;
  cursor: pointer;
}

.ml-8 {
  margin-left: 8px;
}

@media (max-width: 768px) {
  .address-main {
    flex-direction: column;
  }
}
</style>
