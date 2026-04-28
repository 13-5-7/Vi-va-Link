import { vi } from 'vitest'

// localStorage のモック（jsdom の --localstorage-file 警告を回避）
const localStorageMock = (() => {
  let store = {}
  return {
    getItem: (key) => store[key] ?? null,
    setItem: (key, value) => { store[key] = String(value) },
    removeItem: (key) => { delete store[key] },
    clear: () => { store = {} },
  }
})()
Object.defineProperty(globalThis, 'localStorage', { value: localStorageMock, writable: true })

// leaflet は jsdom で動かないためモック
vi.mock('leaflet', () => ({}))
