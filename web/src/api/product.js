import http from './http'

export function getCategories() {
  return http.get('/categories')
}

export function listBrands(params) {
  return http.get('/brands', { params })
}

export function listSPUs(params) {
  return http.get('/spus', { params })
}

export function getSPU(id) {
  return http.get(`/spus/${id}`)
}

export function listSKUs(params) {
  return http.get('/skus', { params })
}
