<template>
  <div class="min-h-screen bg-gray-100 p-6 flex justify-center">
    <div class="w-full max-w-xl">
      <div class="mb-4">
        <button
          v-if="auth.role === 'bus_operator'"
          class="text-blue-700 text-sm hover:underline"
          @click="router.push('/operator/dashboard')"
        >
          ← ダッシュボードへ
        </button>
        <button
          v-else-if="auth.role === 'shipper'"
          class="text-green-700 text-sm hover:underline"
          @click="router.push('/shipper/dashboard')"
        >
          ← ダッシュボードへ
        </button>
      </div>
      <h2 class="text-2xl font-bold text-center text-gray-800 mb-6">
        荷物追跡
      </h2>

      <div class="bg-white rounded-xl shadow p-6 mb-6">
        <form @submit.prevent="handleSearch">
          <label class="block text-sm text-gray-600 mb-1">追跡番号</label>
          <div class="flex gap-3">
            <input
              v-model="trackingNumber"
              type="text"
              required
              placeholder="追跡番号を入力してください"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-lg font-mono text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
            >
            <button
              type="submit"
              :disabled="loading"
              class="px-6 py-2 bg-green-700 text-white rounded-lg text-sm whitespace-nowrap hover:bg-green-800 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
            >
              {{ loading ? '照会中...' : '照会する' }}
            </button>
          </div>
        </form>
      </div>

      <div
        v-if="errorMessage"
        class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ errorMessage }}
      </div>

      <div
        v-if="result"
        class="bg-white rounded-xl shadow p-6"
      >
        <h3 class="font-semibold text-gray-800 mb-4">
          照会結果
        </h3>
        <dl class="space-y-3 text-sm">
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">
              追跡番号:
            </dt>
            <dd class="font-mono font-bold">
              {{ result.tracking_number }}
            </dd>
          </div>
          <div class="flex gap-2 items-center">
            <dt class="text-gray-500 w-28 shrink-0">
              ステータス:
            </dt>
            <dd><BookingStatusBadge :status="result.status" /></dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">
              出発地:
            </dt>
            <dd>{{ result.schedule.origin_name }}</dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">
              目的地:
            </dt>
            <dd>{{ result.schedule.dest_name }}</dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">
              出発日時:
            </dt>
            <dd>{{ formatDate(result.schedule.depart_at) }}</dd>
          </div>
          <div class="flex gap-2">
            <dt class="text-gray-500 w-28 shrink-0">
              最終更新:
            </dt>
            <dd>{{ formatDate(result.status_updated_at) }}</dd>
          </div>
        </dl>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import BookingStatusBadge from '../components/BookingStatusBadge.vue'
import { useAuthStore } from '../stores/auth.js'
import { API_PATH } from '@/const'

const router = useRouter()
const auth = useAuthStore()
const trackingNumber = ref('')
const loading = ref(false)
const errorMessage = ref('')
const result = ref(null)
let pollTimer = null

// 検索実行：APIから最新情報を取得し、定期更新(ポーリング)を開始する
async function handleSearch() {
  loading.value = true; errorMessage.value = ''; result.value = null
  try {
    const res = await axios.get(`${API_PATH.TRACKING}/${trackingNumber.value}`)
    result.value = res.data; startPolling() // 成功時のみ自動更新を開始
  } catch (err) {
    // 404なら未登録、それ以外はシステムエラーとして扱う
    errorMessage.value = err.response?.status === 404? '指定された追跡番号が見つかりません。' : '照会に失敗しました。'
    stopPolling()
  } finally { loading.value = false }
}

// データ更新：ポーリングから呼ばれるサイレントな更新処理
async function refreshTracking() {
  if (!trackingNumber.value || !result.value) return
  try {
    const res = await axios.get(`${API_PATH.TRACKING}/${trackingNumber.value}`)
    result.value = res.data
    // 配送完了またはキャンセルになったら自動更新を停止
    if (result.value.status === 'delivered' || result.value.status === 'cancelled') stopPolling()
  } catch(err) {
    console.error(`追跡情報の更新に失敗しました。:`, err);
  }
}

// ポーリング制御：30秒間隔でステータスを確認
function startPolling() { stopPolling(); if (result.value?.status === 'delivered' || result.value?.status === 'cancelled') return; pollTimer = setInterval(refreshTracking, 30000) }
function stopPolling() { if (pollTimer) { clearInterval(pollTimer); pollTimer = null } }
// ライフサイクル管理：コンポーネント破棄時にタイマーを確実に止める（メモリリーク防止）
onUnmounted(stopPolling)
// 入力変更監視：検索番号が変わったら以前の結果とタイマーをリセット
watch(trackingNumber, () => { stopPolling(); result.value = null })

// 日時を ja-JP 形式 (yyyy/mm/dd hh:mm) にフォーマット
function formatDate(d) {
  if (!d) return '-'
  return new Date(d).toLocaleString('ja-JP', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>
