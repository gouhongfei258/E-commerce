import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { listCartItems, addCartItem, updateCartItem, removeCartItem } from '../api/cart'

export const useCartStore = defineStore('cart', () => {
  const items = ref([])
  const loading = ref(false)

  const totalCount = computed(() =>
    items.value.reduce((sum, item) => sum + item.quantity, 0)
  )

  const totalPrice = computed(() =>
    items.value.reduce((sum, item) => sum + item.price * item.quantity, 0)
  )

  async function fetchItems() {
    loading.value = true
    try {
      const res = await listCartItems()
      items.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  async function addItem(skuId, quantity) {
    await addCartItem({ sku_id: skuId, quantity })
    await fetchItems()
  }

  async function updateQuantity(id, quantity) {
    await updateCartItem(id, quantity)
    await fetchItems()
  }

  async function removeItem(id) {
    await removeCartItem(id)
    await fetchItems()
  }

  function clearLocal() {
    items.value = []
  }

  return { items, loading, totalCount, totalPrice, fetchItems, addItem, updateQuantity, removeItem, clearLocal }
})
