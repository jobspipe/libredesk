import { describe, expect, test } from 'vitest'
import { matchLanguage, resolveWidgetLanguage, pickLocalizedText, localizeWidgetConfig } from './language.js'

const available = ['da-DK', 'de-DE', 'en-US', 'fr-FR', 'pt-BR', 'pt-PT']

describe('matchLanguage', () => {
  test('returns an exact code as is', () => {
    expect(matchLanguage('fr-FR', available)).toBe('fr-FR')
  })

  test('matches case-insensitively and with underscores', () => {
    expect(matchLanguage('fr_fr', available)).toBe('fr-FR')
  })

  test('maps a bare language to the first bundle of that language', () => {
    expect(matchLanguage('fr', available)).toBe('fr-FR')
    expect(matchLanguage('en', available)).toBe('en-US')
  })

  test('falls back on the base language for an unknown region', () => {
    expect(matchLanguage('fr-CA', available)).toBe('fr-FR')
    expect(matchLanguage('en-GB', available)).toBe('en-US')
  })

  test('prefers the exact region over the first bundle of the language', () => {
    expect(matchLanguage('pt-PT', available)).toBe('pt-PT')
  })

  test('returns null when nothing matches or input is empty', () => {
    expect(matchLanguage('ja-JP', available)).toBeNull()
    expect(matchLanguage('', available)).toBeNull()
    expect(matchLanguage(undefined, available)).toBeNull()
  })
})

describe('resolveWidgetLanguage', () => {
  test('uses the fixed inbox language when no page locale is passed', () => {
    expect(resolveWidgetLanguage({ configured: 'de-DE', fallback: 'en-US', availableCodes: available })).toBe('de-DE')
  })

  test('uses the fallback when the inbox has no language and no page locale is passed', () => {
    expect(resolveWidgetLanguage({ configured: '', fallback: 'en-US', availableCodes: available })).toBe('en-US')
  })

  test('auto picks the browser language, now with base-language fallback', () => {
    expect(resolveWidgetLanguage({ configured: 'auto', fallback: 'en-US', browserLang: 'fr-CA', availableCodes: available })).toBe('fr-FR')
  })

  test('auto uses the fallback when the browser language is unavailable', () => {
    expect(resolveWidgetLanguage({ configured: 'auto', fallback: 'en-US', browserLang: 'ja-JP', availableCodes: available })).toBe('en-US')
  })

  test('the page locale overrides a fixed inbox language', () => {
    expect(resolveWidgetLanguage({ requested: 'fr', configured: 'en-US', fallback: 'en-US', availableCodes: available })).toBe('fr-FR')
  })

  test('the page locale overrides the browser language in auto mode', () => {
    expect(resolveWidgetLanguage({ requested: 'en', configured: 'auto', fallback: 'en-US', browserLang: 'fr-CA', availableCodes: available })).toBe('en-US')
  })

  test('an unavailable page locale is ignored', () => {
    expect(resolveWidgetLanguage({ requested: 'ja', configured: 'de-DE', fallback: 'en-US', availableCodes: available })).toBe('de-DE')
  })
})

describe('pickLocalizedText', () => {
  test('returns plain text unchanged', () => {
    expect(pickLocalizedText('How can we help?', 'fr-FR')).toBe('How can we help?')
  })
  test('picks the French half for a French widget', () => {
    expect(pickLocalizedText('How can we help? || Comment pouvons-nous vous aider ?', 'fr-FR')).toBe('Comment pouvons-nous vous aider ?')
    expect(pickLocalizedText('Hi || Salut', 'fr')).toBe('Salut')
  })
  test('picks the first half for any other language', () => {
    expect(pickLocalizedText('How can we help? || Comment pouvons-nous vous aider ?', 'en-US')).toBe('How can we help?')
  })
  test('handles empty values', () => {
    expect(pickLocalizedText('', 'fr-FR')).toBe('')
    expect(pickLocalizedText(undefined, 'fr-FR')).toBe(undefined)
  })
})

describe('localizeWidgetConfig', () => {
  test('localizes the visitor-facing text fields only', () => {
    const config = { greeting_message: 'Hello || Bonjour', introduction_message: 'Help? || Aide ?', chat_introduction: 'Ask || Demandez', language: 'auto' }
    expect(localizeWidgetConfig(config, 'fr-FR')).toEqual({ greeting_message: 'Bonjour', introduction_message: 'Aide ?', chat_introduction: 'Demandez', language: 'auto' })
  })
})
