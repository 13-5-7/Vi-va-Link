<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-5xl mx-auto">
      <div class="flex items-center gap-4 mb-6">
        <button
          class="text-green-700 text-sm hover:underline"
          @click="router.push('/shipper/dashboard')"
        >
          ← ダッシュボードへ
        </button>
        <h2 class="text-xl font-bold text-gray-800">
          スケジュール検索
        </h2>
      </div>

      <div class="bg-white rounded-xl shadow p-6 mb-6">
        <form @submit.prevent="handleSearch">
          <div class="grid grid-cols-3 gap-4 mb-4">
            <!-- 出発地 -->
            <div class="relative">
              <label class="block text-sm text-gray-600 mb-1">
                出発地
                <span
                  v-if="geocoding.origin"
                  class="text-green-600 text-xs ml-1"
                >検索中...</span>
              </label>
              <input
                v-model="form.originName"
                type="text"
                placeholder="例: 東京駅"
                autocomplete="off"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                :class="originPoint ? 'border-green-500' : 'border-gray-300'"
                @input="onOriginNameInput"
                @blur="hideDropdown('origin', 200)"
              >
              <ul
                v-if="suggestions.origin.length"
                class="absolute top-full left-0 right-0 bg-white border border-gray-200 border-t-0 rounded-b-lg z-50 max-h-48 overflow-y-auto shadow-lg"
              >
                <li
                  v-for="s in suggestions.origin"
                  :key="s.place_id"
                  class="px-3 py-2 text-sm cursor-pointer hover:bg-green-50 border-b border-gray-100 last:border-0"
                  @mousedown.prevent="selectSuggestion('origin', s)"
                >
                  {{ s.display_name }}
                </li>
              </ul>
            </div>
            <!-- 目的地 -->
            <div class="relative">
              <label class="block text-sm text-gray-600 mb-1">
                目的地
                <span
                  v-if="geocoding.dest"
                  class="text-green-600 text-xs ml-1"
                >検索中...</span>
              </label>
              <input
                v-model="form.destName"
                type="text"
                placeholder="例: 大阪駅"
                autocomplete="off"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                :class="destPoint ? 'border-green-500' : 'border-gray-300'"
                @input="onDestNameInput"
                @blur="hideDropdown('dest', 200)"
              >
              <ul
                v-if="suggestions.dest.length"
                class="absolute top-full left-0 right-0 bg-white border border-gray-200 border-t-0 rounded-b-lg z-50 max-h-48 overflow-y-auto shadow-lg"
              >
                <li
                  v-for="s in suggestions.dest"
                  :key="s.place_id"
                  class="px-3 py-2 text-sm cursor-pointer hover:bg-green-50 border-b border-gray-100 last:border-0"
                  @mousedown.prevent="selectSuggestion('dest', s)"
                >
                  {{ s.display_name }}
                </li>
              </ul>
            </div>
            <!-- 出発日 -->
            <div>
              <label class="block text-sm text-gray-600 mb-1">出発日</label>
              <input
                v-model="form.departDate"
                type="date"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
              >
            </div>
          </div>

          <div
            v-if="originPoint || destPoint"
            class="flex gap-3 mb-4 text-sm"
          >
            <div
              v-if="originPoint"
              class="flex-1 px-3 py-1.5 bg-green-50 text-green-800 rounded"
            >
              出発地: {{ originPoint.name }}
            </div>
            <div
              v-if="destPoint"
              class="flex-1 px-3 py-1.5 bg-green-50 text-green-800 rounded"
            >
              目的地: {{ destPoint.name }}
            </div>
            <button
              type="button"
              class="px-3 py-1.5 border border-gray-300 text-gray-500 rounded text-xs hover:bg-gray-50 whitespace-nowrap"
              @click="resetPoints"
            >
              リセット
            </button>
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="px-8 py-2 bg-green-700 text-white rounded-lg text-sm hover:bg-green-800 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
          >
            {{ loading ? '検索中...' : '検索する' }}
          </button>
        </form>
      </div>

      <div
        v-if="errorMessage"
        class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ errorMessage }}
      </div>

      <div
        v-if="searched"
        class="grid grid-cols-2 gap-6"
      >
        <!-- 一覧 -->
        <div class="bg-white rounded-xl shadow overflow-hidden">
          <div
            v-if="schedules.length === 0"
            class="p-8 text-center text-gray-400"
          >
            該当するスケジュールがありません
          </div>
          <table
            v-else
            class="w-full text-xs"
          >
            <thead>
              <tr class="bg-green-700 text-white">
                <th class="px-3 py-3 text-left">
                  出発地
                </th>
                <th class="px-3 py-3 text-left">
                  目的地
                </th>
                <th class="px-3 py-3 text-left">
                  出発日時
                </th>
                <th class="px-3 py-3 text-right">
                  残重量
                </th>
                <th class="px-3 py-3 text-center">
                  状態
                </th>
                <th class="px-3 py-3" />
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="s in schedules"
                :key="s.id"
                class="border-b border-gray-100 cursor-pointer hover:bg-gray-50"
                :class="selectedSchedule?.id === s.id ? 'bg-green-50' : ''"
                @click="selectSchedule(s)"
              >
                <td class="px-3 py-3">
                  {{ s.origin_name }}
                </td>
                <td class="px-3 py-3">
                  {{ s.dest_name }}
                </td>
                <td class="px-3 py-3">
                  {{ formatDate(s.depart_at) }}
                </td>
                <td class="px-3 py-3 text-right">
                  {{ s.avail_weight_kg }}kg
                </td>
                <td class="px-3 py-3 text-center">
                  <span
                    :class="statusClass(s.status)"
                    class="px-2 py-0.5 rounded text-xs font-medium"
                  >{{ statusLabel(s.status) }}</span>
                </td>
                <td class="px-3 py-3">
                  <button
                    :disabled="s.status === 'full' || s.status === 'departed' || s.status === 'arrived' || s.status === 'cancelled'"
                    class="px-3 py-1 bg-green-700 text-white rounded text-xs hover:bg-green-800 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                    @click.stop="goToBooking(s)"
                  >
                    予約
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <!-- 地図 -->
        <div class="bg-white rounded-xl shadow p-4">
          <div class="h-80 rounded overflow-hidden">
            <RouteMap
              :origin="selectedOrigin"
              :dest="selectedDest"
              :clickable="false"
              :bounds="mapBounds"
            />
          </div>
          <p
            v-if="selectedSchedule"
            class="mt-3 text-sm text-gray-600"
          >
            <strong>{{ selectedSchedule.origin_name }}</strong> → <strong>{{ selectedSchedule.dest_name }}</strong>
          </p>
          <p
            v-else
            class="mt-3 text-sm text-gray-400 text-center"
          >
            スケジュールを選択すると経路が表示されます
          </p>
          <div
            v-if="selectedSchedule"
            class="mt-3 text-center"
          >
            <button
              class="text-xs text-green-700 underline hover:text-green-900"
              @click="router.push('/shipper/companies')"
            >
              📍 荷物置き場の場所・写真を確認する
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import RouteMap from '../components/RouteMap.vue'
import { API_PATH } from '@/const'

