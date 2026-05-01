<template>
  <MapView
    :markers="markers"
    :route-geo-j-s-o-n="routeGeoJSON"
    :clickable="clickable"
    :center="center"
    :bounds="bounds"
    :loading="loading"
    @point-selected="emit('point-selected', $event)"
  />
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import axios from 'axios'
import MapView from './MapView.vue'

// コンポーネントのprops定義
const props = defineProps({
  origin: {
    type: Object,
    default: null
  },
  dest: {
    type: Object,
    default: null
  },
  clickable: {
    type: Boolean,
    default: false
  },
  center: {
    type: Object,
    default: null
  },
  bounds: {
    type: Array,
    default: null
  }
})

const emit = defineEmits(['point-selected'])

const routeGeoJSON = ref(null)
const loading = ref(false)

// マーカー情報を出発地と目的地から生成するcomputedプロパティ
const markers = computed(() => {
  const result = []
  if (props.origin) result.push({ ...props.origin, label: '出発地', type: 'origin' })
  if (props.dest) result.push({ ...props.dest, label: '目的地', type: 'dest' })
  return result
})

// ルート情報が取得できない場合の代替として、出発地と目的地を結ぶ直線のGeoJSONを生成する関数
function straightLine(origin, dest) {
  return {
    type: 'LineString',
    coordinates: [
      [origin.lng, origin.lat],
      [dest.lng, dest.lat]
    ]
  }
}

// ルート情報をAPIから取得する関数
async function fetchRoute(origin, dest) {
  loading.value = true
  try {
    const res = await axios.get('/api/v1/routing', {
      params: { origin_lat: origin.lat, origin_lng: origin.lng, dest_lat: dest.lat, dest_lng: dest.lng }
    })
    const data = res.data
    if (data.type === 'LineString') {
      routeGeoJSON.value = data
    } else if (data.routes && data.routes[0]?.geometry) {
      routeGeoJSON.value = data.routes[0].geometry
    } else {
      routeGeoJSON.value = straightLine(origin, dest)
    }
  } catch {
    routeGeoJSON.value = straightLine(origin, dest)
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.origin, props.dest],
  ([origin, dest]) => {
    if (origin && dest) {
      fetchRoute(origin, dest)
    } else {
      routeGeoJSON.value = null
    }
  },
  { immediate: true, deep: true }
)
</script>
