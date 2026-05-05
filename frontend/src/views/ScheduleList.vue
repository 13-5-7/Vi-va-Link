<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-5xl mx-auto">
      <div class="flex items-center gap-4 mb-6">
        <button
          class="text-blue-700 text-sm hover:underline"
          @click="router.push('/operator/dashboard')"
        >
          ← ダッシュボードへ
        </button>
        <h2 class="text-xl font-bold text-gray-800">
          スケジュール一覧
        </h2>
        <button
          class="ml-auto px-4 py-2 bg-blue-700 text-white rounded-lg text-sm hover:bg-blue-800 transition-colors"
          @click="router.push('/operator/schedules/new')"
        >
          + 新規登録
        </button>
      </div>

      <div
        v-if="errorMessage"
        class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ errorMessage }}
      </div>

      <div class="grid grid-cols-2 gap-6">
        <!-- 左: 一覧 -->
        <div class="bg-white rounded-xl shadow overflow-hidden">
          <div
            v-if="loading"
            class="p-8 text-center text-gray-400"
          >
            読み込み中...
          </div>
          <div
            v-else-if="schedules.length === 0"
            class="p-8 text-center text-gray-400"
          >
            スケジュールがありません
          </div>
          <table
            v-else
            class="w-full text-sm"
          >
            <thead>
              <tr class="bg-blue-700 text-white">
                <th class="px-3 py-3 text-left">
                  出発地
                </th>
                <th class="px-3 py-3 text-left">
                  目的地
                </th>
                <th class="px-3 py-3 text-left">
                  出発日時
                </th>
                <th class="px-3 py-3 text-left">
                  状態
                </th>
                <th class="px-3 py-3 text-right">
                  残重量
                </th>
                <th class="px-3 py-3" />
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="s in schedules"
                :key="s.id"
                class="border-b border-gray-100 cursor-pointer hover:bg-gray-50"
                :class="selectedSchedule?.id === s.id ? 'bg-blue-50' : ''"
                @click="selectSchedule(s)"
              >
                <td class="px-3 py-3">
                  {{ s.origin_name }}
                </td>
                <td class="px-3 py-3">
                  {{ s.dest_name }}
                </td>
                <td class="px-3 py-3 text-xs">
                  {{ formatDate(s.depart_at) }}
                </td>
                <td class="px-3 py-3">
                  <span
                    :class="statusClass(s.status)"
                    class="px-2 py-0.5 rounded text-xs font-medium"
                  >{{ statusLabel(s.status) }}</span>
                </td>
                <td class="px-3 py-3 text-right text-xs">
                  {{ s.avail_weight_kg }}kg
                </td>
                <td
                  class="px-3 py-3"
                  @click.stop
                >
                  <button
                    v-if="s.status !== 'departed' && s.status !== 'arrived' && s.status !== 'cancelled'"
                    :disabled="s.bookings && s.bookings.filter(b => b.status !== 'cancelled').length > 0"
                    :title="s.bookings && s.bookings.filter(b => b.status !== 'cancelled').length > 0 ? '予約があるため削除できません' : '削除'"
                    class="px-2 py-1 border border-red-400 text-red-600 rounded text-xs hover:bg-red-50 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    @click="deleteSchedule(s)"
                  >
                    削除
                  </button>
                  <button
                    v-if="s.status === 'open' || s.status === 'full'"
                    class="ml-1 px-2 py-1 border border-orange-400 text-orange-600 rounded text-xs hover:bg-orange-50 transition-colors"
                    @click="cancelSchedule(s)"
                  >
                    運行中止
                  </button>
                  <span
                    v-if="s.status === 'cancelled'"
                    class="text-xs text-gray-400"
                  >中止済み</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 右: 地図 + 詳細 -->
        <div class="space-y-4">
          <div class="bg-white rounded-xl shadow p-4">
            <div class="h-64 rounded overflow-hidden">
              <RouteMap
                :origin="selectedOrigin"
                :dest="selectedDest"
                :clickable="false"
                :bounds="mapBounds"
              />
            </div>
          </div>

          <div
            v-if="selectedSchedule"
            class="bg-white rounded-xl shadow p-4"
          >
            <h3 class="font-semibold text-gray-800 mb-3 text-sm">
              スケジュール詳細
            </h3>
            <dl class="grid grid-cols-2 gap-2 text-xs text-gray-600 mb-4">
              <div>
                <dt class="font-medium">
                  出発地
                </dt><dd>{{ selectedSchedule.origin_name }}</dd>
              </div>
              <div>
                <dt class="font-medium">
                  目的地
                </dt><dd>{{ selectedSchedule.dest_name }}</dd>
              </div>
              <div>
                <dt class="font-medium">
                  出発日時
                </dt><dd>{{ formatDate(selectedSchedule.depart_at) }}</dd>
              </div>
              <div>
                <dt class="font-medium">
                  到着予定
                </dt><dd>{{ formatDate(selectedSchedule.arrive_at) }}</dd>
              </div>
              <div>
                <dt class="font-medium">
                  最大重量
                </dt><dd>{{ selectedSchedule.max_weight_kg }} kg</dd>
              </div>
              <div>
                <dt class="font-medium">
                  残重量
                </dt><dd>{{ selectedSchedule.avail_weight_kg }} kg</dd>
              </div>
              <div>
                <dt class="font-medium">
                  最大サイズ
                </dt><dd>{{ selectedSchedule.max_size_cm }} cm</dd>
              </div>
              <div>
                <dt class="font-medium">
                  ステータス
                </dt><dd>
                  <span
                    :class="statusClass(selectedSchedule.status)"
                    class="px-2 py-0.5 rounded text-xs font-medium"
                  >{{ statusLabel(selectedSchedule.status) }}</span>
                </dd>
              </div>
            </dl>

            <!-- ステータス変更 -->
            <div class="flex flex-wrap gap-2 items-center mb-4">
              <span class="text-xs text-gray-500">ステータス変更:</span>
              <button
                v-if="selectedSchedule.status === 'open'"
                class="px-3 py-1 bg-orange-600 text-white rounded text-xs hover:bg-orange-700 transition-colors"
                @click="updateScheduleStatus(selectedSchedule.id, 'full')"
              >
                満載にする
              </button>
              <button
                v-if="selectedSchedule.status === 'open' || selectedSchedule.status === 'full'"
                class="px-3 py-1 bg-gray-600 text-white rounded text-xs hover:bg-gray-700 transition-colors"
                @click="updateScheduleStatus(selectedSchedule.id, 'departed')"
              >
                出発済にする
              </button>
              <button
                v-if="selectedSchedule.status === 'departed'"
                class="px-3 py-1 bg-blue-700 text-white rounded text-xs hover:bg-blue-800 transition-colors"
                @click="updateScheduleStatus(selectedSchedule.id, 'arrived')"
              >
                到着済にする
              </button>
              <span
                v-if="selectedSchedule.status === 'arrived'"
                class="text-xs text-blue-700"
              >到着済み（変更不可）</span>
            </div>

            <!-- 予約一覧 -->
            <h3 class="font-semibold text-gray-800 mb-2 text-sm">
              予約一覧
            </h3>
            <div
              v-if="bookings.length === 0"
              class="text-xs text-gray-400"
            >
              予約はありません
            </div>
            <table
              v-else
              class="w-full text-xs"
            >
              <thead>
                <tr class="bg-gray-50">
                  <th class="px-2 py-2 text-left border-b border-gray-200">
                    追跡番号
                  </th>
                  <th class="px-2 py-2 text-right border-b border-gray-200">
                    重量
                  </th>
                  <th class="px-2 py-2 text-left border-b border-gray-200">
                    受取人
                  </th>
                  <th class="px-2 py-2 text-left border-b border-gray-200">
                    状態 / 操作
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="b in bookings"
                  :key="b.id"
                  class="border-b border-gray-100"
                >
                  <td class="px-2 py-2 font-mono">
                    {{ b.tracking_number }}
                  </td>
                  <td class="px-2 py-2 text-right">
                    {{ b.weight_kg }}kg
                  </td>
                  <td class="px-2 py-2">
                    {{ b.recipient_name }}
                  </td>
                  <td class="px-2 py-2">
                    <span :class="b.status === 'cancelled' ? 'text-gray-400 line-through' : 'text-gray-600'">{{ bookingStatusLabel(b.status) }}</span>
                    <div
                      v-if="b.status !== 'cancelled'"
                      class="flex gap-1 mt-1"
                    >
                      <button
                        v-if="b.status === 'accepted'"
                        class="px-2 py-0.5 border border-gray-300 rounded text-xs hover:bg-gray-50"
                        @click="updateBookingStatus(b.id, 'loaded')"
                      >
                        積載
                      </button>
                      <button
                        v-if="b.status === 'loaded'"
                        class="px-2 py-0.5 border border-gray-300 rounded text-xs hover:bg-gray-50"
                        @click="updateBookingStatus(b.id, 'in_transit')"
                      >
                        出発
                      </button>
                      <button
                        v-if="b.status === 'in_transit'"
                        class="px-2 py-0.5 border border-gray-300 rounded text-xs hover:bg-gray-50"
                        @click="updateBookingStatus(b.id, 'delivered')"
                      >
                        完了
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div
            v-else
            class="bg-white rounded-xl shadow p-8 text-center text-gray-400 text-sm"
          >
            スケジュールを選択すると詳細が表示されます
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import RouteMap from '../components/RouteMap.vue'
import { API_PATH } from '@/const'

