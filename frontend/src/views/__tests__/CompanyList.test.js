import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import CompanyList from '../CompanyList.vue'

vi.mock('axios')

const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
}))

const mockCompanies = [
  {
    id: 'c-001',
    name: '琉球バス交通',
    storage_image_url: '',
    storage_description: 'バスターミナル正面入口を入って左手のスペースです。',
    created_at: '2099-01-01T00:00:00Z',
  },
  {
    id: 'c-002',
    name: '那覇バス',
    storage_image_url: 'data:image/png;base64,abc',
    storage_description: 'カウンター横の棚をご利用ください。',
    created_at: '2099-01-01T00:00:00Z',
  },
  {
    id: 'c-003',
    name: '沖縄バス',
    storage_image_url: '',
    storage_description: '',
    created_at: '2099-01-01T00:00:00Z',
  },
]

describe('CompanyList', () => {
  let pinia

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()
    mockPush.mockClear()
  })

  const mountWrapper = () => mount(CompanyList, { global: { plugins: [pinia] } })

  it('読み込み中はローディング表示', async () => {
    let resolve
    vi.mocked(axios.get).mockReturnValue(new Promise((r) => { resolve = r }))
    const wrapper = mountWrapper()
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('読み込み中')
    resolve({ data: { companies: [] } })
  })

  it('会社一覧を正しく表示する', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { companies: mockCompanies } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('琉球バス交通'))
    expect(wrapper.text()).toContain('那覇バス')
    expect(wrapper.text()).toContain('沖縄バス')
  })

  it('説明文が表示される', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { companies: mockCompanies } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('バスターミナル正面入口'))
    expect(wrapper.text()).toContain('カウンター横の棚')
  })

  it('画像が登録されている場合は img タグが表示される', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { companies: mockCompanies } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('那覇バス'))
    const imgs = wrapper.findAll('img')
    expect(imgs.length).toBeGreaterThan(0)
  })

  it('情報未登録の会社は「まだ登録されていません」を表示', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { companies: mockCompanies } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('沖縄バス'))
    expect(wrapper.text()).toContain('まだ登録されていません')
  })

  it('会社が0件の場合は「バス会社情報がありません」を表示', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { companies: [] } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('バス会社情報がありません'))
  })

  it('API エラー時にエラーメッセージを表示', async () => {
    vi.mocked(axios.get).mockRejectedValue(new Error('network error'))
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('取得に失敗しました'))
  })

  it('GET /api/v1/companies を呼び出す（認証不要）', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { companies: [] } })
    mountWrapper()
    await vi.waitFor(() => expect(axios.get).toHaveBeenCalledWith('/api/v1/companies'))
  })

  it('ダッシュボードへ戻るボタンが表示される', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { companies: [] } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('ダッシュボードへ'))
    const backBtn = wrapper.find('button')
    expect(backBtn.text()).toContain('ダッシュボードへ')
  })

  it('ダッシュボードへ戻るボタンクリックで遷移', async () => {
    vi.mocked(axios.get).mockResolvedValue({ data: { companies: [] } })
    const wrapper = mountWrapper()
    await vi.waitFor(() => expect(wrapper.text()).toContain('ダッシュボードへ'))
    await wrapper.find('button').trigger('click')
    expect(mockPush).toHaveBeenCalledWith('/shipper/dashboard')
  })
})
