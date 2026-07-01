import http from './http'

export function createPayment(data) {
  return http.post('/payments', data)
}

export function processPayment(id) {
  return http.post(`/payments/${id}/process`)
}

export function getPayment(id) {
  return http.get(`/payments/${id}`)
}

export function getPaymentByOrder(orderNo) {
  return http.get(`/payments/by-order/${orderNo}`)
}
