<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white p-8 rounded-xl shadow-md w-full max-w-md">
      <h2 class="text-2xl font-bold text-center text-gray-800 mb-2">
        パスワードリセット
      </h2>
      <p class="text-sm text-gray-500 text-center mb-6">
        登録済みのメールアドレスを入力してください。<br>リセット用のリンクをお送りします。
      </p>

      <div
        v-if="successMessage"
        class="bg-green-50 border border-green-300 text-green-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ successMessage }}
      </div>
      <div
        v-if="errorMessage"
        class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ errorMessage }}
      </div>

      <form
        v-if="!successMessage"
        class="space-y-4"
        @submit.prevent="handleRequest"
      >
        <div>
          <label class="block text-sm text-gray-600 mb-1">メールアドレス</label>
          <input
            v-model="email"
            type="email"
            required
            placeholder="example@email.com"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <button
          type="submit"
          :disabled="loading"
          class="w-full py-3 bg-blue-600 text-white rounded-lg text-base font-medium hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
        >
          {{ loading ? '送信中...' : 'リセットリンクを送信' }}
        </button>
      </form>

      <div class="mt-4 text-center text-sm text-gray-500">
        <router-link
          to="/shipper/login"
          class="text-blue-600 hover:underline"
        >
          ログイン画面に戻る
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { API_PATH } from '@/const'

const email = ref('')
const loading = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

// パスワードリセットリクエストメールを送信するAPI呼び出し
async function handleRequest() {
  loading.value = true
  errorMessage.value = ''
  try {
    await axios.post(API_PATH.AUTH_PASSWORD_RESET_REQUEST, { email: email.value })
    successMessage.value = 'パスワードリセットの手順をメールで送信しました（登録済みの場合）。'
  } catch (err) {
    if (err.response?.status === 429) {
      errorMessage.value = 'リクエストが多すぎます。しばらく待ってから再試行してください。'
    } else {
      errorMessage.value = '送信に失敗しました。しばらく待ってから再試行してください。'
    }
  } finally {
    loading.value = false
  }
}
</script>
