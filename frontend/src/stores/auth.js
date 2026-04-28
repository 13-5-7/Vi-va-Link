import { defineStore } from 'pinia'
import axios from 'axios'

import { API_PATH } from '@/const'

export const useAuthStore = defineStore('auth', {
    state: () => ({
        token: null,
        role: null,
        userId: null
    }),
    actions: {
        // ログイン処理
        async login(email, password, role) {
            const payload = { email, password, role: role }
            
            const { data } = await axios.post(API_PATH.LOGIN, payload)
            
            // stateの更新
            this.token = data.token
            this.role = data.role
            this.userId = data.user_id
            
            // ローカルストレージに値をセット
            localStorage.setItem('token', data.token)
            localStorage.setItem('role', data.role)
            localStorage.setItem('userId', data.user_id)

            // Axiosの共通ヘッダーにトークンをセット
            axios.defaults.headers.common['Authorization'] = `Bearer ${data.token}`
            return data
        },

        // ログアウト処理
        async logout() {
            try {
                await axios.post(API_PATH.LOGOUT)
            } catch (error) {
                console.error('Logout failed:', error)
            } finally {
                // Stateの初期化
                this.$reset() 
                
                // ローカルストレージの値を削除
                localStorage.removeItem('token')
                localStorage.removeItem('role')
                localStorage.removeItem('userId')
                // Axiosのヘッダーを削除
                delete axios.defaults.headers.common['Authorization']
            }
        }
    }
})