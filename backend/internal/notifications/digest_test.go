package notifications

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ebash/barn/backend/internal/db"
)

func TestCompactDigest(t *testing.T) {
	today, overdue, paid := 0, -2, 10
	cases := []struct {
		name    string
		rows    []digestItem
		servers []ServerSummaryItem
		want    []string
		absent  []string
	}{
		{
			name:    "healthy services and servers",
			rows:    []digestItem{{Key: "shop", Kind: "site", Overall: "healthy"}, {Key: "db", Kind: "postgres", Overall: "healthy"}},
			servers: []ServerSummaryItem{{Name: "paid", Status: "online", DaysLeft: &paid}, {Name: "no billing", Status: "online"}},
			want:    []string{"✅ Всё ОК"}, absent: []string{"shop", "paid", "no billing", "Сводка"},
		},
		{
			name: "problems include reasons and escape HTML",
			rows: []digestItem{{Key: "shop", Kind: "site", Overall: "unhealthy", Message: "HTTP <500>"}, {Key: "db", Kind: "postgres", Overall: "degraded", Message: "connection refused"}, {Key: "ok", Overall: "healthy"}},
			want: []string{"Сайт: Shop &amp; API", "HTTP &lt;500&gt;", "База данных: db", "connection refused"}, absent: []string{"Всё ОК", "<500>", "✅"},
		},
		{
			name:    "unknown and missing statuses are not healthy",
			rows:    []digestItem{{Key: "unknown", Overall: "unknown"}, {Key: "missing"}},
			servers: []ServerSummaryItem{{Name: "new", Status: "pending"}},
			want:    []string{"unknown — статус неизвестен", "missing — статус неизвестен", "new — статус неизвестен"}, absent: []string{"Всё ОК"},
		},
		{
			name:    "server availability and billing",
			servers: []ServerSummaryItem{{Name: "offline", Status: "offline", DaysLeft: &overdue}, {Name: "warning", Status: "warning"}, {Name: "online", Status: "online", DaysLeft: &today}},
			want:    []string{"offline — не в сети", "warning — нестабильно", "оплата просрочена на 2 дн.", "online — оплата истекает сегодня"}, absent: []string{"Всё ОК"},
		},
		{name: "empty inventory", want: []string{"Нет объектов для проверки."}, absent: []string{"Всё ОК"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatDigest("compact", "panel & one", tc.rows, map[string]string{"shop": "Shop & API"}, tc.servers, time.Now(), "UTC")
			if !strings.HasPrefix(got, "<b>Barn — panel &amp; one</b>\n") {
				t.Fatalf("missing escaped panel name: %s", got)
			}
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("missing %q in %s", want, got)
				}
			}
			for _, absent := range tc.absent {
				if strings.Contains(got, absent) {
					t.Errorf("unexpected %q in %s", absent, got)
				}
			}
		})
	}
}

func TestDetailedDigestRemainsDefault(t *testing.T) {
	now := time.Now()
	rows := []digestItem{{Key: "shop", Kind: "site", Overall: "healthy"}}
	want := formatDailyDigest("panel", rows, nil, nil, now, "UTC")
	for _, mode := range []string{"", "detailed"} {
		if got := formatDigest(mode, "panel", rows, nil, nil, now, "UTC"); got != want {
			t.Fatalf("mode %q changed detailed digest", mode)
		}
	}
}

func TestValidateDigestMode(t *testing.T) {
	for _, mode := range []string{"", "detailed", "compact", "invalid"} {
		err := validateUpdate(UpdateSettingsRequest{DailyDigestMode: mode}, db.NotificationSetting{})
		if mode == "invalid" {
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid input, got %v", err)
			}
		} else if err != nil {
			t.Fatalf("mode %q: %v", mode, err)
		}
	}
}
