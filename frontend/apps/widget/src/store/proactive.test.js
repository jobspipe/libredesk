// @vitest-environment jsdom
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

vi.mock('@widget/api/index.js', () => ({ default: { nextCampaign: vi.fn(), campaignEvent: vi.fn() } }))
vi.mock('@widget/store/widget.js', () => ({ useWidgetStore: () => ({ config: { has_campaigns: true }, isOpen: false }) }))
vi.mock('@widget/store/chat.js', () => ({ useChatStore: () => ({ getConversations: [] }) }))

const { default: api } = await import('@widget/api/index.js')
const { useProactiveStore } = await import('./proactive.js')

describe('proactive store', () => {
  let posted
  beforeEach(() => {
    setActivePinia(createPinia())
    posted = []
    window.history.replaceState({}, '', '/?parent_origin=https://example.com')
    vi.spyOn(window.parent, 'postMessage').mockImplementation(message => posted.push(message))
  })

  it('passes the retry hint to the loader when no campaign is ready yet', async () => {
    api.nextCampaign.mockResolvedValue({ data: { data: null }, headers: { 'x-campaign-retry-in': '22' } })
    const store = useProactiveStore()
    store.setBrowserKey('a'.repeat(36))
    await store.next({ url: 'https://example.com/blog' })
    expect(posted).toEqual([{ type: 'CAMPAIGN_RETRY', seconds: 22, url: 'https://example.com/blog' }])
  })

  it('posts nothing when there is no campaign and no hint', async () => {
    api.nextCampaign.mockResolvedValue({ data: { data: null }, headers: {} })
    const store = useProactiveStore()
    store.setBrowserKey('a'.repeat(36))
    await store.next({ url: 'https://example.com/' })
    expect(posted).toEqual([])
  })

  it('shows a delivery as before', async () => {
    const delivery = { id: 'd1', snapshot: { message: 'hi' } }
    api.nextCampaign.mockResolvedValue({ data: { data: delivery }, headers: {} })
    const store = useProactiveStore()
    store.setBrowserKey('a'.repeat(36))
    await store.next({ url: 'https://example.com/pricing' })
    expect(posted).toEqual([{ type: 'SHOW_CAMPAIGN', delivery }])
  })
})
