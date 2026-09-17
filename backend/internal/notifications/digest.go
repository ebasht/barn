package notifications

import (
	"fmt"
	"strings"
	"time"
)

func formatDigest(mode, panelName string, rows []digestItem, names map[string]string, servers []ServerSummaryItem, now time.Time, tz string) string {
	if mode != "compact" {
		return formatDailyDigest(panelName, rows, names, servers, now, tz)
	}
	var issues strings.Builder
	for _, item := range rows {
		if item.Overall == "healthy" {
			continue
		}
		name := names[item.Key]
		if name == "" {
			name = item.Key
		}
		kind := "Сайт"
		if item.Kind == "postgres" {
			kind = "База данных"
		}
		status := "статус неизвестен"
		switch item.Overall {
		case "degraded":
			status = "есть проблемы"
		case "unhealthy":
			status = "недоступен"
		}
		fmt.Fprintf(&issues, "%s %s: %s — %s", statusEmoji(item.Overall), kind, escapeHTML(name), status)
		if message := strings.TrimSpace(item.Message); message != "" {
			fmt.Fprintf(&issues, ": %s", escapeHTML(message))
		}
		issues.WriteByte('\n')
	}
	for _, server := range servers {
		if server.Status != "online" {
			status := "статус неизвестен"
			switch server.Status {
			case "warning":
				status = "нестабильно"
			case "offline":
				status = "не в сети"
			}
			fmt.Fprintf(&issues, "%s Сервер: %s — %s\n", serverStatusEmoji(server.Status), escapeHTML(server.Name), status)
		}
		if server.DaysLeft != nil && *server.DaysLeft <= 0 {
			if *server.DaysLeft == 0 {
				fmt.Fprintf(&issues, "💳 Сервер: %s — оплата истекает сегодня\n", escapeHTML(server.Name))
			} else {
				fmt.Fprintf(&issues, "💳 Сервер: %s — оплата просрочена на %d дн.\n", escapeHTML(server.Name), -*server.DaysLeft)
			}
		}
	}

	header := fmt.Sprintf("<b>Barn — %s</b>\n", escapeHTML(panelName))
	if issues.Len() > 0 {
		return header + strings.TrimSpace(issues.String())
	}
	if len(rows) == 0 && len(servers) == 0 {
		return header + "Нет объектов для проверки."
	}
	return header + "✅ Всё ОК"
}
