import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/admin/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { guest: true }
  },
  {
    path: '/admin',
    component: () => import('../views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/admin/dashboard' },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue')
      },
      {
        path: 'orders',
        name: 'OrderList',
        component: () => import('../views/order/OrderList.vue')
      },
      {
        path: 'orders/:id',
        name: 'OrderDetail',
        component: () => import('../views/order/OrderDetail.vue')
      },
      {
        path: 'products/spus',
        name: 'ProductList',
        component: () => import('../views/product/ProductList.vue')
      },
      {
        path: 'products/spus/new',
        name: 'ProductForm',
        component: () => import('../views/product/ProductForm.vue')
      },
      {
        path: 'products/spus/:id/edit',
        name: 'ProductEdit',
        component: () => import('../views/product/ProductForm.vue')
      },
      {
        path: 'products/categories',
        name: 'CategoryManage',
        component: () => import('../views/product/CategoryManage.vue')
      },
      {
        path: 'products/brands',
        name: 'BrandManage',
        component: () => import('../views/product/BrandManage.vue')
      },
      {
        path: 'users',
        name: 'UserList',
        component: () => import('../views/user/UserList.vue')
      },
      {
        path: 'payments',
        name: 'PaymentList',
        component: () => import('../views/payment/PaymentList.vue')
      }
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/admin' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('admin_token')
  if (to.meta.requiresAuth && !token) {
    next('/admin/login')
  } else if (to.meta.guest && token) {
    next('/admin/dashboard')
  } else {
    next()
  }
})

export default router
