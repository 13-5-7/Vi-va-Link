<template>
  <div class="min-h-screen bg-gray-100 p-6">
    <div class="max-w-lg mx-auto">
      <div class="flex items-center gap-4 mb-6">
        <button @click="router.push('/operator/schedules')" class="text-blue-700 text-sm hover:underline">← スケジュール一覧へ</button>
        <h2 class="text-xl font-bold text-gray-800">QRスキャン</h2>
      </div>

      <div class="bg-white rounded-xl shadow p-6 mb-4">
        <p class="text-sm text-gray-500 mb-4 text-center">
          荷物のQRコードをカメラに向けてください。<br>ステータスが自動で更新されます。
        </p>

        <!-- スキャナー -->
        <div v-if="!result && !scanning" class="text-center">
          <button @click="startScan"
            class="px-8 py-3 bg-blue-700 text-white rounded-lg font-medium hover:bg-blue-800 transition-colors">
            📷 スキャン開始
          </button>
        </div>

        <div v-if="scanning && !result">
          <QRScanner scanner-id="operator-qr-scanner" @scanned="onScanned" />
          <button @click="stopScan" class="mt-3 w-full py-2 border border-gray-300 text-gray-500 rounded-lg text-sm hover:bg-gray-50">
            キャンセル
          </button>
        </div>
      </div>

      <!-- スキャン結果 -->
      <div v-if="processing" class="bg-white rounded-xl shadow p-6 text-center text-gray-400">
        処理中...
      </div>

      <div v-if="result" class="bg-white rounded-xl shadow p-6">
        <div v-if="result.success">
          <div class="text-center mb-4">
            <div class="text-5xl mb-2">✅</div>
            <p class="font-bold text-gray-800">ステータスを更新しました</p>
          </div>
          <dl class="space-y-2 text-sm">
            <div class="flex justify-between py-2 border-b border-gray-100">
              <dt class="text-gray-500">追跡番号</dt>
              <dd class="font-mono font-bold">{{ result.tracking_number }}</dd>
            </div>
            <div class="flex justify-between py-2 border-b border-gray-100">
              <dt class="text-gray-500">更新前</dt>
              <dd><span :class="statusClass(result.old_status)" class="px-2 py-0.5 rounded text-xs font-medium">{{ statusLabel(result.old_status) }}</span></dd>
            </div>
            <div class="flex justify-between py-2">
              <dt class="text-gray-500">更新後</dt>
              <dd><span :class="statusClass(result.new_status)" class="px-2 py-0.5 rounded text-xs font-medium">{{ statusLabel(result.new_status) }}</span></dd>
            </div>
          </dl>
          <div class="mt-4 p-3 bg-blue-50 rounded-lg text-sm text-blue-700 text-center">
            {{ actionMessage(result.new_status) }}
          </div>
        </div>

        <div v-else>
          <div class="text-center mb-4">
            <div class="text-5xl mb-2">⚠️</div>
            <p class="font-bold text-gray-800">{{ result.error }}</p>
          </div>
        </div>

        <div class="flex gap-3 mt-6">
          <button @click="reset" class="flex-1 py-2 border border-gray-300 text-gray-600 rounded-lg text-sm hover:bg-gray-50">
            続けてスキャン
          </button>
          <button @click="router.push('/operator/schedules')" class="flex-1 py-2 bg-blue-700 text-white rounded-lg text-sm hover:bg-blue-800">
            一覧へ戻る
          </button>
        </div>
      </div>

      <!-- 手動入力 -->
      <div class="bg-white rounded-xl shadow p-4 mt-4">
        <p class="text-xs text-gray-500 mb-2">カメラが使えない場合は追跡番号を直接入力</p>
        <div class="flex gap-2">
          <input v-model="manualInput" type="text" placeholder="TRK-XXXXXXXX"
            class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <button @click="onScanned(manualInput)" :disabled="!manualInput"
            class="px-4 py-2 bg-blue-700 text-white rounded-lg text-sm hover:bg-blue-800 disabled:opacity-50">
            実行
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>

</script>
