import { describe, expect, it } from 'vitest'
import { deliveryPage, deliveryState, deliveryWho, DELIVERY_STATES } from './campaignHistory.js'

const row = (flags = {}) => ({
  displayed: false,
  opened: false,
  dismissed: false,
  replied: false,
  contact_name: '',
  contact_email: '',
  ...flags
})

describe('deliveryState', () => {
  it('is the furthest thing the visitor did with the message', () => {
    expect(deliveryState(row())).toBe('undisplayed')
    expect(deliveryState(row({ displayed: true }))).toBe('displayed')
    expect(deliveryState(row({ displayed: true, dismissed: true }))).toBe('dismissed')
    expect(deliveryState(row({ displayed: true, opened: true, dismissed: true }))).toBe('opened')
    expect(deliveryState(row({ displayed: true, opened: true, replied: true }))).toBe('replied')
  })

  it('only returns states the list can be filtered by', () => {
    for (const flags of [{}, { displayed: true }, { dismissed: true }, { opened: true }, { replied: true }]) {
      expect(DELIVERY_STATES).toContain(deliveryState(row(flags)))
    }
  })
})

describe('deliveryWho', () => {
  it('names a known contact and falls back to their email', () => {
    expect(deliveryWho(row({ contact_name: 'Ada Lovelace', contact_email: 'ada@example.com' }), 'Visitor')).toBe('Ada Lovelace')
    expect(deliveryWho(row({ contact_email: 'ada@example.com' }), 'Visitor')).toBe('ada@example.com')
  })

  it('calls an anonymous browser a visitor', () => {
    expect(deliveryWho(row(), 'Visitor')).toBe('Visitor')
    expect(deliveryWho(row({ contact_name: 'Visitor' }), 'Besucher')).toBe('Visitor')
  })
})

describe('deliveryPage', () => {
  it('shows the site and path of the page the message appeared on', () => {
    expect(deliveryPage('https://jobspipe.dev/pricing')).toBe('jobspipe.dev/pricing')
    expect(deliveryPage('https://docs.jobspipe.dev/api-reference/jobs')).toBe('docs.jobspipe.dev/api-reference/jobs')
    expect(deliveryPage('https://jobspipe.dev/')).toBe('jobspipe.dev/')
  })

  it('reads a send recorded before pages were kept as unknown', () => {
    expect(deliveryPage('')).toBe('')
    expect(deliveryPage('not a url')).toBe('not a url')
  })
})
