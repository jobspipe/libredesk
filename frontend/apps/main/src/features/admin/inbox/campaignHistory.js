// The states a sent proactive message can be in, in the order the history filter lists them.
export const DELIVERY_STATES = ['replied', 'opened', 'dismissed', 'displayed', 'undisplayed']

export const DELIVERY_STATE_KEYS = {
  replied: 'globals.terms.replied',
  opened: 'globals.terms.opened',
  dismissed: 'globals.terms.dismissed',
  displayed: 'globals.terms.displayed',
  undisplayed: 'widget.campaignHistory.undisplayed'
}

// deliveryState is the furthest thing the visitor did with a message: replied beats opened beats dismissed beats merely displayed.
export const deliveryState = (row) => {
  if (row.replied) return 'replied'
  if (row.opened) return 'opened'
  if (row.dismissed) return 'dismissed'
  if (row.displayed) return 'displayed'
  return 'undisplayed'
}

// deliveryWho names the contact a message was shown to, or the visitor label for a browser that never identified itself.
export const deliveryWho = (row, visitorLabel) => row.contact_name || row.contact_email || visitorLabel

// deliveryPage shortens the page a message was shown on to its host and path.
export const deliveryPage = (url) => {
  try {
    const parsed = new URL(url)
    return parsed.host + parsed.pathname
  } catch {
    return url
  }
}