const router = useRouter()
const schedules = ref([])
const loading = ref(false)
const errorMessage = ref('')
const selectedSchedule = ref(null)
const bookings = ref([])

const selectedOrigin = computed(() => selectedSchedule.value ? { lat: selectedSchedule.value.origin_lat, lng: selectedSchedule.value.origin_lng } : null)
const selectedDest = computed(() => selectedSchedule.value ? { lat: selectedSchedule.value.dest_lat, lng: selectedSchedule.value.dest_lng } : null)
const mapBounds = computed(() => { const s = selectedSchedule.value; return s ? [[s.origin_lat, s.origin_lng], [s.dest_lat, s.dest_lng]] : null })

// スケジュール取得API呼び出し
async function fetchSchedules() {
  loading.value = true; errorMessage.value = ''
  try { const res = await axios.get(API_PATH.SCHEDULES); schedules.value = res.data.schedules || [] }
  catch { errorMessage.value = 'スケジュールの取得に失敗しました。' }
  finally { loading.value = false }
}

// スケジュール選択
function selectSchedule(s) { selectedSchedule.value = s; bookings.value = s.bookings || [] }

// ヘルパー関数
function formatDate(d) {
  if (!d) return '-'
  return new Date(d).toLocaleString('ja-JP', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

// ステータス表示用のラベルとクラス
function statusLabel(s) { return { open: '受付中', full: '満載', departed: '出発済', arrived: '到着済', cancelled: '運行中止' }[s] || s }

// ステータスに応じた背景色と文字色のクラスを返す
function statusClass(s) {
  return { open: 'bg-green-100 text-green-800', full: 'bg-orange-100 text-orange-700', departed: 'bg-gray-100 text-gray-600', arrived: 'bg-blue-100 text-blue-800', cancelled: 'bg-red-100 text-red-600' }[s] || 'bg-gray-100 text-gray-600'
}

// 予約ステータスのラベル
function bookingStatusLabel(s) { return { accepted: '受付済', loaded: '積載済', in_transit: '輸送中', delivered: '配達済', cancelled: 'キャンセル' }[s] || s }

// ステータス更新の共通処理
// updateFn: APIを実行する非同期関数
async function handleStatusUpdate(updateFn) {
  try {
    await updateFn();
    // 1. 一覧を最新の状態に更新
    await fetchSchedules();
    
    // 2. 現在選択中の詳細データも最新の状態に同期
    if (selectedSchedule.value) {
      const updated = schedules.value.find(s => s.id === selectedSchedule.value.id);
      if (updated) {
        selectedSchedule.value = updated;
        bookings.value = updated.bookings || [];
      }
    }
  } catch (err) {
    console.error(err);
    alert('ステータスの更新に失敗しました。');
  }
}

// 各ボタンから呼ばれる関数
async function updateScheduleStatus(id, status) {
  await handleStatusUpdate(() => axios.patch(`${API_PATH.SCHEDULES}/${id}/status`, { status }));
}
// 各ボタンから呼ばれる関数
async function updateBookingStatus(id, status) {
  await handleStatusUpdate(() => axios.patch(`${API_PATH.BOOKINGS}/${id}/status`, { status }));
}

// スケジュール削除API呼び出し
async function deleteSchedule(s) {
  if (!confirm(`「${s.origin_name} → ${s.dest_name}」を削除しますか？`)) return
  try {
    await axios.delete(`${API_PATH.SCHEDULES}/${s.id}`)
    if (selectedSchedule.value?.id === s.id) { selectedSchedule.value = null; bookings.value = []}
    await fetchSchedules()
  } catch (err) { alert(err.response?.data?.error?.message || '削除に失敗しました。') }
}

// スケジュール運行中止API呼び出し
async function cancelSchedule(s) {
  if (!confirm(`「${s.origin_name} → ${s.dest_name}」の運行を中止しますか？\n受付済みの予約は全てキャンセルされます。`)) return
  try {
    await axios.post(`${API_PATH.SCHEDULES}/${s.id}/cancel`)
    if (selectedSchedule.value?.id === s.id) { selectedSchedule.value = null; bookings.value = [] }
    await fetchSchedules()
  } catch (err) { alert(err.response?.data?.error?.message || '運行中止に失敗しました。') }
}

// コンポーネントマウント時にスケジュールを取得
onMounted(fetchSchedules)
</script>
