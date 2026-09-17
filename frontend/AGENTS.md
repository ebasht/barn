# Frontend

Пути и команды ниже — относительно `frontend/`. Next.js 15 App Router,
React 19, TypeScript strict; alias `@/*` указывает на корень frontend.

## Точки входа

- `app/layout.tsx`, `components/AppShell.tsx`, `Nav.tsx`, `AuthGate.tsx` — оболочка.
  Страницы по областям: `app/sites`, `servers`, `databases`, `backups`,
  `notifications`, `payments`; общие элементы — `components/`.
- HTTP: `lib/api.ts`; контракт: `lib/types.ts`; нормализация: `lib/normalize.ts`.
  Переиспользуй API-клиент и обработку ошибок/авторизации.
- Адрес API: `lib/api-base.ts`; в клиентском коде предпочитай `getApiBase()`
  статическому legacy `API_BASE`. Поддерживаются auto/same-origin установки.
- Токен: `lib/auth-token.ts`; QR-вход: `app/auth/mobile`, `lib/qr-auth-code.ts`.
- Переводы: `lib/i18n/messages/en.ts` и `ru.ts`, хук/контекст в
  `lib/i18n/context.tsx`. Новые тексты добавляй для обеих локалей.
- Стили: `app/globals.css`, `app/theme-overrides.css`; переиспользуй существующие
  компоненты и классы. Учитывай мобильный экран.
- `next.config.ts`: standalone build и редиректы legacy `/fleet` → `/servers`.
  Не убирай редиректы как неиспользуемые страницы.

## Проверки

При установленных зависимостях:

```sh
npx --no-install tsc --noEmit --incremental false
# Тесты утилит, Node 22.6+ с поддержкой strip-types:
node --experimental-strip-types --test lib/env-file.test.ts lib/site-import.test.ts
# При изменениях маршрутов, сборки или общей оболочки:
npm run build
```

`*.test.ts` исключены из `tsconfig.json`: проверка типов не запускает тесты.
В package.json нет `test` script; тесты используют `node:test`.
`lint` ссылается на `next lint`, конфигурации ESLint в репозитории нет:
не считай наличие script доказательством работающего lint.
Не редактируй `.next`, `next-env.d.ts`, `tsconfig.tsbuildinfo` вручную.
