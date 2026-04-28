import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import LoginShipper from '../LoginShipper.vue'

vi.mock('axios')
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

describe('LoginShipper', () => {
  let pinia

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()
  })

  it('フォームの初期状態が正しい', () => {
    const wrapper = mount(LoginShipper, { global: { plugins: [pinia] } })
    expect(wrapper.find('input[type="email"]').element.value).toBe('')
    expect(wrapper.find('input[type="password"]').element.value).toBe('')
    expect(wrapper.find('button[type="submit"]').text()).toBe('ログイン')
  })

  it('ログイン成功時に role=shipper で API を呼ぶ', async () => {
    vi.mocked(axios.post).mockResolvedValue({
      data: { token: 'shipper-token', role: 'shipper', user_id: 'uid-2' },
    })

    const wrapper = mount(LoginShipper, { global: { plugins: [pinia] } })
    await wrapper.find('input[type="email"]').setValue('shipper@example.com')
    await wrapper.find('input[type="password"]').setValue('password')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(axios.post).toHaveBeenCalledWith(
        '/api/v1/auth/login',
        { email: 'shipper@example.com', password: 'password', role: 'shipper' }
      )
    )
  })

  it('401エラー時に認証エラーメッセージを表示', async () => {
    vi.mocked(axios.post).mockRejectedValue({ response: { status: 401 } })

    const wrapper = mount(LoginShipper, { global: { plugins: [pinia] } })
    await wrapper.find('input[type="email"]').setValue('x@y.com')
    await wrapper.find('input[type="password"]').setValue('wrong')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('メールアドレスまたはパスワードが正しくありません')
    )
  })

  it('401以外のエラー時に汎用エラーメッセージを表示', async () => {
    vi.mocked(axios.post).mockRejectedValue({ response: { status: 503 } })

    const wrapper = mount(LoginShipper, { global: { plugins: [pinia] } })
    await wrapper.find('input[type="email"]').setValue('a@b.com')
    await wrapper.find('input[type="password"]').setValue('pass')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('ログインに失敗しました')
    )
  })

  it('ログイン中はボタンが無効化', async () => {
    vi.mocked(axios.post).mockReturnValue(new Promise(() => {}))

    const wrapper = mount(LoginShipper, { global: { plugins: [pinia] } })
    await wrapper.find('input[type="email"]').setValue('a@b.com')
    await wrapper.find('input[type="password"]').setValue('pass')
    await wrapper.find('form').trigger('submit')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('button[type="submit"]').element.disabled).toBe(true)
    expect(wrapper.find('button[type="submit"]').text()).toBe('ログイン中...')
  })
})
