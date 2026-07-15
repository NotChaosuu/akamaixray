# Akamai response headers a defender should recognise

Akamai's edge exposes a lot of state in response headers. Many of those headers are the fastest way to identify who runs Akamai, what's active on a route, and what the edge is doing to a response. This doc names the ones that matter and what they carry.

## Edge presence

| Header | Meaning |
|---|---|
| `server-timing` | Akamai edge nodes emit rich `server-timing` entries. Substrings like `ak_p`, `origin`, `cdn-cache` in server-timing are strong Akamai markers. |
| `x-akamai-transformed` | The response body was transformed by an Akamai edge policy (image-manager compression, JS/CSS minification, or bot-management response mutation). |
| `x-cache` / `x-cache-key` / `x-check-cacheable` | Akamai edge cache diagnostics. Present on many Akamai deployments even when other markers are stripped. |
| `x-akamai-request-id` | Correlation ID for a specific edge request; useful if you're on the tenant side and need to look up a request in log tail. |
| `x-served-by` | Sometimes exposes edge node identity. |

Presence of `server-timing` with Akamai-shaped values is the single most reliable header-only signal that traffic is transiting Akamai. Absence of these headers does **not** rule Akamai out — many tenants strip them at the edge.

## Bot Manager markers

| Header | Meaning |
|---|---|
| `x-akamai-bot-*` | Occasional Bot Manager markers on some tenants; not universally emitted. |
| `x-akamai-mgmt-*` | Management-layer markers. |
| `set-cookie: _abck=…` / `bm_sz=…` / `ak_bmsc=…` / `sbsd=…` | The definitive Bot Manager signal. See [`cookies.md`](cookies.md). |

The cookie set alone is more reliable than any Akamai-labelled header for identifying Bot Manager, because Akamai actively suppresses BM headers on many deployments to avoid signalling to attackers.

## `server-timing` shapes worth logging

Real Akamai `server-timing` values look like semicolon-delimited entries with fixed prefixes. Examples of prefixes that commonly appear:

- `cdn-cache; desc=HIT` / `desc=MISS`
- `edge; dur=<ms>`
- `origin; dur=<ms>`
- `ak_p; desc="…"` — Akamai proprietary marker with a coded value describing edge processing

`akamaixray` treats any `server-timing` value with `ak_p` in it as a confirmed Akamai edge marker.

## Cache diagnostic headers

Akamai's cache-control diagnostics can appear even on API responses:

- `x-cache: TCP_HIT` / `TCP_MISS` / `TCP_MEM_HIT`
- `x-cache-key: /L/…` — the cache key format Akamai uses internally

These are not sensitive on their own — they're operator-facing diagnostics — but they're strong evidence of an Akamai edge in path.

## `Vary` and `Content-Encoding`

Nothing Akamai-specific here on its own, but pay attention to:

- Aggressive `Vary` values on API responses often correlate with Bot Manager being on the route (BM's decision depends on cookies + UA + client hints, so those get added to `Vary`).
- `content-encoding: br` (Brotli) on responses combined with Akamai edge markers is the standard-issue modern deployment.

## What `akamaixray` reads

`akamaixray headers <url>` performs a single GET and classifies the response headers into:

- Confirmed Akamai edge markers
- Confirmed Bot Manager cookies
- Possible Bot Manager markers (weak signals)
- SBSD markers
- RUM / Boomerang presence

No behavioural probing. No additional requests beyond the single GET.

## References

See [`references.md`](references.md).
