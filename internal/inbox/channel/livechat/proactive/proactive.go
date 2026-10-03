// Package proactive picks which livechat campaign message to show a widget visitor and records what they do with it.
package proactive

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/abhinavxd/libredesk/internal/dbutil"
	"github.com/abhinavxd/libredesk/internal/envelope"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/go-i18n"
	"github.com/zerodha/logf"
)

var (
	//go:embed queries.sql
	efs embed.FS
)

const maxDeliveryURLLength = 2048

type Manager struct {
	q    queries
	lo   *logf.Logger
	i18n *i18n.I18n
	mu   sync.Mutex
}

type Opts struct {
	DB   *sqlx.DB
	Lo   *logf.Logger
	I18n *i18n.I18n
}

type queries struct {
	History     *sqlx.Stmt `query:"history"`
	Reserve     *sqlx.Stmt `query:"reserve"`
	GetDelivery *sqlx.Stmt `query:"get-delivery"`
	RecordEvent *sqlx.Stmt `query:"record-event"`
	BindContact *sqlx.Stmt `query:"bind-contact"`
	Stats       *sqlx.Stmt `query:"stats"`
	Deliveries  *sqlx.Stmt `query:"deliveries"`
}

func New(opts Opts) (*Manager, error) {
	var q queries
	if err := dbutil.ScanSQLFile("queries.sql", &q, opts.DB, efs); err != nil {
		return nil, err
	}
	return &Manager{q: q, lo: opts.Lo, i18n: opts.I18n}, nil
}

func (m *Manager) Lock() func() {
	m.mu.Lock()
	return m.mu.Unlock
}

func (m *Manager) History(inboxID int, ctx Context) ([]Delivery, error) {
	if ctx.ContactID > 0 {
		if _, err := m.q.BindContact.Exec(inboxID, ctx.BrowserKey, ctx.ContactID); err != nil {
			return nil, m.error(err)
		}
	}
	out := make([]Delivery, 0)
	if err := m.q.History.Select(&out, inboxID, ctx.BrowserKey, ctx.ContactID); err != nil {
		return nil, m.error(err)
	}
	return out, nil
}

func (m *Manager) Reserve(inboxID int, campaign Campaign, ctx Context, snapshot Snapshot) (Delivery, error) {
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return Delivery{}, m.error(err)
	}
	delivery := Delivery{CampaignID: campaign.ID, Snapshot: raw}
	if err := m.q.Reserve.Get(&delivery.ID, campaign.ID, inboxID, ctx.BrowserKey, ctx.SessionKey, ctx.ContactID, raw, deliveryURL(ctx.URL), ctx.Mobile); err != nil {
		return Delivery{}, m.error(err)
	}
	return delivery, nil
}

func (m *Manager) Get(id string, inboxID int, browserKey string, contactID int) (Delivery, error) {
	var out Delivery
	if err := m.q.GetDelivery.Get(&out, id, inboxID, browserKey, contactID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, envelope.NewError(envelope.NotFoundError, m.i18n.T("globals.messages.notFound"), nil)
		}
		return out, m.error(err)
	}
	return out, nil
}

func (m *Manager) RecordEvent(id, event string) error {
	if _, err := m.q.RecordEvent.Exec(id, event); err != nil {
		return m.error(err)
	}
	return nil
}

func (m *Manager) Stats(inboxID int, from, to time.Time) ([]Stats, error) {
	out := make([]Stats, 0)
	if err := m.q.Stats.Select(&out, inboxID, from, to); err != nil {
		return nil, m.error(err)
	}
	return out, nil
}

// Deliveries returns one page of the proactive messages sent in an inbox, newest first, with the total that match.
func (m *Manager) Deliveries(inboxID int, from, to time.Time, filter DeliveryFilter, page, pageSize int) ([]DeliveryRecord, int, error) {
	out := make([]DeliveryRecord, 0)
	if err := m.q.Deliveries.Select(&out, inboxID, from, to, filter.CampaignID, filter.State, filter.URL, pageSize, (page-1)*pageSize); err != nil {
		return nil, 0, m.error(err)
	}
	if len(out) == 0 {
		return out, 0, nil
	}
	return out, out[0].Total, nil
}

// deliveryURL keeps the page a message was shown on without its query string or fragment, which can carry tokens.
func deliveryURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	u.RawQuery, u.Fragment, u.User = "", "", nil
	out := u.String()
	if len(out) > maxDeliveryURLLength {
		return out[:maxDeliveryURLLength]
	}
	return out
}

func (m *Manager) error(err error) error {
	m.lo.Error("proactive message operation failed", "error", err)
	return envelope.NewError(envelope.GeneralError, m.i18n.T("globals.messages.somethingWentWrong"), nil)
}

// RepeatVisitor suppresses a campaign for a browser (or identified contact) that has already
// received any proactive message from this inbox, so a visitor is messaged once in total.
const RepeatVisitor = "visitor"

// Repeats are the repeat rules a campaign may carry; the admin form offers the same list.
var Repeats = []string{RepeatVisitor, "once", "session", "interval"}

func Suppression(c Campaign, ctx Context, history []Delivery, cooldown time.Duration) string {
	for _, d := range history {
		if !d.Displayed && !d.Dismissed && !d.Replied && ctx.Now.Sub(d.CreatedAt) > time.Minute {
			continue
		}
		// "visitor": one proactive message per person, whichever campaign sent it.
		if c.Repeat == RepeatVisitor {
			return "repeat"
		}
		if d.CampaignID == c.ID {
			if d.Replied || c.Repeat == "once" || c.Repeat == "session" && d.SessionKey == ctx.SessionKey || c.Repeat == "interval" && ctx.Now.Sub(d.CreatedAt) < time.Duration(c.RepeatHours)*time.Hour {
				return "repeat"
			}
		}
		if cooldown > 0 && ctx.Now.Sub(d.CreatedAt) < cooldown {
			return "cooldown"
		}
	}
	return ""
}
