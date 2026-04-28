<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-xl mx-auto">
      <div class="flex items-center gap-4 mb-6">
        <button @click="router.push('/shipper/schedules')" class="text-green-700 text-sm hover:underline">← スケジュール検索へ</button>
        <h2 class="text-xl font-bold text-gray-800">予約登録</h2>
      </div>

      <!-- 予約完了 + QRコード -->
      <div v-if="trackingNumber" class="bg-white rounded-xl shadow p-6 text-center">
        <div class="text-green-600 text-4xl mb-3">✅</div>
        <p class="font-bold text-lg text-gray-800 mb-1">予約が完了しました</p>
        <p class="text-sm text-gray-500 mb-4">このQRコードを荷物に貼り付けてください</p>

        <div class="bg-gray-50 rounded-xl p-6 mb-4 inline-block">
          <QRCodeDisplay :value="trackingNumber" :size="220" />
        </div>

        <p class="font-mono text-lg font-bold text-gray-700 mb-1">{{ trackingNumber }}</p>
        <p class="text-xs text-gray-400 mb-4">荷物置き場にQRコードが見えるよう置いてください</p>

        <div class="bg-green-50 border border-green-200 rounded-lg px-4 py-3 mb-6 text-sm text-green-800">
          📍 <button @click="router.push('/shipper/companies')" class="underline hover:text-green-900">荷物置き場の場所・写真を確認する</button>
        </div>

        <div class="flex gap-3 justify-center">
          <button @click="printQR"
            class="px-5 py-2 border border-gray-300 text-gray-600 rounded-lg text-sm hover:bg-gray-50 transition-colors">
            🖨️ 印刷
          </button>
          <button @click="router.push('/shipper/bookings')"
            class="px-5 py-2 bg-green-700 text-white rounded-lg text-sm hover:bg-green-800 transition-colors">
            予約一覧へ
          </button>
        </div>
      </div>

      <div v-if="errorMessage" class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm">{{ errorMessage }}</div>

      <div v-if="!trackingNumber" class="bg-white rounded-xl shadow p-6">
        <div v-if="scheduleId" class="bg-gray-50 rounded px-3 py-2 mb-6 text-xs text-gray-500">
          スケジュールID: <span class="font-mono">{{ scheduleId }}</span>
        </div>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm text-gray-600 mb-1">重量 (kg) <span class="text-red-500 text-xs">※最大10kg</span></label>
              <input v-model.number="form.weightKg" type="number" min="0.1" max="10" step="0.1" required
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                :class="form.weightKg > 10 ? 'border-red-400 bg-red-50' : ''" />
              <p v-if="form.weightKg > 10" class="text-red-500 text-xs mt-1">10kg以下にしてください</p>
            </div>
            <div>
              <label class="block text-sm text-gray-600 mb-1">3辺合計 (cm) <span class="text-red-500 text-xs">※最大140cm</span></label>
              <input v-model.number="form.sizeCm" type="number" min="0.1" max="140" step="0.1" required
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                :class="form.sizeCm > 140 ? 'border-red-400 bg-red-50' : ''" />
              <p v-if="form.sizeCm > 140" class="text-red-500 text-xs mt-1">140cm以下にしてください</p>
            </div>
          </div>
          <div>
            <label class="block text-sm text-gray-600 mb-1">内容物の概要</label>
            <input v-model="form.contentDesc" type="text" required placeholder="例: 衣類、書籍など"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500" />
          </div>
          <div>
            <label class="block text-sm text-gray-600 mb-1">受取人名</label>
            <input v-model="form.recipientName" type="text" required placeholder="例: 山田 太郎"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500" />
          </div>
          <div>
            <label class="block text-sm text-gray-600 mb-1">受取人電話番号</label>
            <input v-model="form.recipientPhone" type="tel" required placeholder="例: 090-1234-5678"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500" />
          </div>
          <div>
            <label class="block text-sm text-gray-600 mb-1">受取人住所</label>
            <input v-model="form.recipientAddr" type="text" required placeholder="例: 東京都渋谷区..."
              class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500" />
          </div>
          <button type="submit" :disabled="loading || form.weightKg > 10 || form.sizeCm > 140"
            class="w-full py-3 bg-green-700 text-white rounded-lg font-medium hover:bg-green-800 disabled:opacity-60 disabled:cursor-not-allowed transition-colors">
            {{ loading ? '予約中...' : '予約する' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>

</script>
