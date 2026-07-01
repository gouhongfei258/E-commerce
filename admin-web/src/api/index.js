import http from './http'

export const login = (username, password) =>
  http.post('/auth/login', { username, password })

export const register = (data) =>
  http.post('/auth/register', data)

export const getDashboard = () =>
  http.get('/admin/dashboard')

export const listOrders = (params) =>
  http.get('/admin/orders', { params })

export const getOrder = (id) =>
  http.get(`/admin/orders/${id}`)

export const shipOrder = (id) =>
  http.post(`/admin/orders/${id}/ship`)

export const refundOrder = (id) =>
  http.post(`/admin/orders/${id}/refund`)

export const listSPUs = (params) =>
  http.get('/admin/spus', { params })

export const createSPU = (data) =>
  http.post('/admin/spus', data)

export const updateSPU = (id, data) =>
  http.put(`/admin/spus/${id}`, data)

export const listSKUs = (spuId) =>
  http.get(`/admin/spus/${spuId}/skus`)

export const createSKUs = (data) =>
  http.post('/admin/skus', data)

export const updateSKU = (id, data) =>
  http.put(`/admin/skus/${id}`, data)

export const getCategoryTree = () =>
  http.get('/admin/categories')

export const createCategory = (data) =>
  http.post('/admin/categories', data)

export const updateCategory = (id, data) =>
  http.put(`/admin/categories/${id}`, data)

export const deleteCategory = (id) =>
  http.delete(`/admin/categories/${id}`)

export const listBrands = (params) =>
  http.get('/admin/brands', { params })

export const createBrand = (data) =>
  http.post('/admin/brands', data)

export const listUsers = (params) =>
  http.get('/admin/users', { params })

export const listPayments = (params) =>
  http.get('/admin/payments', { params })
