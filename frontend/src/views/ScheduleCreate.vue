<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-3xl mx-auto">
      <div class="flex items-center gap-4 mb-6">
        <button
          class="text-blue-700 text-sm hover:underline"
          @click="router.push('/operator/dashboard')"
        >
          ← ダッシュボードへ
        </button>
        <h2 class="text-xl font-bold text-gray-800">
          スケジュール登録
        </h2>
      </div>

      <div
        v-if="successMessage"
        class="bg-green-50 border border-green-300 text-green-800 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ successMessage }}
      </div>
      <div
        v-if="errorMessage"
        class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded mb-4 text-sm"
      >
        {{ errorMessage }}
      </div>

      <!-- 地図 -->
      <div class="bg-white rounded-xl shadow p-6 mb-6">
        <p class="text-sm text-gray-500 mb-3">
          地名を入力して候補から選択するか、地図をクリックして出発地・目的地を選択してください
        </p>
        <div class="h-96 rounded overflow-hidden">
          <RouteMap
            :origin="origin"
            :dest="dest"
            :clickable="!submitted"
            :center="mapCenter"
            @point-selected="handlePointSelected"
          />
        </div>
        <div class="flex gap-4 mt-3 text-sm">
          <div class="flex-1 px-3 py-2 bg-blue-50 rounded">
            <strong>出発地:</strong>
            <span v-if="origin"> {{ origin.name || `${origin.lat.toFixed(4)}, ${origin.lng.toFixed(4)}` }}</span>
            <span
              v-else
              class="text-gray-400"
            > 未選択</span>
          </div>
          <div class="flex-1 px-3 py-2 bg-pink-50 rounded">
            <strong>目的地:</strong>
            <span v-if="dest"> {{ dest.name || `${dest.lat.toFixed(4)}, ${dest.lng.toFixed(4)}` }}</span>
            <span
              v-else
              class="text-gray-400"
            > 未選択</span>
          </div>
        </div>
        <div
          v-if="origin || dest"
          class="mt-2"
        >
          <button
            class="text-sm text-gray-500 border border-gray-300 px-3 py-1 rounded hover:bg-gray-50"
            @click="resetPoints"
          >
            地点をリセット
          </button>
        </div>
      </div>

      <!-- フォーム -->
      <div class="bg-white rounded-xl shadow p-6">
        <h3 class="font-semibold text-gray-800 mb-4">
          スケジュール情報
        </h3>
        <form
          class="space-y-4"
          @submit.prevent="handleSubmit"
        >
          <div class="grid grid-cols-2 gap-4">
            <!-- 出発地名 -->
            <div class="relative">
              <label class="block text-sm text-gray-600 mb-1">
                出発地名
                <span
                  v-if="geocoding.origin"
                  class="text-blue-600 text-xs ml-1"
                >検索中...</span>
              </label>
              <input
                v-model="form.originName"
                type="text"
                required
                placeholder="例: 東京駅"
                autocomplete="off"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                :class="origin ? 'border-blue-500' : 'border-gray-300'"
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
                  class="px-3 py-2 text-sm cursor-pointer hover:bg-blue-50 border-b border-gray-100 last:border-0"
                  @mousedown.prevent="selectSuggestion('origin', s)"
                >
                  {{ s.display_name }}
                </li>
              </ul>
            </div>
            <!-- 目的地名 -->
            <div class="relative">
              <label class="block text-sm text-gray-600 mb-1">
                目的地名
                <span
                  v-if="geocoding.dest"
                  class="text-blue-600 text-xs ml-1"
                >検索中...</span>
              </label>
              <input
                v-model="form.destName"
                type="text"
                required
                placeholder="例: 大阪駅"
                autocomplete="off"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-pink-500"
                :class="dest ? 'border-pink-500' : 'border-gray-300'"
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
                  class="px-3 py-2 text-sm cursor-pointer hover:bg-pink-50 border-b border-gray-100 last:border-0"
                  @mousedown.prevent="selectSuggestion('dest', s)"
                >
                  {{ s.display_name }}
                </li>
              </ul>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm text-gray-600 mb-1">出発日時</label>
              <input
                v-model="form.departAt"
                type="datetime-local"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
            </div>
            <div>
              <label class="block text-sm text-gray-600 mb-1">到着予定日時</label>
              <input
                v-model="form.arriveAt"
                type="datetime-local"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm text-gray-600 mb-1">積載可能重量 (kg) <span class="text-red-500 text-xs">※1個最大10kg</span></label>
              <input
                v-model.number="form.maxWeightKg"
                type="number"
                min="0.1"
                max="10000"
                step="0.1"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
              <p class="text-xs text-gray-400 mt-1">
                バス全体の積載可能重量（例: 100kg）
              </p>
            </div>
            <div>
              <label class="block text-sm text-gray-600 mb-1">積載可能サイズ (cm) <span class="text-red-500 text-xs">※1個最大140cm</span></label>
              <input
                v-model.number="form.maxSizeCm"
                type="number"
                min="0.1"
                max="140"
                step="0.1"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                :class="form.maxSizeCm > 140 ? 'border-red-400 bg-red-50' : ''"
              >
              <p
                v-if="form.maxSizeCm > 140"
                class="text-red-500 text-xs mt-1"
              >
                システム上限は140cmです
              </p>
              <p
                v-else
                class="text-xs text-gray-400 mt-1"
              >
                3辺合計の上限（最大140cm）
              </p>
            </div>
          </div>

          <button
            type="submit"
            :disabled="loading || !origin || !dest || submitted || form.maxSizeCm > 140"
            class="w-full py-3 bg-blue-700 text-white rounded-lg font-medium hover:bg-blue-800 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
          >
            {{ loading ? '登録中...' : '登録する' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import RouteMap from '../components/RouteMap.vue'

const router = useRouter()
const origin = ref(null)
const dest = ref(null)
const clickCount = ref(0)
const submitted = ref(false)
const loading = ref(false)
const successMessage = ref('')
const errorMessage = ref('')
const mapCenter = ref(null)
const form = reactive({ originName: '', destName: '', departAt: '', arriveAt: '', maxWeightKg: '', maxSizeCm: '' })
const suggestions = reactive({ origin: [], dest: [] })
const geocoding = reactive({ origin: false, dest: false })
const _debounceTimers = { origin: null, dest: null }

// 地名をクエリして候補を取得する関数
async function geocode(query) {
  const url = `https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(query)}&format=json&limit=8&countrycodes=jp&accept-language=ja`
  const res = await fetch(url, { headers: { 'Accept-Language': 'ja' } })
  const results = await res.json()
  const seen = new Set()
  const filtered = []
  // 結果を重複排除しつつ、地名を短縮して表示する
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

// 地名を短縮して表示する関数
function shortName(full) {
  const [mainName, ...rest] = full.split(',').map(p => p.trim());
  
  if (!mainName) return '';

  const isUsable = (p) => p && p !== '日本' && !/^\d/.test(p);

  const subParts = rest
    .filter(isUsable)
    .slice(0, 2);

  return subParts.length > 0 
    ? `${mainName}（${subParts.join(', ')}）` 
    : mainName;
}

// 入力に応じて候補を更新する関数（デバウンス付き）
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

// 出発地の入力
function onOriginNameInput() {
  if (origin.value?.fromGeocode) origin.value = null
  updateSuggestions('origin', form.originName)
}
// 目的地の入力
function onDestNameInput() {
  if (dest.value?.fromGeocode) dest.value = null
  updateSuggestions('dest', form.destName)
}

// 候補が選択されたときの処理
function selectSuggestion(type, s) {
  const point = { lat: parseFloat(s.lat), lng: parseFloat(s.lon), name: s.display_name, fromGeocode: true }
  if (type === 'origin') {
    origin.value = point; form.originName = s.display_name; suggestions.origin = []
    if (!dest.value) clickCount.value = 1
    mapCenter.value = { lat: point.lat, lng: point.lng }
  } else {
    dest.value = point; form.destName = s.display_name; suggestions.dest = []
    if (!origin.value) clickCount.value = 0
  }
}

// ドロップダウンを非表示にする関数
function hideDropdown(type, delay) { setTimeout(() => { suggestions[type] = [] }, delay) }

// 地図上で地点が選択されたときの処理
function handlePointSelected(point) {
  if (submitted.value) return
  if (clickCount.value === 0) { origin.value = { ...point, fromGeocode: false }; clickCount.value = 1; mapCenter.value = { lat: point.lat, lng: point.lng } }
  else if (clickCount.value === 1) { dest.value = { ...point, fromGeocode: false }; clickCount.value = 2 }
}

// 地点選択をリセットする関数
function resetPoints() { origin.value = null; dest.value = null; clickCount.value = 0; submitted.value = false; successMessage.value = ''; errorMessage.value = '' }

// フォームの送信処理
async function handleSubmit() {
  if (!origin.value || !dest.value) { errorMessage.value = '出発地と目的地を選択してください。'; return }
  loading.value = true; errorMessage.value = ''; successMessage.value = ''
  try {
    await axios.post('/api/v1/schedules', {
      origin_lat: origin.value.lat, origin_lng: origin.value.lng, origin_name: form.originName,
      dest_lat: dest.value.lat, dest_lng: dest.value.lng, dest_name: form.destName,
      depart_at: new Date(form.departAt).toISOString(), arrive_at: new Date(form.arriveAt).toISOString(),
      max_weight_kg: form.maxWeightKg, max_size_cm: form.maxSizeCm,
    })
    submitted.value = true; successMessage.value = 'スケジュールを登録しました。'
  } catch (err) { errorMessage.value = err.response?.data?.message || 'スケジュールの登録に失敗しました。' }
  finally { loading.value = false }
}

// デバウンス関数
function debounce(fn, delay) {
  let timeoutId;
  return (...args) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => fn(...args), delay);
  };
}
</script>
