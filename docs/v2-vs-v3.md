# v2 vs v3 — what changed for defenders

Practitioner shorthand for Akamai Bot Manager sensor payloads uses `v1.x`, `v2`, and `v3` labels for successive payload families. Those labels are not perfectly stable — the same tenant may serve a "v3" family with tenant-specific tweaks, and rollouts happen per-customer — but the families are distinct enough at a structural level to be useful reference points.

This doc summarises what a defender needs to know about the transition, cross-referencing [Edioff/akamai-analysis](https://github.com/Edioff/akamai-analysis) for the v2 depth and this repo for the v3 shift.

## The families in one line each

- **v1.x** — Legacy sensor family. Older obfuscation, smaller script, mostly retired on public web tenants but still seen on some internal / mobile paths.
- **v2** — The generation Edioff's repo dissects in detail. 100+ collected signals, distinct string-obfuscation family, `_abck` and `bm_sz` cookie flow already established.
- **v3** — Current generation seen on new-web deployments. Larger scripts, different obfuscation, an additional per-session file-hash step in the client cipher, and heavier reliance on cross-layer consistency (see [`tls-thesis.md`](tls-thesis.md)).

## Structural deltas

| Aspect | v2 | v3 |
|---|---|---|
| Payload envelope prefix | Family-specific | Starts with `3;0;1;0;` |
| Script size (typical) | 200–400 KB | 500–560 KB |
| Cipher structure | Static-per-script transform | Element-swap seeded by a per-session **file_hash** value, plus character-substitution seeded by a cookie hash |
| Session binding | Cookie-hash from `bm_sz` | Cookie-hash from `bm_sz` **plus** file_hash from the running script |
| Cross-request rotation | Script relatively stable per tenant | Script content and size rotate per session on many tenants |
| Field count in plaintext | ~30 core fields in older public deobfuscators | ~52 fields on current variants (added timing + new constant fields) |

## What v3 added that Edioff doesn't cover

- The **file_hash** session-binding step: v3's cipher takes an additional seed that only the running script produces. This is why static extraction of cipher keys works less reliably on v3 — the extractor has to run the script's own machinery.
- The **script rotation per session**: on many current tenants, the same target serves different-sized scripts across sessions. Caching the parsed script is a defender-side foot-gun.
- Newer variant fields (`sww`, `s0xx`, `te`/`nte`/`mte`, `pus`, `og`, `tcd`) beyond the field set older deobfuscators exposed.
- The **primary detection layer shift**: on many current v3 deployments, TLS / HTTP/2 / header-order signals gate a large fraction of requests **before** the sensor even runs. See [`tls-thesis.md`](tls-thesis.md).

## What stays the same across v2 and v3

- The cookie family (`_abck`, `bm_sz`, `ak_bmsc`) and the coarse `~0~` / `~-1~` verdict semantics on `_abck`.
- The four broad server-side check categories: consistency, temporal, behavioural, correlation (per Scrapfly's summary).
- The general shape of the field categories: script constants, device profile, session timing, behavioural telemetry, server-validated tokens.
- Cross-layer mismatch as the dominant reason for rejection.

## How to tell which family a target runs

`akamaixray identify` looks at:

1. Presence of the `3;0;1;0;` envelope in observable POST bodies (only visible if you own the tenant or are running through your own instrumentation).
2. Script size on the initial GET (v3 sits in the 500–560 KB band).
3. Cookie set composition (`_abck` alone vs `_abck` + newer additions).

Result is a family label with a confidence marker, not a version string — because Akamai has never published version strings for these families and the community labels are approximate.

## References

- [Edioff/akamai-analysis](https://github.com/Edioff/akamai-analysis) — v2 deep dive; if you want v2 semantics, that's the reference.
- [glizzykingdreko's v3 sensor deep-dive](https://medium.com/@glizzykingdreko/akamai-v3-sensor-data-deep-dive-into-encryption-decryption-and-bypass-tools-da0adad2a784) — the encryption/decryption side.
- See [`references.md`](references.md).
