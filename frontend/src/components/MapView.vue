<template>
  <div class="relative h-full w-full">
    <div ref="mapContainer" class="h-full w-full"></div>
    <!-- ローディングオーバーレイ -->
    <div v-if="loading" class="absolute inset-0 bg-white/60 flex items-center justify-center z-[1000] pointer-events-none">
      <div class="flex items-center gap-2 bg-white px-4 py-2 rounded-full shadow text-sm text-gray-600">
        <svg class="animate-spin h-4 w-4 text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"/>
        </svg>
        経路を取得中...
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import L from 'leaflet'

const props = defineProps({
    markers: { type: Array, default: () => [] },
    routeGeoJSON: { type: Object, default: null },
    clickable: { type: Boolean, default: false },
    center: { type: Object, default: null },
    bounds: { type: Array, default: null },
    loading: { type: Boolean, default: false }
})

const emit = defineEmits(['point-selected'])

const mapContainer = ref(null)
let map = null
let markerLayers = []
let routeLayer = null

const ICON_COLORS = {
    origin: 'blue',
    dest: 'red'
}

function createIcon(type) {
    const color = ICON_COLORS[type] || 'gray'
    return L.divIcon({
        className: '',
        html: `<div style="
            width: 14px; height: 14px;
            background: ${color};
            border: 2px solid white;
            border-radius: 50%;
            box-shadow: 0 0 4px rgba(0,0,0,0.4);
        "></div>`,
        iconSize: [14, 14],
        iconAnchor: [7, 7]
    })
}

function clearMarkers() {
    markerLayers.forEach(m => m.remove())
    markerLayers = []
}

function renderMarkers(markers) {
    if (!map) return
    clearMarkers()
    markers.forEach(({ lat, lng, label, type }) => {
        const marker = L.marker([lat, lng], { icon: createIcon(type) })
        if (label) marker.bindTooltip(label, { permanent: false })
        marker.addTo(map)
      markerLayers.push(marker)
    })
}

function renderRoute(geojson) {
    if (!map) return
    if (routeLayer) {
        routeLayer.remove()
        routeLayer = null
    }
    if (!geojson) return
    routeLayer = L.geoJSON(geojson, {
        style: { color: '#3b82f6', weight: 4, opacity: 0.8 }
    }).addTo(map)
}

onMounted(() => {
    map = L.map(mapContainer.value).setView([35.6812, 139.7671], 10)

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(map)

    if (props.clickable) {
        map.on('click', (e) => {
            emit('point-selected', { lat: e.latlng.lat, lng: e.latlng.lng })
        })
    }
    
    renderMarkers(props.markers)
    renderRoute(props.routeGeoJSON)
})

onUnmounted(() => {
    if (map) {
        map.remove()
        map = null
    }
})

watch(() => props.markers, (val) => renderMarkers(val), { deep: true })
watch(() => props.routeGeoJSON, (val) => renderRoute(val))
watch(() => props.center, (val) => {
    if (map && val) map.flyTo([val.lat, val.lng], 12, { duration: 1 })
})
watch(() => props.bounds, (val) => {
    if (map && val) map.flyToBounds(val, { padding: [40, 40], duration: 1 })
})
</script>
