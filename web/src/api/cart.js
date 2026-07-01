import http from './http'

export function listCartItems() {
  return http.get('/cart')
}

export function addCartItem(data) {
  return http.post('/cart/items', data)
}

export function updateCartItem(id, quantity) {
  return http.put(`/cart/items/${id}`, { quantity })
}

export function removeCartItem(id) {
  return http.delete(`/cart/items/${id}`)
}

export function checkout(data) {
  return http.post('/cart/checkout', data)
}
