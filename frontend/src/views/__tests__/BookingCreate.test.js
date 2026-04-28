import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import BookingCreate from '../BookingCreate.vue'

vi.mock('axios')
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
  useRoute: () => ({ query: { schedule_id: 'sch-uuid-001' } }),
}))

// QRCodeDisplay は jsdom では動作しないためスタブ化
vi.mock('../../components/QRCodeDisplay.vue', () => ({
  default: { template: '<div data-testid="qr-code"></div>' },
}))

describe('BookingCreate', () => {
  let pinia

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()
  })

  const mountWrapper = () => mount(BookingCreate, { global: { plugins: [pinia] } })

  it('フォームの初期状態が正しい', () => {
    const wrapper = mountWrapper()
    expect(wrapper.find('button[type="submit"]').text()).toBe('予約する')
    expect(wrapper.find('button[type="submit"]').element.disabled).toBe(false)
  })

  it('スケジュールIDが表示される', async () => {
    const wrapper = mountWrapper()
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('sch-uuid-001')
  })

  it('予約成功時にQRコードと追跡番号を表示', async () => {
    vi.mocked(axios.post).mockResolvedValue({
      data: { tracking_number: 'TRK-ABCD1234' },
    })

    const wrapper = mountWrapper()
    await wrapper.find('input[type="number"]').setValue('5')
    await wrapper.findAll('input[type="number"]')[1].setValue('60')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('TRK-ABCD1234')
    )
    expect(wrapper.text()).toContain('予約が完了しました')
  })

  it('CAPACITY_EXCEEDED エラー時にメッセージを表示', async () => {
    vi.mocked(axios.post).mockRejectedValue({
      response: { data: { error: { code: 'CAPACITY_EXCEEDED' } } },
    })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('積載重量が超過しています')
    )
  })

  it('SIZE_EXCEEDED エラー時にメッセージを表示', async () => {
    vi.mocked(axios.post).mockRejectedValue({
      response: { data: { error: { code: 'SIZE_EXCEEDED' } } },
    })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('積載サイズが超過しています')
    )
  })

  it('WEIGHT_LIMIT_EXCEEDED エラー時にメッセージを表示', async () => {
    vi.mocked(axios.post).mockRejectedValue({
      response: { data: { error: { code: 'WEIGHT_LIMIT_EXCEEDED' } } },
    })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('10kg以下')
    )
  })

  it('SIZE_LIMIT_EXCEEDED エラー時にメッセージを表示', async () => {
    vi.mocked(axios.post).mockRejectedValue({
      response: { data: { error: { code: 'SIZE_LIMIT_EXCEEDED' } } },
    })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('140cm以下')
    )
  })

  it('NOT_FOUND エラー時にメッセージを表示', async () => {
    vi.mocked(axios.post).mockRejectedValue({
      response: { data: { error: { code: 'NOT_FOUND' } } },
    })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('スケジュールが見つかりません')
    )
  })

  it('重量が10kgを超えるとバリデーションエラーを表示', async () => {
    const wrapper = mountWrapper()
    const weightInput = wrapper.find('input[type="number"]')
    await weightInput.setValue('11')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('10kg以下にしてください')
  })

  it('サイズが140cmを超えるとバリデーションエラーを表示', async () => {
    const wrapper = mountWrapper()
    const inputs = wrapper.findAll('input[type="number"]')
    await inputs[1].setValue('141')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('140cm以下にしてください')
  })

  it('重量超過時は送信ボタンが無効化', async () => {
    const wrapper = mountWrapper()
    await wrapper.find('input[type="number"]').setValue('11')
    await wrapper.vm.$nextTick()
    expect(wrapper.find('button[type="submit"]').element.disabled).toBe(true)
  })

  it('サイズ超過時は送信ボタンが無効化', async () => {
    const wrapper = mountWrapper()
    await wrapper.findAll('input[type="number"]')[1].setValue('141')
    await wrapper.vm.$nextTick()
    expect(wrapper.find('button[type="submit"]').element.disabled).toBe(true)
  })
})
