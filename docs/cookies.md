# Akamai Bot Manager cookie taxonomy

Every cookie Akamai Bot Manager and related edge products can set, what each one carries semantically, and what a defender or authorised tester should read from it. **No forgery, no rotation logic — this is a read-side reference.**

## Quick table

| Cookie | Set by | Semantics | Format hint |
|--------|--------|-----------|-------------|
| `_abck` | Bot Manager | Mitigation state / challenge verdict token. The presence or absence of the `~0~` substring is the coarse "sensor accepted" flag observed in the wild. | Delimited fields around `~n~` markers; deployment-specific, treat as opaque |
| `bm_sz` | Bot Manager | Challenge-parameter bundle handed to the client script. Its structure carries per-session parameters the sensor script reads back. | `~`-delimited fields; one field is commonly consumed as a per-session cookie hash by the client script |
| `ak_bmsc` | Bot Manager | Session mitigation state co-cookie; often issued alongside `_abck` after a successful challenge exchange | Deployment-specific, treat as opaque |
| `bm_sv` | Bot Manager | Session-verification token issued after the client passes a subordinate check | Opaque |
| `bm_mi` | Bot Manager | Mobile-integration marker seen on some deployments running the mobile SDK path | Opaque |
| `bm_so` | Bot Manager | Session-origin / short-lived binding observed on some tenants | Opaque |
| `bm_lso` | Bot Manager | Long-session origin binding observed on some tenants | Opaque |
| `sbsd` | Server Bot Detect (SBSD) | Companion product to Bot Manager on some deployments; separate short-lived challenge cookie | Opaque |
| `sbsd_o` | SBSD | Long-form SBSD state | Opaque |
| `RT` | Real-User Monitoring (Boomerang) | Not Bot Manager; RUM/beacon library. Presence indicates the Boomerang RUM script is also on the page, which itself is a signal source for Akamai | JSON-like `key=value&…` payload |

## `_abck` — the one everyone talks about

`_abck` is the visible verdict cookie for a Bot Manager challenge cycle. Its **internal structure is deployment-specific and Akamai has never publicly documented it**; treating the raw string as opaque is the right default.

Two things ARE publicly-verifiable from watching real traffic:

- The **presence of `~0~`** as a substring after a successful sensor exchange is the coarse "accepted" marker most researchers key on ([scrapfly](https://scrapfly.io/blog/posts/akamai-bot-manager-understanding-abck-cookies-and-sensor-data), [Kameleo](https://kameleo.io/glossary/akamai-abck-cookie)).
- The **`~-1~` marker** is the pre-challenge / rejected state.

`~0~` on the sensor endpoint does NOT guarantee downstream protected endpoints will accept the session — Akamai risk-scores the whole request context (IP reputation, behavioural coherence, cross-layer consistency). A "low-trust `~0~`" is a real observed state where the sensor cleared but strict endpoints still 403. That's Akamai working as designed, not a cookie parsing bug.

## `bm_sz` — the challenge-parameter carrier

`bm_sz` arrives before the sensor script executes. The client script parses it and reads at least one field back as a per-session identifier used inside the encrypted sensor payload. This is why session-independent replays fail: `bm_sz` is bound to the session that generated it, and any downstream cookie the script mints inherits that binding.

The delimited `~`-field format has been widely observed but the semantics of each field position are deployment-specific — treat as opaque and read only the structural fact that fields exist.

## Failure-mode taxonomy (what usually causes cookie rejection)

Failure mode | Underlying cause
---|---
Stale value | Client sent an expired `_abck` well past the deployment TTL
Replay mismatch | `_abck`/`bm_sz`/`ak_bmsc` from one IP or TLS fingerprint reused from another
Cross-layer conflict | Cookie says accepted but the underlying TLS/H2/UA don't match the browser that generated it
Integrity failure | The cookie contents were modified in transit or by tooling

Scrapfly's summary of this is the cleanest: **"Akamai blocking is usually caused by cross-layer mismatch rather than one isolated signal."**

## Defender read-notes

- If you administer an Akamai tenant, the cookie names on your traffic tell you which BM products are active (Bot Manager Standard, SBSD, Boomerang RUM).
- If you're an authorised pentester documenting a target's stack, the presence of `sbsd`/`sbsd_o` indicates the customer bought SBSD in addition to BM — a signal about their bot posture.
- Missing `_abck` on a route that Bot Manager is supposed to cover is worth investigating on your own tenant: either the route is unprotected, or the request never reached the challenge layer.

## References

See [`references.md`](references.md).
