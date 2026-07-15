# akamaixray

Passive identification of Akamai Bot Manager on a target, and a defender-oriented reference for what the platform actually exposes on the wire. Point it at any URL and it tells you whether Akamai's edge is in path, whether Bot Manager is active on that route, which family (v2 / v3) the deployment most likely runs, and what every observable cookie and header semantically means.

The docs under [`docs/`](docs/) are the main deliverable — a defender's atlas of Akamai artefacts. The CLI is the small tool that reads those artefacts off a target for you.

```
$ akamaixray identify https://www.whirlpool.ca/

  akamaixray verdict
  target: https://www.whirlpool.ca/

  Akamai Bot Manager (family unknown from single GET)

  bot manager cookies
    _abck — Mitigation-state verdict token
        verdict: pre-challenge / rejected (~-1~ marker present)
    ak_bmsc — Session mitigation-state co-cookie
    bm_sz — Challenge-parameter bundle read by the client script

  edge markers
    server-timing: cdn-cache; desc=HIT
        server-timing with Akamai edge markers (cdn-cache / ak_p / edge / origin)
    server-timing: ak_p; desc="1784060120450_..."
        server-timing with Akamai edge markers (cdn-cache / ak_p / edge / origin)
    x-akamai-transformed: 9 - 0 pmb=mNONE,1mRUM,2
        Response body was transformed by an Akamai edge policy

  notes
    - Bot Manager cookies present but no sensor script observed in page HTML.
      Akamai often lazy-loads the sensor script from its challenge flow — a
      plain GET may not include it. On current deployments Bot Manager is most
      likely v3.
```

## Install

```bash
go install github.com/NotChaosuu/akamaixray/cmd/akamaixray@latest
```

Or clone + build:

```bash
git clone https://github.com/NotChaosuu/akamaixray.git
cd akamaixray
go build -o akamaixray ./cmd/akamaixray
```

Requires Go 1.22+. No external dependencies — the whole thing is stdlib.

## Usage

```bash
akamaixray identify <url>       # verdict + family + script hints
akamaixray cookies  <url>       # semantic labels for observed BM cookies
akamaixray headers  <url>       # Akamai-relevant response headers
akamaixray scan     <url>       # everything (same as identify)

# flags on any subcommand
--json                          # emit JSON
--timeout N                     # request timeout in seconds (default 15)
--ua "..."                      # override User-Agent
--method M                      # HTTP method (default GET)
--body "..."                    # request body, or @file to read from a file
--header "Name: Value"          # extra request header (repeatable)
```

### Homepage looks clean? Try the login route

On a lot of tenants the Bot Manager cookies only show up on the protected route, not the homepage. A homepage GET comes back "edge present, BM not observed" even though the site clearly runs Bot Manager — because the challenge only fires on `/login`, `/checkout`, `/api/auth`, and the like. Point a POST at that endpoint instead:

```bash
# form-encoded login endpoint
akamaixray identify --method POST \
  --header "Content-Type: application/x-www-form-urlencoded" \
  --body "email=x@y.com&password=x" \
  https://target.com/login

# JSON auth endpoint, body read from a file
akamaixray identify --method POST \
  --header "Content-Type: application/json" \
  --body @creds.json \
  https://target.com/api/auth
```

Use throwaway values — the point is to observe which cookies the endpoint hands back, not to authenticate. akamaixray never retries, never solves anything, and sends exactly the one request you asked for.

One request per invocation. No probing of the sensor endpoint, no encoding, no request-forgery. If you want the request-side TLS / HTTP/2 fingerprint story instead, that's what [wafprobe](https://github.com/NotChaosuu/wafprobe) is for — the two are companion tools.

## What's in `docs/`

Substantive reference material, cited to public research. This is what actually makes the repo worth stars — the CLI is the sample tool that reads what the docs describe.

| Doc | Covers |
|---|---|
| [`cookies.md`](docs/cookies.md) | Every BM / SBSD / RUM cookie a defender should recognise, with semantic labels |
| [`headers.md`](docs/headers.md) | The `x-akamai-*` and `server-timing` markers, cache diagnostics, edge identity signals |
| [`sensor-endpoints.md`](docs/sensor-endpoints.md) | How the sensor script GET + POST look on the wire, why the URLs rotate |
| [`field-map-v3.md`](docs/field-map-v3.md) | Semantic map of the ~52 fields inside the v3 `sensor_data` plaintext |
| [`script-variants.md`](docs/script-variants.md) | The per-session script rotation reality, size bands, observable markers |
| [`v2-vs-v3.md`](docs/v2-vs-v3.md) | What changed from v2 to v3, and what stays the same |
| [`tls-thesis.md`](docs/tls-thesis.md) | The 2026 shift: TLS / HTTP-2 as primary detection layer |
| [`references.md`](docs/references.md) | Every source cited across the docs |

## What this repo deliberately does NOT contain

- No sensor-payload encoder or decoder.
- No file-hash extractor or brute-forcer.
- No `dvc` / `jsrf` reversal logic.
- No script deobfuscation output.
- No cookie forgery, no session-replay helpers.

The purpose here is to name and understand what Akamai's stack exposes, from the outside. If you want the encoding side, [glizzykingdreko/akamai-v3-sensor-data-helper](https://github.com/glizzykingdreko/akamai-v3-sensor-data-helper) is the reference. If you want the v2 deep dissection, [Edioff/akamai-analysis](https://github.com/Edioff/akamai-analysis) is the reference.

## Research and education

This project exists to make anti-bot detection stacks understandable — to defenders running Akamai, to authorised testers documenting a target's posture, and to researchers who want a well-cited jumping-off point instead of another orphaned blog post. Everything in the docs is derived from publicly-available research and observable network behaviour. Every claim carries a citation in [`docs/references.md`](docs/references.md).

If you spot a factual error, open an issue or a PR with the source. If you're a defender running Akamai and want a semantic label added to a cookie or header you're seeing, same thing.

## Similar Tools Which I uploaded 

- [wafprobe](https://github.com/NotChaosuu/wafprobe) — multi-vendor WAF / anti-bot detection from the request-side TLS fingerprint dimension
- [apkxray](https://github.com/NotChaosuu/apkxray) — Android APK triage (secrets, SDKs, endpoints, deep links)
- [authmap](https://github.com/NotChaosuu/authmap) — auth-flow reconnaissance for web apps
- [tlsprint](https://github.com/NotChaosuu/tlsprint) — TLS/JA3/JA4 self-fingerprint reporter

---

**Chaosuu** | I'm an independent researcher who works on anti-bot systems and web/API security.

- GitHub: [NotChaosuu](https://github.com/NotChaosuu)
- Telegram: [t.me/chaosuudev](https://t.me/chaosuudev)
- X: [@NotChaosuu](https://x.com/NotChaosuu)
- Reddit: [u/NotChaosuu404](https://www.reddit.com/user/NotChaosuu404)

## License

MIT. See [`LICENSE`](LICENSE).
