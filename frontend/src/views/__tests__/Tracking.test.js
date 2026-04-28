import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import Tracking from '../Tracking.vue'

vi.mock('axios')
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))
vi.mock('../../components/BookingStatusBadge.vue', () => ({
  default: { props: ['status'], template: '<span>{{ status }}</span>' },
}))

describe('Tracking', () => {
  let pinia

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  const mountWrapper = () => mount(Tracking, { global: { plugins: [pinia] } })

  it('初期状態: 入力フォームが表示される', () => {
    const wrapper = mountWrapper()
    expect(wrapper.find('input[type="text"]').exists()).toBe(true)
    expect(wrapper.find('button[type="submit"]').text()).toBe('照会する')
  })

  it('照会成功時に追跡情報を表示', async () => {
    vi.mocked(axios.get).mockResolvedValue({
      data: {
        tracking_number: 'TRK-TEST001',
        status: 'in_transit',
        status_updated_at: '2099-01-01T10:00:00Z',
        schedule: {
          origin_name: '東京駅',
          dest_name: '大阪駅',
          depart_at: '2099-01-01T08:00:00Z',
        },
      },
    })

    const wrapper = mountWrapper()
    await wrapper.find('input[type="text"]').setValue('TRK-TEST001')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('TRK-TEST001')
    )
    expect(wrapper.text()).toContain('東京駅')
    expect(wrapper.text()).toContain('大阪駅')
  })

  it('404エラー時に「追跡番号が見つかりません」を表示', async () => {
    vi.mocked(axios.get).mockRejectedValue({ response: { status: 404 } })

    const wrapper = mountWrapper()
    await wrapper.find('input[type="text"]').setValue('TRK-NOTEXIST')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('指定された追跡番号が見つかりません')
    )
  })

  it('その他エラー時に汎用エラーメッセージを表示', async () => {
    vi.mocked(axios.get).mockRejectedValue({ response: { status: 500 } })

    const wrapper = mountWrapper()
    await wrapper.find('input[type="text"]').setValue('TRK-ERR')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('照会に失敗しました')
    )
  })

  it('照会中はボタンが無効化', async () => {
    vi.mocked(axios.get).mockReturnValue(new Promise(() => {}))

    const wrapper = mountWrapper()
    await wrapper.find('input[type="text"]').setValue('TRK-TEST')
    await wrapper.find('form').trigger('submit')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('button[type="submit"]').element.disabled).toBe(true)
    expect(wrapper.find('button[type="submit"]').text()).toBe('照会中...')
  })

  it('照会後30秒ごとに自動更新する', async () => {
    vi.mocked(axios.get).mockResolvedValue({
      data: {
        tracking_number: 'TRK-POLL',
        status: 'accepted',
        status_updated_at: '2099-01-01T00:00:00Z',
        schedule: { origin_name: 'A', dest_name: 'B', depart_at: '2099-01-01T00:00:00Z' },
      },
    })

    const wrapper = mountWrapper()
    await wrapper.find('input[type="text"]').setValue('TRK-POLL')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => expect(axios.get).toHaveBeenCalledTimes(1))

    vi.advanceTimersByTime(30_000)
    await vi.waitFor(() => expect(axios.get).toHaveBeenCalledTimes(2))
  })

  it('delivered になったら自動更新を停止する', async () => {
    vi.mocked(axios.get).mockResolvedValue({
      data: {
        tracking_number: 'TRK-DONE',
        status: 'delivered',
        status_updated_at: '2099-01-01T00:00:00Z',
        schedule: { origin_name: 'A', dest_name: 'B', depart_at: '2099-01-01T00:00:00Z' },
      },
    })

    const wrapper = mountWrapper()
    await wrapper.find('input[type="text"]').setValue('TRK-DONE')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => expect(axios.get).toHaveBeenCalledTimes(1))

    vi.advanceTimersByTime(30_000)
    // delivered なのでポーリングは追加呼び出しなし
    expect(axios.get).toHaveBeenCalledTimes(1)
  })

  it('追跡番号を変更するとポーリングがリセットされ結果がクリアされる', async () => {
    vi.mocked(axios.get).mockResolvedValue({
      data: {
        tracking_number: 'TRK-A',
        status: 'accepted',
        status_updated_at: '2099-01-01T00:00:00Z',
        schedule: { origin_name: 'A', dest_name: 'B', depart_at: '2099-01-01T00:00:00Z' },
      },
    })

    const wrapper = mountWrapper()
    await wrapper.find('input[type="text"]').setValue('TRK-A')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => expect(wrapper.text()).toContain('TRK-A'))

    // 追跡番号を変更
    await wrapper.find('input[type="text"]').setValue('TRK-B')
    await wrapper.vm.$nextTick()
    // 結果がクリアされる
    expect(wrapper.find('[class*="照会結果"]').exists()).toBe(false)
  })
})
