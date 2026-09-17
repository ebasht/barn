# Установка, инфраструктура и релизы

Пути ниже — относительно корня репозитория. Скрипты используют Bash.

## Карта

- Локальная разработка: `scripts/local-up.sh`, `local-down.sh`, `dev-run.sh`,
  `migrate.sh`; `docker-compose.yml`, корневой `Makefile`.
- VPS: `scripts/install.sh`, общие функции `install-lib.sh`,
  `barn-up.sh`, `barn-upgrade.sh`, `barn-migrate.sh`, `barn-db-check.sh`.
  Проверяй соответствующие `dock-pilot-*` wrappers и Compose-файлы при изменениях.
- nginx панели: `scripts/configure-panel-nginx.sh`, `install/nginx-panel*.template`.
  Сервисы агента: `install/barn-agent.service`, `dockpilot-agent.service`.
- Сборка: `scripts/docker-build.sh`, `docker-export.sh`, `make-release.sh`;
  `docker-compose.build.yml`, Dockerfiles в backend/frontend.
- `.github/workflows/release.yml`: push тега `v*` собирает и публикует релиз.
  Сохраняются assets Barn и legacy DockPilot.

## Инварианты и проверка

- Новые установки используют `/opt/barn`; legacy `/opt/dock-pilot` поддерживается.
  Проверяй оба варианта имён, env fallbacks и установку с доменом/без домена.
- Обновление выполняет `barn-upgrade.sh`; повторный install не эквивалентен upgrade.
- Конфигурацию изучай через `.env*.example`; не добавляй реальные токены в вывод.
- Начни с `bash -n scripts/<изменённый-скрипт>.sh` и `git diff --check`.
  Это проверка синтаксиса, не функциональный тест установки.
- Для Compose проверяй выбранный вариант и нужные env-переменные;
  не выводи resolved config с реальными секретами в журнал.
- Не запускай reset/recover/restore/install/upgrade, публикацию или миграции
  в качестве проверки синтаксиса: они меняют данные, хост или внешние ресурсы.
