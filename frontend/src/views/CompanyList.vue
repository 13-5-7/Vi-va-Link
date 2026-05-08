<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-3xl mx-auto">
      <div class="flex items-center gap-4 mb-6">
        <button
          class="text-green-700 text-sm hover:underline"
          @click="router.push('/shipper/dashboard')"
        >
          ← ダッシュボードへ
        </button>
        <h2 class="text-xl font-bold text-gray-800">
          バス会社・荷物置き場一覧
        </h2>
      </div>

      <p class="text-sm text-gray-500 mb-6">
        各バス会社の荷物置き場の場所・写真を確認できます。予約前に必ずご確認ください。
      </p>

      <div
        v-if="loading"
        class="bg-white rounded-xl shadow p-8 text-center text-gray-400"
      >
        読み込み中...
      </div>
      <div
        v-else-if="errorMessage"
        class="bg-red-50 border border-red-300 text-red-700 px-4 py-3 rounded text-sm"
      >
        {{ errorMessage }}
      </div>
      <div
        v-else-if="companies.length === 0"
        class="bg-white rounded-xl shadow p-8 text-center text-gray-400"
      >
        バス会社情報がありません
      </div>

      <div
        v-else
        class="space-y-4"
      >
        <div
          v-for="c in companies"
          :key="c.id"
          class="bg-white rounded-xl shadow overflow-hidden"
        >
          <div class="px-6 py-4 border-b border-gray-100 flex items-center gap-3">
            <span class="text-2xl">🚌</span>
            <h3 class="font-bold text-gray-800 text-lg">
              {{ c.name }}
            </h3>
          </div>

          <div class="p-6">
            <div
              v-if="c.storage_image_url || c.storage_description"
              class="grid grid-cols-1 gap-4"
              :class="c.storage_image_url ? 'md:grid-cols-2' : ''"
            >
              <!-- 画像 -->
              <div v-if="c.storage_image_url">
                <p class="text-xs text-gray-500 mb-2 font-medium">
                  荷物置き場の写真
                </p>
                <img
                  :src="c.storage_image_url"
                  alt="荷物置き場"
                  class="w-full max-h-56 object-cover rounded-lg border border-gray-200"
                >
              </div>
              <!-- 説明 -->
              <div v-if="c.storage_description">
                <p class="text-xs text-gray-500 mb-2 font-medium">
                  置き場の説明
                </p>
                <p class="text-sm text-gray-700 leading-relaxed whitespace-pre-wrap">
                  {{ c.storage_description }}
                </p>
              </div>
            </div>
            <div
              v-else
              class="text-sm text-gray-400 text-center py-4"
            >
              荷物置き場の情報はまだ登録されていません
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_PATH } from '@/const'

const router = useRouter()
const companies = ref([])
const loading = ref(true)
const errorMessage = ref('')

// 画面初期表示に必要なバス会社一覧を取得し、stateを更新する
async function fetchCompanies() {
  try {
    const res = await axios.get(API_PATH.COMPANIES)
    companies.value = res.data.companies || []
  } catch {
    errorMessage.value =  'バス会社情報の取得に失敗しました。'
  } finally {
    loading.value = false
  }
}

// 初期表示時にリストが空だとユーザーが混乱するため、マウント時に取得
onMounted(fetchCompanies)
</script>
