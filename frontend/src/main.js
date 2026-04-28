import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { useAuthStore } from './stores/auth'

import axios from 'axios'
import App from './App.vue'
import router from './router'
import './style.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

const auth = useAuthStore()
if (auth.token) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${auth.token}`
}

app.mount('#app')