const router = useRouter()
const loading = ref(false)
const errorMessage = ref('')
const searched = ref(false)
const schedules = ref([])
const selectedSchedule = ref(null)
const originPoint = ref(null)
const destPoint = ref(null)
const form = reactive({ originName: '', destName: '', departDate: '' })
const suggestions = reactive({ origin: [], dest: [] })
const geocoding = reactive({ origin: false, dest: false })

// --- 地名検索ロジック (Geocoding) ---

/** 地名を短縮して表示する関数 */
function shortName(full) {
  const [mainName, ...rest] = full.split(',').map(p => p.trim())
  if (!mainName) return ''
  const isUsable = (p) => p && p !== '日本' && !/^\d/.test(p)
  const subParts = rest.filter(isUsable).slice(0, 2)
  return subParts.length > 0 ? `${mainName}（${subParts.join(', ')}）` : mainName
}

/** 地名をクエリして候補を取得する関数 */
async function geocode(query) {
  const url = `https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(query)}&format=json&limit=8&countrycodes=jp&accept-language=ja`
  const res = await fetch(url, { headers: { 'Accept-Language': 'ja' } })
  const results = await res.json()
  
  const seen = new Set()
  const filtered = []
  for (const s of results) {
    const key = `${parseFloat(s.lat).toFixed(4)},${parseFloat(s.lon).toFixed(4)}`
    if (!seen.has(key)) {
      seen.add(key)
      filtered.push({ ...s, display_name: shortName(s.display_name) })
    }
    if (filtered.length >= 5) break
  }
  return filtered
}

