import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import ScheduleList from '../ScheduleList.vue'

vi.mock('axios')
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))
vi.mock('../../components/RouteMap.vue', () => ({
  default: { template: '<div data-testid="route-map"></div>' },
}))

const mockSchedules = [
  {
    id: 'sch-001',
    origin_name: '東京駅',
    dest_name: '大阪駅',
    origin_lat: 35.68, origin_lng: 139.76,
    dest_lat: 34.69, dest_lng: 135.50,
    depart_at: '2099-06-01T10:00:00Z',
    arrive_at: '2099-06-01T16:00:00Z',
    max_weight_kg: 100,
    max_size_cm: 140,
    avail_weight_kg: 80,
    status: 'open',
    bookings: [],
  },
  {
    id: 'sch-002',
    origin_name: '名古屋駅',
    dest_name: '京都駅',
    origin_lat: 35.17, origin_lng: 136.88,
    dest_lat: 35.01, dest_lng: 135.75,
    depart_at: '2099-06-02T09:00:00Z',
    arrive_at: '2099-06-02T11:00:00Z',
    max_weight_kg: 50,
    max_size_cm: 140,
    avail_weight_kg: 50,
    status: 'full',
    bookings: [{ id: 'b-001', tracking_number: 'TRK-001', weight_kg: 5, recipient_name: '山田', status: 'accepted' }],
  },
]

describe('ScheduleList', () => {
  let pinia

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()
  })

  const mountWrapper = () => mount(ScheduleList, { global: { plugins: [pinia] } })

  it('読み込み中はローディング表示', async () => {
    let resolveGet
    vi.mocked(axios.get).mockReturnValue(new Promise((r) => { resolveGet = r }))
    const wrapper = mountWrapper()
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('読み込み中')
    resolveGet({ data: { schedules: [] } })
  })

  it('スケジュールがない場合は「スケジュールがありません」を表示', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: [] } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('スケジュールがありません'))
  })

  it('スケジュール一覧を正しく表示する', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('東京駅'))
    expect(wrapper.text()).toContain('大阪駅')
    expect(wrapper.text()).toContain('名古屋駅')
  })

  it('ステータスラベルが正しく表示される', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('受付中'))
    expect(wrapper.text()).toContain('満載')
  })

  it('スケジュールを選択すると詳細が表示される', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('東京駅'))

    const rows = wrapper.findAll('tbody tr')
    await rows[0].trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('スケジュール詳細')
    expect(wrapper.text()).toContain('80')  // avail_weight_kg
  })

  it('スケジュール選択時に予約一覧が表示される', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('名古屋駅'))

    const rows = wrapper.findAll('tbody tr')
    await rows[1].trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('TRK-001')
    expect(wrapper.text()).toContain('山田')
  })

  it('予約がないスケジュールでは削除ボタンが有効', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('東京駅'))

    // sch-001 は bookings: [] なので削除ボタンが有効
    const deleteBtn = wrapper.findAll('button').find(b => b.text() === '削除')
    expect(deleteBtn?.element.disabled).toBe(false)
  })

  it('予約があるスケジュールでは削除ボタンが無効', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('名古屋駅'))

    const deleteBtns = wrapper.findAll('button').filter(b => b.text() === '削除')
    // sch-002 は bookings あり → disabled
    const disabledBtn = deleteBtns.find(b => b.element.disabled)
    expect(disabledBtn).toBeTruthy()
  })

  it('API エラー時にエラーメッセージを表示', async () => {
    vi.mocked(axios.get).mockRejectedValue(new Error('network error'))
    const wrapper = mountWrapper()
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('スケジュールの取得に失敗しました')
    )
  })

  it('ステータス更新後にスケジュール一覧を再取得する', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })
    vi.mocked(axios.patch).mockResolvedValue({})

    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('東京駅'))

    // sch-001 を選択
    await wrapper.findAll('tbody tr')[0].trigger('click')
    await wrapper.vm.$nextTick()

    // 「満載にする」ボタンをクリック
    const fullBtn = wrapper.findAll('button').find(b => b.text() === '満載にする')
    await fullBtn?.trigger('click')

    await vi.waitFor(() => expect(axios.patch).toHaveBeenCalledWith(
      '/api/v1/schedules/sch-001/status', { status: 'full' }
    ))
    expect(axios.get).toHaveBeenCalledTimes(2) // 初回 + 更新後
  })
})
