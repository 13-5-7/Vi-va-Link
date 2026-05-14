<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white p-8 rounded-xl shadow-md w-full max-w-md">
      <h2 class="text-2xl font-bold text-center text-gray-800 mb-2">
        Bus Operator 新規登録
      </h2>
      <p class="text-sm text-center text-gray-500 mb-6">
        バス会社から発行された招待コードが必要です
      </p>

      <div
        v-if="errorMessage"
        class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ errorMessage }}
      </div>
      <div
        v-if="successMessage"
        class="bg-green-50 border border-green-300 text-green-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ successMessage }}
      </div>

      <form
        v-if="!successMessage"
        class="space-y-4"
        @submit.prevent="handleRegister"
      >
        <div>
          <label class="block text-sm text-gray-600 mb-1">招待コード <span class="text-red-500">*</span></label>
          <input
            v-model="inviteCode"
            type="text"
            required
            placeholder="例: OKINAWA-2024"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono uppercase"
          >
          <p class="text-xs text-gray-400 mt-1">
            所属バス会社から発行されたコードを入力してください
          </p>
        </div>
        <div>
          <label class="block text-sm text-gray-600 mb-1">メールアドレス <span class="text-red-500">*</span></label>
          <input
            v-model="email"
            type="email"
            required
            placeholder="operator@example.com"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <div>
          <label class="block text-sm text-gray-600 mb-1">パスワード <span class="text-red-500">*</span></label>
          <input
            v-model="password"
            type="password"
            required
            placeholder="8文字以上"
            minlength="8"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <button
          type="submit"
          :disabled="loading"
          class="w-full py-3 bg-blue-700 text-white rounded-lg text-base font-medium hover:bg-blue-800 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
        >
          {{ loading ? '登録中...' : '登録する' }}
        </button>
      </form>

      <div class="mt-6 text-center">
        <button
          class="text-sm text-blue-700 hover:underline"
          @click="router.push('/operator/login')"
        >
          ← ログイン画面へ戻る
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_PATH } from '@/const'

const router = useRouter()
const inviteCode = ref('')
const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

// 招待コードを利用したバス会社担当者（オペレーター）の登録処理
// 前提：管理画面等で事前に有効な招待コードが生成されていること
async function handleRegister() {
  loading.value = true
  errorMessage.value = ''
  try {
    await axios.post(API_PATH.AUTH_REGISTER, {
      email: email.value,
      password: password.value,
      role: 'bus_operator',
      invite_code: inviteCode.value.trim().toUpperCase(),
    })
    successMessage.value = '登録が完了しました。ログイン画面からサインインしてください。'
  } catch (err) {
    const code = err.response?.data?.error?.code
    if (code === 'INVALID_INVITE_CODE') {
      errorMessage.value = '招待コードが無効または使用済みです。バス会社にお問い合わせください。'
    } else if (code === 'EMAIL_ALREADY_EXISTS') {
      errorMessage.value = 'このメールアドレスはすでに登録されています。'
    } else {
      errorMessage.value = '登録に失敗しました。しばらく経ってから再度お試しください。'
    }
  } finally {
    loading.value = false
  }
}
</script>
