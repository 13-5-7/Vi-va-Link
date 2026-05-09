<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-2xl mx-auto">
      <div class="flex items-center gap-4 mb-6">
        <button
          class="text-blue-700 text-sm hover:underline"
          @click="router.push('/operator/dashboard')"
        >
          ← ダッシュボードへ
        </button>
        <h2 class="text-xl font-bold text-gray-800">
          マイページ
        </h2>
      </div>

      <div
        v-if="loading"
        class="bg-white rounded-xl shadow p-8 text-center text-gray-400"
      >
        読み込み中...
      </div>
      <div
        v-else-if="company"
        class="space-y-6"
      >
        <!-- 会社情報 -->
        <div class="bg-white rounded-xl shadow p-6">
          <h3 class="font-semibold text-gray-800 mb-4">
            所属バス会社
          </h3>
          <div class="flex items-center gap-3">
            <span class="text-3xl">🚌</span>
            <span class="text-xl font-bold text-blue-700">{{ company.name }}</span>
          </div>
        </div>

        <!-- 荷物置き場情報 -->
        <div class="bg-white rounded-xl shadow p-6">
          <h3 class="font-semibold text-gray-800 mb-4">
            荷物置き場の案内
          </h3>
          <p class="text-sm text-gray-500 mb-4">
            荷主が迷わず安全な場所へ届けられるよう、荷物置き場の写真と説明を登録してください。
          </p>

          <!-- 現在の画像 -->
          <div
            v-if="company.storage_image_url"
            class="mb-4"
          >
            <p class="text-xs text-gray-500 mb-2">
              現在登録中の画像:
            </p>
            <img
              :src="company.storage_image_url"
              alt="荷物置き場"
              class="w-full max-h-64 object-cover rounded-lg border border-gray-200"
            >
          </div>
          <div
            v-else
            class="mb-4 h-32 bg-gray-50 border-2 border-dashed border-gray-200 rounded-lg flex items-center justify-center text-gray-400 text-sm"
          >
            画像未登録
          </div>

          <form
            class="space-y-4"
            @submit.prevent="handleSave"
          >
            <!-- 画像アップロード -->
            <div>
              <label class="block text-sm text-gray-600 mb-1">画像を選択</label>
              <input
                type="file"
                accept="image/*"
                class="w-full text-sm text-gray-500 file:mr-3 file:py-2 file:px-4 file:rounded-lg file:border-0 file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                @change="onFileChange"
              >
              <p class="text-xs text-gray-400 mt-1">
                JPG / PNG / WebP（最大 2MB）
              </p>
            </div>

            <!-- プレビュー -->
            <div
              v-if="previewURL"
              class="rounded-lg overflow-hidden border border-gray-200"
            >
              <img
                :src="previewURL"
                alt="プレビュー"
                class="w-full max-h-48 object-cover"
              >
            </div>

            <!-- 説明文 -->
            <div>
              <label class="block text-sm text-gray-600 mb-1">置き場の説明</label>
              <textarea
                v-model="form.description"
                rows="3"
                placeholder="例: バスターミナル正面入口を入って左手、カウンター横のスペースです。"
                class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
              />
            </div>

            <div
              v-if="saveMessage"
              class="text-sm"
              :class="saveError ? 'text-red-600' : 'text-green-700'"
            >
              {{ saveMessage }}
            </div>

            <button
              type="submit"
              :disabled="saving"
              class="w-full py-2.5 bg-blue-700 text-white rounded-lg text-sm font-medium hover:bg-blue-800 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
            >
              {{ saving ? '保存中...' : '保存する' }}
            </button>
          </form>
        </div>
      </div>

      <div
        v-else
        class="bg-white rounded-xl shadow p-8 text-center text-gray-400"
      >
        会社情報が見つかりません
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { API_PATH } from '@/const'

const router = useRouter()
const loading = ref(true)
const saving = ref(false)
const company = ref(null)
const previewURL = ref('')
const saveMessage = ref('')
const saveError = ref(false)
const form = reactive({ description: '', imageBase64: '' })

// 自社の基本情報取得APIを呼び出し、フォームの初期値を設定する
// 取得失敗時は画面を空の状態にする（エラーハンドリングは今後の課題）
async function fetchCompany() {
  try {
    const res = await axios.get(API_PATH.COMPANIES_ME)
    company.value = res.data
    form.description = res.data.storage_description || ''
  } catch {
    // 取得失敗時は情報をクリア。ユーザーへの通知が必要な場合はここに実装
    company.value = null
  } finally {
    loading.value = false
  }
}

// 添付画像のバリデーションとプレビュー生成
// 2MB制限をチェックし、Base64形式で保持する
function onFileChange(e) {
  const file = e.target.files[0]
  if (!file) return
  if (file.size > 2 * 1024 * 1024) {
    saveMessage.value = '画像は2MB以下にしてください。'
    saveError.value = true
    return
  }
  const reader = new FileReader()
  reader.onload = (ev) => {
    previewURL.value = ev.target.result
    form.imageBase64 = ev.target.result // Data URL をそのまま保存
  }
  reader.readAsDataURL(file)
}

// マイページ登録API処理呼び出し
// 編集内容を保存。画像が未選択の場合は既存のURLを維持する
async function handleSave() {
  saving.value = true
  saveMessage.value = ''
  saveError.value = false
  try {
    const res = await axios.patch(API_PATH.COMPANIES_ME_STORAGE, {
      storage_image_url: form.imageBase64 || company.value?.storage_image_url || '',
      storage_description: form.description,
    })
    company.value = res.data
    previewURL.value = ''
    saveMessage.value = '保存しました。'
  } catch {
    saveMessage.value = '保存に失敗しました。'
    saveError.value = true
  } finally {
    saving.value = false
  }
}

// 初期表示に必要な自社プロフィールの取得（失敗しても画面は表示させる）
onMounted(fetchCompany)
</script>
