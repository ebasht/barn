# Контекст проекта

Barn (ранее DockPilot) — панель управления сайтами и ботами на VPS: Go API,
Next.js UI, PostgreSQL, Docker, nginx, certbot. Имя каталога `dock-pilot` историческое.

## Как читать проект экономно

- Сначала `git status --short`, затем только файлы нужной подсистемы и её `AGENTS.md`.
- Используй карту ниже и точечный `rg -n` по каталогу; не загружай весь репозиторий.
- Не читай lock-файлы, сгенерированный код, бинарники, `node_modules`, `.next`,
  `dist`, `*.tsbuildinfo`, если задача непосредственно их не касается.
- Секреты не нужны для знакомства с проектом: настройки смотри в `.env*.example`
  и `backend/internal/config/config.go`, не выводи содержимое рабочих `.env`.
- Читай подробные документы только по теме задачи. Не дублируй их в контекстных файлах.

## Где искать

| Задача | Точки входа (пути от корня) |
| --- | --- |
| API, маршруты, авторизация | `backend/internal/api/router.go`, `auth.go`, `*_handler.go` |
| Запуск API, wiring, фоновые процессы | `backend/cmd/server/main.go` |
| Сайты, контейнеры, деплой | `backend/internal/sites`, `deployments`, `docker`, `nginx`, `ssl` |
| Базы и бэкапы | `backend/internal/pgdb`, `panelbackup`, `s3util`; `frontend/components/backups` |
| Серверы, pairing, мониторинг | `backend/internal/servers`, `agent`, `metrics`; [docs/servers.md](docs/servers.md), [docs/barn-agent.md](docs/barn-agent.md) |
| Telegram, оплата | `backend/internal/notifications`, `billing` |
| MCP | `backend/internal/barnmcp`, `backend/internal/api/mcp_handler.go`; [docs/barn-mcp.md](docs/barn-mcp.md) |
| UI, API-клиент, типы | `frontend/app`, `frontend/components`, `frontend/lib/api.ts`, `frontend/lib/types.ts` |
| SQL | `backend/migrations`, `backend/schema/schema.sql`, `backend/queries`, `backend/sqlc.yaml` |
| Установка и релиз | `scripts`, `install`, `docker-compose*.yml`, `.github/workflows/release.yml`; [README.md](README.md) |

## Работа и проверки

- Инструкции по областям: [backend/AGENTS.md](backend/AGENTS.md),
  [frontend/AGENTS.md](frontend/AGENTS.md), [scripts/AGENTS.md](scripts/AGENTS.md).
  Для изменений `install/`, Compose и release workflow также прочитай инструкции scripts.
- Локальный запуск: `make setup` → `make up` → `make dev-run` (API :8080, UI :3000).
  Это установка зависимостей, запуск Docker и миграций; для правки документации не требуется.
- Для локального деплоя предусмотрен `DEPLOY_MODE=stub`; `real` меняет инфраструктуру.
- Сохраняй совместимость Barn/DockPilot: legacy wrappers, имена образов, пути,
  переменные и aliases не удаляй как «мёртвый код» без проверки пользователей.
- `make reset` удаляет данные PostgreSQL; `make pushandrelease` коммитит всё,
  создаёт тег и делает push. Это не команды проверки.
- Запускай проверки затронутой области; сообщай, что проверено и что не удалось.
  Для изменений только Markdown достаточно проверки ссылок и `git diff --check`.
- При изменении точек входа или команд обновляй соответствующий `AGENTS.md`.
  Храни здесь устойчивые факты и ссылки, а не историю сессий и копии исходников.