/** 入力に応じて候補を更新する関数（デバウンス適用） */
const updateSuggestions = debounce(async (type, query) => {
  if (query.length < 2) {
    suggestions[type] = []
    return
  }
  geocoding[type] = true
  try {
    suggestions[type] = await geocode(query)
  } finally {
    geocoding[type] = false
  }
}, 400)

// --- イベントハンドラ (Handlers) ---

function onOriginNameInput() {
  if (originPoint.value) originPoint.value = null
  updateSuggestions('origin', form.originName)
}

function onDestNameInput() {
  if (destPoint.value) destPoint.value = null
  updateSuggestions('dest', form.destName)
}

function selectSuggestion(type, s) {
  const point = { lat: parseFloat(s.lat), lng: parseFloat(s.lon), name: s.display_name }
  if (type === 'origin') {
    originPoint.value = point
    form.originName = s.display_name
  } else {
    destPoint.value = point
    form.destName = s.display_name
  }
  suggestions[type] = []
}

function hideDropdown(type, delay) {
  setTimeout(() => { suggestions[type] = [] }, delay)
}

function resetPoints() {
  originPoint.value = null
  destPoint.value = null
  form.originName = ''
  form.destName = ''
  searched.value = false
  schedules.value = []
}

// --- 検索ロジック (Search) ---

const RADIUS = 0.3
function _toBBox(p) {
  return {
    origin_lat_min: p.lat - RADIUS, origin_lat_max: p.lat + RADIUS,
    origin_lng_min: p.lng - RADIUS, origin_lng_max: p.lng + RADIUS
  }
}

async function handleSearch() {
  loading.value = true
  errorMessage.value = ''
  searched.value = false
  selectedSchedule.value = null

  try {
    let params = {}
    if (originPoint.value) {
      const bbox = {
        origin_lat_min: originPoint.value.lat - RADIUS, origin_lat_max: originPoint.value.lat + RADIUS,
        origin_lng_min: originPoint.value.lng - RADIUS, origin_lng_max: originPoint.value.lng + RADIUS
      }
      params = { ...params, ...bbox }
    }
    if (destPoint.value) {
      const bbox = {
        dest_lat_min: destPoint.value.lat - RADIUS, dest_lat_max: destPoint.value.lat + RADIUS,
        dest_lng_min: destPoint.value.lng - RADIUS, dest_lng_max: destPoint.value.lng + RADIUS
      }
      params = { ...params, ...bbox }
    }
    if (form.departDate) {
      params.depart_at_from = `${form.departDate}T00:00:00Z`
      params.depart_at_to = `${form.departDate}T23:59:59Z`
    }

    const res = await axios.get(API_PATH.SCHEDULES_SEARCH, { params })
    schedules.value = res.data.schedules || []
    searched.value = true
  } catch (err) {
    errorMessage.value = 'スケジュールの検索に失敗しました。'
  } finally {
    loading.value = false
  }
}

// --- 算出プロパティ (Computed) ---

const selectedOrigin = computed(() => 
  selectedSchedule.value ? { lat: selectedSchedule.value.origin_lat, lng: selectedSchedule.value.origin_lng } : null
)
const selectedDest = computed(() => 
  selectedSchedule.value ? { lat: selectedSchedule.value.dest_lat, lng: selectedSchedule.value.dest_lng } : null
)
const mapBounds = computed(() => {
  const s = selectedSchedule.value
  return s ? [[s.origin_lat, s.origin_lng], [s.dest_lat, s.dest_lng]] : null
})

// --- ユーティリティ (Utilities) ---

function selectSchedule(s) { selectedSchedule.value = s }
function goToBooking(s) { router.push({ name: 'BookingCreate', query: { schedule_id: s.id } }) }

function formatDate(d) {
  if (!d) return '-'
  return new Date(d).toLocaleString('ja-JP', { 
    year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' 
  })
}

function statusLabel(s) {
  const labels = { open: '受付中', full: '満載', departed: '出発済', arrived: '到着済', cancelled: '運行中止' }
  return labels[s] || s
}

function statusClass(s) {
  const classes = {
    open: 'bg-green-100 text-green-800',
    full: 'bg-orange-100 text-orange-700',
    departed: 'bg-gray-100 text-gray-600',
    arrived: 'bg-blue-100 text-blue-800',
    cancelled: 'bg-red-100 text-red-600'
  }
  return classes[s] || 'bg-gray-100 text-gray-600'
}

/** 汎用デバウンス関数 */
function debounce(fn, delay) {
  let timeoutId
  return (...args) => {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => fn(...args), delay)
  }
}
</script>
