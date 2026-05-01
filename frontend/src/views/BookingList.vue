<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-4xl mx-auto">
      <div class="flex items-center gap-4 mb-6">
        <button
          class="text-green-700 text-sm hover:underline"
          @click="router.push('/shipper/dashboard')"
        >
          ← ダッシュボードへ
        </button>
        <h2 class="text-xl font-bold text-gray-800">
          予約一覧
        </h2>
      </div>

      <div
        v-if="errorMessage"
        class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ errorMessage }}
      </div>

      <div class="bg-white rounded-xl shadow overflow-hidden">
        <div
          v-if="loading"
          class="p-8 text-center text-gray-400"
        >
          読み込み中...
        </div>
        <div
          v-else-if="bookings.length === 0"
          class="p-8 text-center text-gray-400"
        >
          予約がありません
        </div>
        <table
          v-else
          class="w-full text-sm"
        >
          <thead>
            <tr class="bg-green-700 text-white">
              <th class="px-4 py-3 text-left">
                追跡番号
              </th>
              <th class="px-4 py-3 text-left">
                スケジュール
              </th>
              <th class="px-4 py-3 text-center">
                ステータス
              </th>
              <th class="px-4 py-3 text-left">
                登録日時
              </th>
              <th class="px-4 py-3" />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="b in bookings"
              :key="b.id"
              class="border-b border-gray-100 hover:bg-gray-50"
            >
              <td class="px-4 py-3 font-mono text-xs">
                {{ b.tracking_number }}
              </td>
              <td class="px-4 py-3">
                <span v-if="scheduleMap[b.schedule_id]">
                  {{ scheduleMap[b.schedule_id].origin_name }} → {{ scheduleMap[b.schedule_id].dest_name }}
                </span>
                <span
                  v-else
                  class="text-gray-400 text-xs"
                >{{ b.schedule_id }}</span>
              </td>
              <td class="px-4 py-3 text-center">
                <BookingStatusBadge :status="b.status" />
              </td>
              <td class="px-4 py-3 text-xs text-gray-500">
                {{ formatDate(b.created_at) }}
              </td>
              <td class="px-4 py-3 text-center">
                <button
                  v-if="b.status === 'accepted'"
                  class="px-3 py-1 border border-red-400 text-red-600 rounded text-xs hover:bg-red-50 transition-colors"
                  @click="cancelBooking(b)"
                >
                  キャンセル
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>

</script>
