import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'

const routes = [
  { path: '/', redirect: '/shipper/login' },
  { path: '/operator/login', name: 'LoginOperator', component: () => import('../views/LoginOperator.vue') },
  { path: '/operator/register', name: 'RegisterOperator', component: () => import('../views/RegisterOperator.vue') },
  { path: '/shipper/login', name: 'LoginShipper', component: () => import('../views/LoginShipper.vue') },
  { path: '/operator/dashboard', name: 'OperatorDashboard', component: () => import('../views/OperatorDashboard.vue'), meta: { requiresAuth: true, role: 'bus_operator' } },
  { path: '/operator/schedules/new', name: 'ScheduleCreate', component: () => import('../views/ScheduleCreate.vue'), meta: { requiresAuth: true, role: 'bus_operator' } },
  { path: '/operator/schedules', name: 'ScheduleList', component: () => import('../views/ScheduleList.vue'), meta: { requiresAuth: true, role: 'bus_operator' } },
  { path: '/operator/mypage', name: 'OperatorMyPage', component: () => import('../views/OperatorMyPage.vue'), meta: { requiresAuth: true, role: 'bus_operator' } },
  { path: '/operator/qrscan', name: 'QRScan', component: () => import('../views/QRScan.vue'), meta: { requiresAuth: true, role: 'bus_operator' } },
  { path: '/shipper/dashboard', name: 'ShipperDashboard', component: () => import('../views/ShipperDashboard.vue'), meta: { requiresAuth: true, role: 'shipper' } },
  { path: '/shipper/schedules', name: 'ScheduleSearch', component: () => import('../views/ScheduleSearch.vue'), meta: { requiresAuth: true, role: 'shipper' } },
  { path: '/shipper/bookings', name: 'BookingList', component: () => import('../views/BookingList.vue'), meta: { requiresAuth: true, role: 'shipper' } },
  { path: '/shipper/bookings/new', name: 'BookingList', BookingCreate: () => import('../views/BookingCreate.vue'), meta: { requiresAuth: true, role: 'shipper' } },
  { path: '/shipper/companies', name: 'CompanyList', component: () => import('../views/CompanyList.vue'), meta: { requiresAuth: true, role: 'shipper' } },
  { path: '/tracking', name: 'Tracking', component: () => import('../views/TrackingView.vue') },
  { path: '/password-reset', name: 'PasswordResetRequest', component: () => import('../views/PasswordResetRequest.vue') },
  { path: '/password-reset/confirm', name: 'PasswordResetConfirm', component: () => import('../views/PasswordResetConfirm.vue') },
]

const router = createRouter({
    history: createWebHistory(),
    routes,
})

router.beforeEach((to, _from) => {
    const auth = useAuthStore()

    if (!to.meta.requiresAuth) {
        return true
    }

    const loginPath = to.meta.role === 'bus_operator' 
        ? '/operator/login' 
        : '/shipper/login'

    if (!auth.token || (to.meta.role && auth.role !== to.meta.role)) {
        return loginPath
    }

    return true
})
export default router
