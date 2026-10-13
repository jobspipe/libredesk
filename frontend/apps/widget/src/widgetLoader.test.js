// @vitest-environment jsdom
import { beforeAll, describe, expect, test } from 'vitest'
import { readFileSync } from 'node:fs'
import path from 'node:path'

const loaderSource = readFileSync(path.resolve(__dirname, '../../../../static/widget.js'), 'utf8')

let Libredesk

beforeAll(() => {
  window.eval(loaderSource)
  Libredesk = window.Libredesk
})

const iframeParams = (config) => {
  const widget = new Libredesk({ baseURL: 'https://desk.example.com', inboxID: '7', ...config })
  return new URL(widget.iframeURL()).searchParams
}

describe('widget loader iframe URL', () => {
  test('passes the page locale as lang', () => {
    expect(iframeParams({ locale: 'fr' }).get('lang')).toBe('fr')
  })

  test('omits lang when no locale is configured', () => {
    const params = iframeParams({})
    expect(params.has('lang')).toBe(false)
    expect(params.get('inbox_id')).toBe('7')
    expect(params.get('parent_origin')).toBe(window.location.origin)
  })
})
