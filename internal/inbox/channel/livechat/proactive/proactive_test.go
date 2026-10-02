package proactive

import (
	"sync"
	"testing"
	"time"

	"github.com/abhinavxd/libredesk/internal/testutil"
	"github.com/google/uuid"
	"github.com/zerodha/logf"
)

func TestConcurrentDeliveryAndOwnership(t *testing.T) {
	db := testutil.NewDB(t, "widget_delivery")
	lo := logf.New(logf.Opts{})
	m, err := New(Opts{DB: db, Lo: &lo, I18n: testutil.NewI18n(t)})
	if err != nil {
		t.Fatal(err)
	}
	var inboxID int
	if err := db.Get(&inboxID, `INSERT INTO inboxes(name,channel) VALUES ('Test','livechat') RETURNING id`); err != nil {
		t.Fatal(err)
	}
	ctx := Context{BrowserKey: uuid.NewString(), SessionKey: uuid.NewString(), Now: time.Now()}
	c := Campaign{ID: uuid.NewString(), Repeat: "once"}
	var wg sync.WaitGroup
	var deliveries []Delivery
	for range 12 {
		wg.Go(func() {
			unlock := m.Lock()
			defer unlock()
			history, err := m.History(inboxID, ctx)
			if err != nil {
				t.Error(err)
				return
			}
			if Suppression(c, ctx, history, 24*time.Hour) != "" {
				return
			}
			delivery, err := m.Reserve(inboxID, c, ctx, Snapshot{Message: "Hello"})
			if err != nil {
				t.Error(err)
				return
			}
			deliveries = append(deliveries, delivery)
		})
	}
	wg.Wait()
	if len(deliveries) != 1 {
		t.Fatalf("got %d reservations", len(deliveries))
	}
	d := deliveries[0]
	if _, err := m.Get(d.ID, inboxID, uuid.NewString(), 0); err == nil {
		t.Fatal("another browser accessed delivery")
	}
	if _, err := m.Get(d.ID, inboxID+1, ctx.BrowserKey, 0); err == nil {
		t.Fatal("another inbox accessed delivery")
	}
	for range 2 {
		for _, event := range []string{"displayed", "opened", "dismissed"} {
			if err := m.RecordEvent(d.ID, event); err != nil {
				t.Fatal(err)
			}
		}
	}
	stats, err := m.Stats(inboxID, ctx.Now.Add(-time.Minute), ctx.Now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 || stats[0].Displayed != 1 || stats[0].Opened != 1 || stats[0].Dismissed != 1 || stats[0].Replied != 0 {
		t.Fatalf("duplicate events inflated stats: %+v", stats)
	}
	var contactID int
	if err := db.Get(&contactID, `INSERT INTO users(type,first_name) VALUES ('contact','Test') RETURNING id`); err != nil {
		t.Fatal(err)
	}
	ctx.ContactID = contactID
	if _, err := m.History(inboxID, ctx); err != nil {
		t.Fatal(err)
	}
	ctx.BrowserKey = uuid.NewString()
	history, err := m.History(inboxID, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if Suppression(c, ctx, history, 24*time.Hour) != "repeat" {
		t.Fatal("identified contact lost delivery history")
	}
}

func TestDeliveriesListsEverySendNewestFirst(t *testing.T) {
	db := testutil.NewDB(t, "widget_delivery_list")
	lo := logf.New(logf.Opts{})
	m, err := New(Opts{DB: db, Lo: &lo, I18n: testutil.NewI18n(t)})
	if err != nil {
		t.Fatal(err)
	}
	var inboxID, otherInboxID, contactID int
	if err := db.Get(&inboxID, `INSERT INTO inboxes(name,channel) VALUES ('Chat','livechat') RETURNING id`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&otherInboxID, `INSERT INTO inboxes(name,channel) VALUES ('Other','livechat') RETURNING id`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&contactID, `INSERT INTO users(type,first_name,last_name,email) VALUES ('contact','Ada','Lovelace','ada@example.com') RETURNING id`); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	welcome := Campaign{ID: uuid.NewString(), Repeat: "once"}
	pricing := Campaign{ID: uuid.NewString(), Repeat: "once"}
	reserve := func(inbox int, c Campaign, ctx Context, message string) Delivery {
		t.Helper()
		ctx.BrowserKey, ctx.SessionKey, ctx.Now = uuid.NewString(), uuid.NewString(), now
		d, err := m.Reserve(inbox, c, ctx, Snapshot{Message: message, Sender: "JobsPipe"})
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	ignored := reserve(inboxID, welcome, Context{URL: "https://jobspipe.dev/"}, "Ask anything about the API")
	replied := reserve(inboxID, pricing, Context{URL: "https://jobspipe.dev/pricing", Mobile: true, ContactID: contactID}, "Ask about our pricing plans here")
	neverShown := reserve(inboxID, welcome, Context{URL: "https://jobspipe.dev/blog/a"}, "Ask anything about the API")
	reserve(otherInboxID, welcome, Context{URL: "https://example.com/"}, "Not ours")

	for id, events := range map[string][]string{ignored.ID: {"displayed"}, replied.ID: {"displayed", "opened"}} {
		for _, event := range events {
			if err := m.RecordEvent(id, event); err != nil {
				t.Fatal(err)
			}
		}
	}
	db.MustExec(`UPDATE widget_campaign_deliveries SET replied = TRUE WHERE id = $1`, replied.ID)
	db.MustExec(`UPDATE widget_campaign_deliveries SET created_at = $2 WHERE id = $1`, ignored.ID, now.Add(-2*time.Hour))
	db.MustExec(`UPDATE widget_campaign_deliveries SET created_at = $2 WHERE id = $1`, replied.ID, now.Add(-time.Hour))
	db.MustExec(`UPDATE widget_campaign_deliveries SET created_at = $2 WHERE id = $1`, neverShown.ID, now.Add(-30*time.Minute))

	from, to := now.Add(-24*time.Hour), now.Add(time.Minute)
	rows, total, err := m.Deliveries(inboxID, from, to, DeliveryFilter{}, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(rows) != 2 || rows[0].ID != neverShown.ID || rows[1].ID != replied.ID {
		t.Fatalf("want 3 total and the two newest of this inbox, got total=%d rows=%+v", total, rows)
	}
	got := rows[1]
	if got.URL != "https://jobspipe.dev/pricing" || !got.Mobile || got.Message != "Ask about our pricing plans here" || got.CampaignID != pricing.ID {
		t.Fatalf("send lost what was shown and where: %+v", got)
	}
	if !got.Displayed || !got.Opened || got.Dismissed || !got.Replied {
		t.Fatalf("send lost what the visitor did: %+v", got)
	}
	if got.ContactID != contactID || got.ContactName != "Ada Lovelace" || got.ContactEmail != "ada@example.com" {
		t.Fatalf("send lost who it was shown to: %+v", got)
	}
	if rows[0].Displayed || rows[0].ContactID != 0 || rows[0].ContactName != "" {
		t.Fatalf("an undisplayed send to an anonymous browser reads wrong: %+v", rows[0])
	}

	page2, total, err := m.Deliveries(inboxID, from, to, DeliveryFilter{}, 2, 2)
	if err != nil || total != 3 || len(page2) != 1 || page2[0].ID != ignored.ID {
		t.Fatalf("second page: total=%d rows=%+v err=%v", total, page2, err)
	}

	for _, tc := range []struct {
		name   string
		filter DeliveryFilter
		want   []string
	}{
		{"one campaign", DeliveryFilter{CampaignID: pricing.ID}, []string{replied.ID}},
		{"replied", DeliveryFilter{State: "replied"}, []string{replied.ID}},
		{"displayed", DeliveryFilter{State: "displayed"}, []string{replied.ID, ignored.ID}},
		{"not displayed", DeliveryFilter{State: "undisplayed"}, []string{neverShown.ID}},
		{"page contains", DeliveryFilter{URL: "/blog/"}, []string{neverShown.ID}},
	} {
		rows, total, err := m.Deliveries(inboxID, from, to, tc.filter, 1, 50)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		var ids []string
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
		if total != len(tc.want) || len(ids) != len(tc.want) {
			t.Fatalf("%s: got %v total=%d, want %v", tc.name, ids, total, tc.want)
		}
		for i := range ids {
			if ids[i] != tc.want[i] {
				t.Fatalf("%s: got %v, want %v", tc.name, ids, tc.want)
			}
		}
	}

	if rows, total, err := m.Deliveries(inboxID, now.Add(-10*time.Minute), to, DeliveryFilter{}, 1, 50); err != nil || total != 0 || len(rows) != 0 {
		t.Fatalf("range filter: total=%d rows=%+v err=%v", total, rows, err)
	}
}
