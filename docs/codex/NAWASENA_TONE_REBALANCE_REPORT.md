# Nawasena Tone Rebalance Report

> Superseded for current state by [MULTI_THEME_REVISION_REPORT.md](MULTI_THEME_REVISION_REPORT.md). The former carousel described below was rolled back at owner request; use the current report for implementation and verification status.

## Status

**READY FOR OWNER VISUAL REVIEW**

The Nawasena tone adjustment and published-news hero carousel are implemented. Automated checks pass. Manual visual acceptance is still pending; the exact responsive matrix and authenticated admin screens remain for owner review.

## Visual Changes

- The aspiration band no longer uses a full solid gold background in Nawasena. It uses a warm near-black panel, a fine gold border, a restrained gold icon, light copy, and the existing CTA.
- The large vision card now uses the same dark brand panel in Nawasena rather than a gold block.
- Contact remains on dark canvas with an elevated information surface and gold icon accents.
- The hero's decorative illustration has been replaced with a featured-news carousel. It retains the existing left/right hero grid and responsive frame proportions.
- News slide hierarchy is cover image, category/date, short title/excerpt, and article-specific “Baca berita” link. Cover images crop with `object-fit: cover` in a fixed-ratio frame.
- Previous/next controls and small pagination are keyboard accessible; touch swipes are supported. Autoplay runs every 5.5 seconds, pauses while hovered/focused, pauses briefly after manual navigation, and does not auto-rotate under reduced-motion preference.
- When a published article has no cover, the existing media fallback icon is used. When the public endpoint fails or returns no articles, the same-size hero frame shows “Belum ada berita terbaru.” without exposing technical errors.

## Token Adjustments

- Added semantic `brand-panel` and `brand-panel-foreground` tokens.
- Nawasena maps those tokens to a warm near-black `#1C1912` with white foreground, preserving gold as a small accent and CTA color.
- Classic maps the same panel tokens to its original blue primary and white foreground, preserving the previous vision and aspiration treatment.
- Existing palette identity remains based on `#FAC90A`, `#FFFFFF`, `#5B4919`, and `#000000`. No full-bleed gold panel remains in the Nawasena vision/aspiration sections.

## Section Review

- **Landing:** public article data now requests at most five records from `/public/articles`. The existing API filters out drafts and future publications and sorts published articles by publication date. The list page and hero reuse this existing public data; there is no new endpoint or dummy carousel content.
- **Aspirasi:** dark brand panel in Nawasena; original blue panel in Classic. The route and CTA are unchanged.
- **Kontak:** dark surface and readable contact details are retained; only small icon accents use gold.
- **Admin representative screens:** shared semantic tokens apply to admin surfaces and the login page. Classic panel mapping remains blue. Automated rendering/build covers all admin routes; this visual pass only spot-checked the login page. Authenticated dashboard and representative forms still need owner review.

## Regression

- `pnpm api:integration`: **PASS** — real MariaDB-backed `TestDatabaseIntegration` ran, including 25 subcases and migrations 00001–00011.
- `pnpm verify`: **PASS** — 21 frontend test files / 101 tests, lint, format, typecheck, Next.js production build, Go formatting, vet, tests, and build.
- `pnpm audit`: **PASS** — no known vulnerabilities found.
- `git diff --check`: **PASS** after report save.
- `source_docs/` and `design_reference/`: unchanged in the previous dual-theme check; rechecked before handoff.

Carousel tests cover actual article-specific links, category/date, cover cropping, stable empty state, manual next/previous/dot selection, and 5.5-second rotation with an interaction pause. Reduced-motion behavior is implemented and remains covered by the CSS preference rules and the carousel's motion-aware timer.

## Responsive Review

The production build preserves responsive aspect ratios and touch handling. Exact visual inspection at all requested widths remains pending owner review; do not treat this as a responsive acceptance sign-off.

| Viewport | Nawasena | Classic | Status |
|----------|----------|---------|--------|
| 1440×900 | Pending owner review | Pending owner review | Not captured at exact size |
| 1280×800 | Pending owner review | Pending owner review | Existing browser spot check was 1280×720 |
| 360×800 | Pending owner review | Pending owner review | Not captured at exact size |
| 390×844 | Pending owner review | Pending owner review | Not captured at exact size |
| 412×915 | Pending owner review | Pending owner review | Not captured at exact size |

## Release State

- Existing `v1.0.0-campus-handoff` tag remains unchanged.
- No production deployment was performed.
- No new tag or GitHub Release was created.
- Existing handoff ZIPs and deployment handoff artifacts were not regenerated or overwritten.
- No commit was created.

## Owner Review

Preview: **http://127.0.0.1:3290/**

Inspect both Nawasena and Classic, especially the landing hero carousel, aspiration band, contact section, structure, registration, and authenticated admin representative screens. The local API remains on `http://127.0.0.1:8080/healthz`.
