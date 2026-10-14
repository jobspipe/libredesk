const normalize = (code) => String(code || '').trim().replace(/_/g, '-').toLowerCase()
const baseOf = (code) => normalize(code).split('-')[0]

// Map a language tag (`fr`, `fr-CA`, `fr-FR`) to one of the available bundle codes,
// trying an exact match first and then any bundle sharing the base language.
export function matchLanguage (code, availableCodes = []) {
  const wanted = normalize(code)
  if (!wanted) return null
  const exact = availableCodes.find(c => normalize(c) === wanted)
  if (exact) return exact
  const base = baseOf(wanted)
  return availableCodes.find(c => baseOf(c) === base) || null
}

// Pick the widget UI language. A locale passed by the host page (`requested`) wins,
// then the inbox setting (a fixed code, or `auto` for the browser language), then the fallback.
export function resolveWidgetLanguage ({ requested, configured, fallback, browserLang, availableCodes = [] }) {
  const fromPage = matchLanguage(requested, availableCodes)
  if (fromPage) return fromPage
  if (configured === 'auto') return matchLanguage(browserLang, availableCodes) || fallback
  return configured || fallback
}

// Admin text fields can hold "English || Français": the widget shows the half
// matching its language. Text without the separator is shown as is.
const TEXT_SEPARATOR = '||'
const LOCALIZED_FIELDS = ['greeting_message', 'introduction_message', 'chat_introduction']

export function pickLocalizedText (text, lang) {
  if (typeof text !== 'string' || !text.includes(TEXT_SEPARATOR)) return text
  const [first, second] = text.split(TEXT_SEPARATOR).map(part => part.trim())
  const base = String(lang || '').toLowerCase().split(/[-_]/)[0]
  return base === 'fr' && second ? second : first
}

export function localizeWidgetConfig (config, lang) {
  const localized = { ...config }
  for (const field of LOCALIZED_FIELDS) {
    if (field in localized) localized[field] = pickLocalizedText(localized[field], lang)
  }
  return localized
}
