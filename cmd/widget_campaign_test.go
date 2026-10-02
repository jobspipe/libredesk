package main

import (
	"testing"
	"time"

	bhmodels "github.com/abhinavxd/libredesk/internal/business_hours/models"
)

func TestWithinWorkingHours(t *testing.T) {
	tests := []struct {
		name  string
		clock string
		day   bhmodels.WorkingHours
		want  bool
	}{
		{name: "normal open", clock: "09:00", day: bhmodels.WorkingHours{Open: "09:00", Close: "17:00"}, want: true},
		{name: "normal close", clock: "17:00", day: bhmodels.WorkingHours{Open: "09:00", Close: "17:00"}},
		{name: "overnight evening", clock: "23:00", day: bhmodels.WorkingHours{Open: "22:00", Close: "02:00"}, want: true},
		{name: "overnight morning", clock: "01:00", day: bhmodels.WorkingHours{Open: "22:00", Close: "02:00"}, want: true},
		{name: "overnight closed", clock: "12:00", day: bhmodels.WorkingHours{Open: "22:00", Close: "02:00"}},
		{name: "same time closed", clock: "09:00", day: bhmodels.WorkingHours{Open: "09:00", Close: "09:00"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := withinWorkingHours(tc.clock, tc.day); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseCampaignCooldown(t *testing.T) {
	tests := []struct {
		value string
		want  time.Duration
		valid bool
	}{
		{value: "0s", valid: true},
		{value: "10m", want: 10 * time.Minute, valid: true},
		{value: "1h30m", want: 90 * time.Minute, valid: true},
		{value: ""},
		{value: "tomorrow"},
		{value: "-1m"},
	}

	for _, tc := range tests {
		t.Run(tc.value, func(t *testing.T) {
			got, err := parseCampaignCooldown(tc.value)
			if tc.valid && (err != nil || got != tc.want) {
				t.Fatalf("got %v, %v; want %v", got, err, tc.want)
			}
			if !tc.valid && err == nil {
				t.Fatalf("accepted %q", tc.value)
			}
		})
	}
}

func TestParseDeliveryFilter(t *testing.T) {
	campaign := "3f0c2d0e-8a55-4c1e-9d55-1b2c3d4e5f60"
	tests := []struct {
		name                 string
		campaign, state, url string
		valid                bool
	}{
		{name: "no filter", valid: true},
		{name: "campaign, state and page", campaign: campaign, state: "replied", url: "/pricing", valid: true},
		{name: "undisplayed", state: "undisplayed", valid: true},
		{name: "unknown state", state: "purchased"},
		{name: "campaign that is not an id", campaign: "1 OR 1=1"},
		{name: "overlong page", url: string(make([]byte, 513))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseDeliveryFilter(tc.campaign, tc.state, tc.url)
			if tc.valid != (err == nil) {
				t.Fatalf("valid=%v, err=%v", tc.valid, err)
			}
			if tc.valid && (got.CampaignID != tc.campaign || got.State != tc.state || got.URL != tc.url) {
				t.Fatalf("filter changed on the way in: %+v", got)
			}
		})
	}
}
