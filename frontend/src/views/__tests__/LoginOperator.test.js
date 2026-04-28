import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import LoginOperator from '../LoginOperator.vue'

vi.mock('axios')
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

describe('LoginOperator', () => {
  let pinia

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()
  })

  it('フォームの初期状態が正しい', () => {
    const wrapper = mount(LoginOperator, { global: { plugins: [pinia] } })
    expect(wrapper.find('input[type="email"]').element.value).toBe('')
    expect(wrapper.find('input[type="password"]').element.value).toBe('')
    expect(wrapper.find('button[type="submit"]').text()).toBe('ログイン')
    expect(wrapper.find('button[type="submit"]').element.disabled).toBe(false)
  })

  it('エラーメッセージが初期状態では非表示', () => {
    const wrapper = mount(LoginOperator, { global: { plugins: [pinia] } })
    expect(wrapper.find('[class*="text-red"]').exists()).toBe(false)
  })

  it('ログイン成功時にダッシュボードへリダイレクト', async () => {
    vi.mocked(axios.post).mockResolvedValue({
      data: { token: 'test-token', role: 'bus_operator', user_id: 'uid-1' },
    })

    const wrapper = mount(LoginOperator, { global: { plugins: [pinia] } })
    await wrapper.find('input[type="email"]').setValue('operator@example.com')
    await wrapper.find('input[type="password"]').setValue('password')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() => expect(axios.post).toHaveBeenCalledWith(
      '/api/v1/auth/login',
      { email: 'operator@example.com', password: 'password', role: 'bus_operator' }
    ))
  })

  it('401エラー時に認証エラーメッセージを表示', async () => {
    vi.mocked(axios.post).mockRejectedValue({ response: { status: 401 } })

    const wrapper = mount(LoginOperator, { global: { plugins: [pinia] } })
    await wrapper.find('input[type="email"]').setValue('wrong@example.com')
    await wrapper.find('input[type="password"]').setValue('wrong')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('メールアドレスまたはパスワードが正しくありません')
    )
  })

  it('401以外のエラー時に汎用エラーメッセージを表示', async () => {
    vi.mocked(axios.post).mockRejectedValue({ response: { status: 500 } })

    const wrapper = mount(LoginOperator, { global: { plugins: [pinia] } })
    await wrapper.find('input[type="email"]').setValue('a@b.com')
    await wrapper.find('input[type="password"]').setValue('pass')
    await wrapper.find('form').trigger('submit')
    await vi.waitFor(() =>
      expect(wrapper.text()).toContain('ログインに失敗しました')
    )
  })

  it('ログイン中はボタンが無効化されローディングテキストを表示', async () => {
    // 解決しない Promise でローディング状態を維持
    vi.mocked(axios.post).mockReturnValue(new Promise(() => {}))

    const wrapper = mount(LoginOperator, { global: { plugins: [pinia] } })
    await wrapper.find('input[type="email"]').setValue('a@b.com')
    await wrapper.find('input[type="password"]').setValue('pass')
    await wrapper.find('form').trigger('submit')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('button[type="submit"]').element.disabled).toBe(true)
    expect(wrapper.find('button[type="submit"]').text()).toBe('ログイン中...')
  })
})
