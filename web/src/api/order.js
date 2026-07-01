import http from './http'

export function createOrder(data) {
  return http.post('/orders', data)
}

export function listOrders(params) {
  return http.get('/orders', { params })
}

export function getOrder(id) {
  return http.get(`/orders/${id}`)
}

export function cancelOrder(id) {
  return http.post(`/orders/${id}/cancel`)
}
