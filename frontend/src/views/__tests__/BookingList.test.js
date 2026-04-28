import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import BookingList from '../BookingList.vue'

vi.mock('axios')
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))
vi.mock('../../components/BookingStatusBadge.vue', () => ({
  default: { props: ['status'], template: '<span>{{ status }}</span>' },
}))

const mockBookings = [
  {
    id: 'b-001',
    schedule_id: 'sch-001',
    tracking_number: 'TRK-AAA001',
    status: 'accepted',
    created_at: '2099-01-01T00:00:00Z',
    weight_kg: 3,
  },
  {
    id: 'b-002',
    schedule_id: 'sch-001',
    tracking_number: 'TRK-BBB002',
    status: 'delivered',
    created_at: '2099-01-02T00:00:00Z',
    weight_kg: 5,
  },
]

const mockSchedule = {
  id: 'sch-001',
  origin_name: '東京駅',
  dest_name: '大阪駅',
}

describe('BookingList', () => {
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

  const mountWrapper = () => mount(BookingList, { global: { plugins: [pinia] } })

  it('読み込み中はローディング表示', async () => {
    // axios.get が解決する前の loading=true 状態を確認
    let resolveGet
    vi.mocked(axios.get).mockReturnValue(new Promise((r) => { resolveGet = r }))
    const wrapper = mountWrapper()
    // onMounted で fetchBookings が呼ばれ loading=true になる
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('読み込み中')
    resolveGet({ data: { bookings: [] } })
  })

  it('予約がない場合は「予約がありません」を表示', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { bookings: [] } })

    const wrapper = mountWrapper()
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('予約がありません')
    )
  })

  it('予約一覧を正しく表示する', async () => {
    vi.mocked(axios.get).mockImplementation((url) => {
      if (url === '/api/v1/bookings') return Promise.resolve({ data: { bookings: mockBookings } })
      if (url.includes('/api/v1/schedules/')) return Promise.resolve({ data: mockSchedule })
      return Promise.reject(new Error('unexpected'))
    })

    const wrapper = mountWrapper()
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('TRK-AAA001')
    )
    expect(wrapper.text()).toContain('TRK-BBB002')
    expect(wrapper.text()).toContain('東京駅')
    expect(wrapper.text()).toContain('大阪駅')
  })

  it('スケジュール取得失敗時はスケジュールIDをフォールバック表示', async () => {
    vi.mocked(axios.get).mockImplementation((url) => {
      if (url === '/api/v1/bookings') return Promise.resolve({ data: { bookings: mockBookings } })
      return Promise.reject(new Error('schedule fetch failed'))
    })

    const wrapper = mountWrapper()
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('TRK-AAA001')
    )
    // スケジュール名の代わりにIDが表示される
    expect(wrapper.text()).toContain('sch-001')
  })

  it('API エラー時にエラーメッセージを表示', async () => {
    vi.mocked(axios.get).mockRejectedValue({
      response: { status: 401, data: { error: { message: 'Unauthorized' } } },
      message: 'Request failed',
    })

    const wrapper = mountWrapper()
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('予約一覧の取得に失敗しました')
    )
  })

  it('60秒ごとに自動更新する', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { bookings: [] } })

    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(axios.get).toHaveBeenCalledTimes(1))

    vi.advanceTimersByTime(60_000)
    await vi.waitFor(() => expect(axios.get).toHaveBeenCalledTimes(2))

    wrapper.unmount()
  })

  it('アンマウント時にポーリングが停止する', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { bookings: [] } })

    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(axios.get).toHaveBeenCalledTimes(1))

    wrapper.unmount()
    vi.advanceTimersByTime(60_000)
    // アンマウント後は追加呼び出しなし
    expect(axios.get).toHaveBeenCalledTimes(1)
  })
})
