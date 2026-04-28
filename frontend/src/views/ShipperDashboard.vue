<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-2xl mx-auto">
      <div class="bg-green-700 text-white rounded-xl px-8 py-6 mb-6 flex items-center justify-between">
        <div>
          <h2 class="text-xl font-bold mb-1">Shipper ダッシュボード</h2>
          <p class="text-sm opacity-80">荷物輸送管理システム</p>
        </div>
        <button @click="handleLogout"
          class="px-4 py-2 bg-white/20 border border-white/40 text-white rounded-lg text-sm hover:bg-white/30 transition-colors">
          ログアウト
        </button>
      </div>

      <div class="grid grid-cols-3 gap-4">
        <div @click="goToScheduleSearch"
          class="bg-white rounded-xl shadow p-8 cursor-pointer hover:shadow-lg transition-shadow">
          <div class="text-4xl mb-3">🔍</div>
          <h3 class="text-green-700 font-semibold mb-2">スケジュール検索</h3>
          <p class="text-sm text-gray-500">運行スケジュールを検索して予約します</p>
        </div>
        <div @click="goToBookingList"
          class="bg-white rounded-xl shadow p-8 cursor-pointer hover:shadow-lg transition-shadow">
          <div class="text-4xl mb-3">📦</div>
          <h3 class="text-green-700 font-semibold mb-2">予約一覧</h3>
          <p class="text-sm text-gray-500">自分の予約状況を確認します</p>
        </div>
        <div @click="goToTracking"
          class="bg-white rounded-xl shadow p-8 cursor-pointer hover:shadow-lg transition-shadow">
          <div class="text-4xl mb-3">📍</div>
          <h3 class="text-green-700 font-semibold mb-2">荷物追跡</h3>
          <p class="text-sm text-gray-500">追跡番号で荷物の状況を確認します</p>
        </div>
        <div @click="goToCompanyList"
          class="bg-white rounded-xl shadow p-8 cursor-pointer hover:shadow-lg transition-shadow col-span-3">
          <div class="text-4xl mb-3">🏢</div>
          <h3 class="text-green-700 font-semibold mb-2">荷物置き場を確認する</h3>
          <p class="text-sm text-gray-500">各バス会社の荷物置き場の場所・写真を確認できます</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

// スケジュール検索画面へ遷移
function goToScheduleSearch() {
    router.push({ name: 'ScheduleSearch'})
}
// 予約一覧画面へ遷移
function goToBookingList() {
    router.push({ name: 'BookingList'})
}
// 追跡画面へ遷移
function goToTracking() {
    router.push({ name: 'Tracking'})
}
// 荷物置き場確認画面へ遷移
function goToCompanyList() {
    router.push({ name: 'CompanyList'})
}
</script>
