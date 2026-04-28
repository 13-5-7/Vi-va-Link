import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import ScheduleSearch from '../ScheduleSearch.vue'

vi.mock('axios')

// push を外部変数として定義し、vi.mock のホイスティングに対応
const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
}))

vi.mock('../../components/RouteMap.vue', () => ({
  default: { template: '<div data-testid="route-map"></div>' },
}))

global.fetch = vi.fn()

const mockSchedules = [
  {
    id: 'sch-001',
    origin_name: '東京駅',
    dest_name: '大阪駅',
    origin_lat: 35.68, origin_lng: 139.76,
    dest_lat: 34.69, dest_lng: 135.50,
    depart_at: '2099-06-01T10:00:00Z',
    avail_weight_kg: 80,
    max_size_cm: 140,
    status: 'open',
  },
  {
    id: 'sch-002',
    origin_name: '東京駅',
    dest_name: '名古屋駅',
    origin_lat: 35.68, origin_lng: 139.76,
    dest_lat: 35.17, dest_lng: 136.88,
    depart_at: '2099-06-02T09:00:00Z',
    avail_weight_kg: 0,
    max_size_cm: 140,
    status: 'full',
  },
]

describe('ScheduleSearch', () => {
  let pinia

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()
    mockPush.mockClear()
  })

  const mountWrapper = () => mount(ScheduleSearch, { global: { plugins: [pinia] } })

  it('初期状態: 検索フォームが表示される', () => {
    const wrapper = mountWrapper()
    expect(wrapper.find('button[type="submit"]').text()).toBe('検索する')
    expect(wrapper.find('input[type="date"]').exists()).toBe(true)
  })

  it('初期状態: 検索結果エリアは非表示', () => {
    const wrapper = mountWrapper()
    expect(wrapper.find('table').exists()).toBe(false)
  })

  it('検索成功時にスケジュール一覧を表示', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => expect(wrapper.text()).toContain('東京駅'))
    expect(wrapper.text()).toContain('大阪駅')
    expect(wrapper.text()).toContain('名古屋駅')
  })

  it('検索結果が0件の場合は「該当するスケジュールがありません」を表示', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: [] } })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('該当するスケジュールがありません')
    )
  })

  it('API エラー時にエラーメッセージを表示', async () => {
    vi.mocked(axios.get).mockRejectedValue(new Error('network error'))

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('スケジュールの検索に失敗しました')
    )
  })

  it('検索中はボタンが無効化', async () => {
    vi.mocked(axios.get).mockReturnValue(new Promise(() => {}))

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('button[type="submit"]').element.disabled).toBe(true)
    expect(wrapper.find('button[type="submit"]').text()).toBe('検索中...')
  })

  it('status=open のスケジュールは予約ボタンが有効', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => expect(wrapper.text()).toContain('東京駅'))

    const bookingBtns = wrapper.findAll('button').filter(b => b.text() === '予約')
    const enabledBtn = bookingBtns.find(b => !b.element.disabled)
    expect(enabledBtn).toBeTruthy()
  })

  it('status=full のスケジュールは予約ボタンが無効', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => expect(wrapper.text()).toContain('名古屋駅'))

    const bookingBtns = wrapper.findAll('button').filter(b => b.text() === '予約')
    const disabledBtn = bookingBtns.find(b => b.element.disabled)
    expect(disabledBtn).toBeTruthy()
  })

  it('スケジュールを選択すると selectedSchedule がセットされる', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: mockSchedules } })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => expect(wrapper.text()).toContain('東京駅'))

    await wrapper.findAll('tbody tr')[0].trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.vm.selectedSchedule?.id).toBe('sch-001')
  })

  it('予約ボタンクリックで予約ページへ遷移', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: [mockSchedules[0]] } })

    const wrapper = mountWrapper()
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => expect(wrapper.text()).toContain('東京駅'))

    const bookingBtn = wrapper.findAll('button').find(b => b.text() === '予約' && !b.element.disabled)
    await bookingBtn?.trigger('click')

    expect(mockPush).toHaveBeenCalledWith('/shipper/bookings/new?schedule_id=sch-001')
  })

  it('リセットボタンで出発地・目的地がクリアされる', async () => {
    const wrapper = mountWrapper()
    wrapper.vm.originPoint = { lat: 35.68, lng: 139.76, name: '東京駅' }
    wrapper.vm.form.originName = '東京駅'
    await wrapper.vm.$nextTick()

    const resetBtn = wrapper.findAll('button').find(b => b.text() === 'リセット')
    await resetBtn?.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.vm.originPoint).toBeNull()
    expect(wrapper.vm.form.originName).toBe('')
  })

  it('出発日を指定して検索すると日付パラメータが送信される', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { schedules: [] } })

    const wrapper = mountWrapper()
    await wrapper.find('input[type="date"]').setValue('2099-06-01')
    await wrapper.find('form').trigger('submit')

    await vi.waitFor(() => expect(axios.get).toHaveBeenCalled())
    const callArgs = vi.mocked(axios.get).mock.calls[0]
    expect(callArgs[1].params).toHaveProperty('depart_at_from')
    expect(callArgs[1].params).toHaveProperty('depart_at_to')
  })
})
