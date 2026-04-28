<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-2xl mx-auto">
      <div class="bg-blue-700 text-white rounded-xl px-8 py-6 mb-6 flex items-center justify-between">
        <div>
          <h2 class="text-xl font-bold mb-1">Operator ダッシュボード</h2>
          <p class="text-sm opacity-80">バス運行管理システム</p>
        </div>
        <button @click="handleLogout"
          class="px-4 py-2 bg-white/20 border border-white/40 text-white rounded-lg text-sm hover:bg-white/30 transition-colors">
          ログアウト
        </button>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div @click="goToScheduleCreate"
          class="bg-white rounded-xl shadow p-8 cursor-pointer hover:shadow-lg transition-shadow">
          <div class="text-4xl mb-3">🗓️</div>
          <h3 class="text-blue-700 font-semibold mb-2">スケジュール登録</h3>
          <p class="text-sm text-gray-500">新しい運行スケジュールを登録します</p>
        </div>
        <div @click="goToScheduleList"
          class="bg-white rounded-xl shadow p-8 cursor-pointer hover:shadow-lg transition-shadow">
          <div class="text-4xl mb-3">📋</div>
          <h3 class="text-blue-700 font-semibold mb-2">スケジュール一覧</h3>
          <p class="text-sm text-gray-500">登録済みスケジュールと予約状況を確認します</p>
        </div>
        <div @click="goToOperatorMyPage"
          class="bg-white rounded-xl shadow p-8 cursor-pointer hover:shadow-lg transition-shadow col-span-2">
          <div class="text-4xl mb-3">🏢</div>
          <h3 class="text-blue-700 font-semibold mb-2">マイページ</h3>
          <p class="text-sm text-gray-500">所属会社の情報・荷物置き場の案内を管理します</p>
        </div>
        <div @click="goToQRScan"
          class="bg-white rounded-xl shadow p-8 cursor-pointer hover:shadow-lg transition-shadow col-span-2">
          <div class="text-4xl mb-3">📷</div>
          <h3 class="text-blue-700 font-semibold mb-2">QRスキャン</h3>
          <p class="text-sm text-gray-500">荷物のQRコードを読み取ってステータスを更新します</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth.js'

const router = useRouter()
const auth = useAuthStore()

// ログアウト処理を実行し画面遷移
function handleLogout() {
    auth.logout()
    router.push({ name: 'LoginOperator'})
}
// スケジュール登録画面へ遷移
function goToScheduleCreate(){
    router.push({ name: 'ScheduleCreate'})
}
// スケジュール一覧画面へ遷移
function goToScheduleList(){
    router.push({ name: 'ScheduleList'})
}
// オペレーター専用マイページへ遷移
function goToOperatorMyPage(){
    router.push({ name: 'OperatorMyPage'})
}
// QRスキャン画面へ遷移
function goToQRScan(){
    router.push({ name: 'QRScan'})
}
</script>
