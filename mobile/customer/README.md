# Monti Customer Mobile (Flutter)

Implements TASK-0012 — the 9-screen customer ordering flow per DES-0008
§3.1-§3.9, with QR scan, voice WS client, and 4-language localization.

## Bootstrap (run once)

This worktree contains only the Dart source (`lib/`, `test/`) and asset
manifests. Native platform folders are regenerated locally so this repo
stays platform-agnostic and CI-friendly:

```bash
flutter create --project-name monti_customer --platforms=android,ios,web .
flutter pub get
flutter gen-l10n
```

`flutter create` will fill in `android/`, `ios/`, and `web/` without
touching the curated `lib/` and `test/` trees.

## Routes

| Route                       | Screen                       | DES-0008 §  |
|-----------------------------|------------------------------|-------------|
| `/`                         | `WelcomeScreen`              | §3.1        |
| `/qr`                       | `QrScanScreen`               | (entry)     |
| `/language`                 | `LanguageSelectionScreen`    | §3.2        |
| `/categories`               | `ChooseCategoryScreen`       | §3.3        |
| `/menu/:catId`              | `MenuListScreen`             | §3.4        |
| `/item/:id`                 | `ItemDetailScreen`           | §3.5        |
| `/cart`                     | `CartScreen`                 | §3.6        |
| `/checkout`                 | `CheckoutPaymentScreen`      | §3.7        |
| `/thank-you/:orderNumber`   | `ThankYouScreen`             | §3.8        |
| `/order/:orderNumber`       | `OrderStatusScreen`          | §3.9        |

## Tests

```bash
flutter analyze
flutter test                       # widget + unit tests
flutter test --update-goldens      # regenerate baselines under test/golden/
flutter build apk --debug          # smoke build
```

## Backend dependencies

- `GET /api/v1/menu/categories`, `GET /api/v1/menu/items` (TASK-0009).
- `POST /api/v1/public/qr/redeem` (TASK-0009).
- `POST /api/v1/public/orders`, `GET /api/v1/public/orders/{order_number}`,
  WS `/api/v1/public/orders/{order_number}/stream` (TASK-0009 / DES-0007 §7).
- WS `/api/v1/voice/sessions` (TASK-0010 / DES-0009 §2).

Set `--dart-define=MONTI_API_BASE=https://monti.dev.example.com` to point
at a non-local backend.
