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
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import BookingStatusBadge from '../components/BookingStatusBadge.vue'
import { API_PATH } from '@/const'

const router = useRouter()
const bookings = ref([])
const loading = ref(false)
const errorMessage = ref('')
const scheduleMap = ref({})
let pollTimer = null

// 予約一覧を取得し、付随するスケジュール情報も更新する
async function fetchBookings() {
  if (bookings.value.length === 0) loading.value = true
  errorMessage.value = ''
  try {
    const res = await axios.get(API_PATH.BOOKINGS)
    bookings.value = res.data.bookings || []
    await fetchSchedules()
  } catch (err) {
    const status = err.response?.status
    const msg = err.response?.data?.error?.message || err.message
    errorMessage.value = `予約一覧の取得に失敗しました。(${status}: ${msg})`
  }
  finally { loading.value = false }
}

// 予約に関連するスケジュール詳細を個別に取得してマッピングする
async function fetchSchedules() {
  // const ids = {...new Set(bookings.value.map(b => b.schedule_id))}
  const ids = Array.from(new Set(bookings.value.map(b => b.schedule_id)))
  await Promise.all(ids.map(async id => {
    try {
      const res = await axios.get(`${API_PATH.SCHEDULES}/${id}`); scheduleMap.value[id] = res.data
    } catch(err) {
      console.error(`スケジュールの取得に失敗しました(ID: ${id}):`, err);
    }
  }))
}

// 日時を ja-JP 形式 (yyyy/mm/dd hh:mm) にフォーマット
function formatDate(d) {
  if (!d) return '-'
  return new Date(d).toLocaleString('ja-JP', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

// 指定された予約を削除し、一覧を再取得する
async function cancelBooking(b) {
  if (!confirm(`追跡番号「${b.tracking_number}」の予約をキャンセルしますか？`)) return
  try {
    await axios.delete(`${API_PATH.BOOKINGS}/${b.id}`)
    await fetchBookings()
  } catch (err) {
    const code = err.response?.data?.error?.code
    if (code === 'CANNOT_CANCEL') alert('この予約はキャンセルできません（積載済み以降は不可）。')
    else if (code === 'FORBIDDEN') alert('この予約をキャンセルする権限がありません。')
    else alert('キャンセルに失敗しました。')
  }
}

// コンポーネントのマウント時に初回実行と60秒毎の定期更新（ポーリング）を設定
onMounted(() => { fetchBookings(); pollTimer = setInterval(fetchBookings, 60000) })
// アンマウント時にタイマーを破棄してメモリリークを防止
onUnmounted(() => clearInterval(pollTimer))
</script>
