# e2e/golden — visual baselines

Checked-in golden images for the kiosk (TEST-0010) and mobile (TEST-0011)
surfaces. Diffs go under `_actual/` (gitignored).

## Kiosk (TEST-0010)

- `kiosk-root-1920.png` — DES-0008 §2 baseline at 1920x1080.
- `kiosk-root-1280-reflow.png` — DES-0008 §2 1-column reflow at 1280x800.
- `kiosk-promo-applied.png`, `kiosk-promo-invalid.png` — TC-7.

## Mobile (TEST-0011)

- `screen-3-1.png` through `screen-3-9.png` — TC-1..TC-9 wireframe goldens.

## Tooling

Kiosk: Playwright `expect(page).toHaveScreenshot()` with `threshold: 0.5%`.
Mobile: Flutter `matchesGoldenFile()` with default tolerance.

## CI pin

- Headless Chromium: 124.0.6367.* (kiosk).
- Flutter: 3.19.6 stable (mobile).

When the spec changes, re-record with `make e2e-record` (TBD) and check the
new baselines in via the docs-agent so the diff is auditable.
