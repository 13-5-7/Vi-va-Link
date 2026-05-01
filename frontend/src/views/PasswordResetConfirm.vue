<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white p-8 rounded-xl shadow-md w-full max-w-md">
      <h2 class="text-2xl font-bold text-center text-gray-800 mb-2">
        新しいパスワードを設定
      </h2>
      <p class="text-sm text-gray-500 text-center mb-6">
        新しいパスワードを入力してください（8文字以上）。
      </p>

      <div
        v-if="successMessage"
        class="bg-green-50 border border-green-300 text-green-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ successMessage }}
        <div class="mt-2">
          <router-link
            to="/shipper/login"
            class="text-green-700 font-medium underline"
          >
            ログイン画面へ
          </router-link>
        </div>
      </div>
      <div
        v-if="errorMessage"
        class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ errorMessage }}
      </div>

      <div
        v-if="!token"
        class="text-center text-red-600 text-sm"
      >
        無効なリセットリンクです。パスワードリセットを再度お申し込みください。
      </div>

      <form
        v-else-if="!successMessage"
        class="space-y-4"
        @submit.prevent="handleReset"
      >
        <div>
          <label class="block text-sm text-gray-600 mb-1">新しいパスワード</label>
          <input
            v-model="newPassword"
            type="password"
            required
            placeholder="8文字以上"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <div>
          <label class="block text-sm text-gray-600 mb-1">パスワード（確認）</label>
          <input
            v-model="confirmPassword"
            type="password"
            required
            placeholder="もう一度入力"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-base focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <button
          type="submit"
          :disabled="loading"
          class="w-full py-3 bg-blue-600 text-white rounded-lg text-base font-medium hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
        >
          {{ loading ? '更新中...' : 'パスワードを更新' }}
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

</script>
