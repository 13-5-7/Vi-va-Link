<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white p-8 rounded-xl shadow-md w-full max-w-md">
      <h2 class="text-2xl font-bold text-center text-gray-800 mb-6">Bus Operator ログイン</h2>

      <div v-if="errorMessage" class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm">
        {{ errorMessage }}
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label class="block text-sm text-gray-600 mb-1">メールアドレス</label>
          <input v-model="email" type="email" required placeholder="operator@example.com"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>
        <div>
          <label class="block text-sm text-gray-600 mb-1">パスワード</label>
          <input v-model="password" type="password" required placeholder="パスワードを入力"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>
        <button type="submit" :disabled="loading"
          class="w-full py-3 bg-blue-700 text-white rounded-lg text-base font-medium hover:bg-blue-800 disabled:opacity-60 disabled:cursor-not-allowed transition-colors">
          {{ loading ? 'ログイン中...' : 'ログイン' }}
        </button>
      </form>

      <div class="mt-6 text-center">
        <button @click="router.push('/operator/register')" class="text-sm text-blue-700 hover:underline">
          新規登録（招待コードをお持ちの方）
        </button>
      </div>
      <div class="mt-2 text-center">
        <router-link to="/password-reset" class="text-sm text-blue-600 hover:underline">パスワードをお忘れですか？</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'
import { ROLES } from '@/const'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

// ログインボタンを押した時の動き
const handleLogin = async () => {
  // メールアドレス、パスワード必須バリデーション
  if (!email.value || !password.value) return

  loading.value = true
  errorMessage.value = ''
  try {
    await auth.login(email.value, password.value, ROLES.OPERATOR)

    router.push({ name: 'OperatorDashboard'})
  } catch (err) {
    console.error('Login Error:', err)
    errorMessage.value = err.response?.status === 401 
      ? 'メールアドレスまたはパスワードが正しくありません。'
      : 'ログインに失敗しました。'
  } finally {
    loading.value = false
  }
}
</script>
