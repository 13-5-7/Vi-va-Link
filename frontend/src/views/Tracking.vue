<template>
  <div class="min-h-screen bg-gray-100 p-6 flex justify-center">
    <div class="w-full max-w-xl">
      <div class="mb-4">
        <button v-if="auth.role === 'bus_operator'" @click="router.push('/operator/dashboard')"
          class="text-blue-700 text-sm hover:underline">← ダッシュボードへ</button>
        <button v-else-if="auth.role === 'shipper'" @click="router.push('/shipper/dashboard')"
          class="text-green-700 text-sm hover:underline">← ダッシュボードへ</button>
      </div>
      <h2 class="text-2xl font-bold text-center text-gray-800 mb-6">荷物追跡</h2>

      <div class="bg-white rounded-xl shadow p-6 mb-6">
        <form @submit.prevent="handleSearch">
          <label class="block text-sm text-gray-600 mb-1">追跡番号</label>
          <div class="flex gap-3">
            <input v-model="trackingNumber" type="text" required placeholder="追跡番号を入力してください"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-lg font-mono text-sm focus:outline-none focus:ring-2 focus:ring-green-500" />
            <button type="submit" :disabled="loading"
              class="px-6 py-2 bg-green-700 text-white rounded-lg text-sm whitespace-nowrap hover:bg-green-800 disabled:opacity-60 disabled:cursor-not-allowed transition-colors">
              {{ loading ? '照会中...' : '照会する' }}
            </button>
          </div>
        </form>
      </div>

      <div v-if="errorMessage" class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm">{{ errorMessage }}</div>

      <div v-if="result" class="bg-white rounded-xl shadow p-6">
        <h3 class="font-semibold text-gray-800 mb-4">照会結果</h3>
        <dl class="space-y-3 text-sm">
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">追跡番号:</dt>
            <dd class="font-mono font-bold">{{ result.tracking_number }}</dd>
          </div>
          <div class="flex gap-2 items-center">
            <dt class="text-gray-500 w-28 shrink-0">ステータス:</dt>
            <dd><BookingStatusBadge :status="result.status" /></dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">出発地:</dt>
            <dd>{{ result.schedule.origin_name }}</dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">目的地:</dt>
            <dd>{{ result.schedule.dest_name }}</dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">出発日時:</dt>
            <dd>{{ formatDate(result.schedule.depart_at) }}</dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">最終更新:</dt>
            <dd>{{ formatDate(result.status_updated_at) }}</dd>
          </div>
        </dl>
      </div>
    </div>
  </div>
</template>

<script setup>

</script>
