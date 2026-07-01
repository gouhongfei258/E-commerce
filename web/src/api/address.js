import http from './http'

export function listAddresses() {
  return http.get('/addresses')
}

export function createAddress(data) {
  return http.post('/addresses', data)
}

export function updateAddress(id, data) {
  return http.put(`/addresses/${id}`, data)
}

export function deleteAddress(id) {
  return http.delete(`/addresses/${id}`)
}

export function setDefaultAddress(id) {
  return http.put(`/addresses/${id}/default`)
}
