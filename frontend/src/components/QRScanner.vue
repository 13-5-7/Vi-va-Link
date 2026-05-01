<template>
  <div class="relative">
    <video
      ref="video"
      class="w-full rounded-lg bg-black"
      autoplay
      playsinline
      muted
    />
    <canvas
      ref="canvas"
      class="hidden"
    />
    <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
      <div class="w-52 h-52 border-2 border-white rounded-lg opacity-60" />
    </div>
    <p
      v-if="error"
      class="text-red-500 text-sm mt-2"
    >
      {{ error }}
    </p>
  </div>
</template>

<script setup>

</script>
