# Sensor endpoints — how the client script talks to Akamai

The v3 sensor script fetches challenge parameters, gathers its telemetry, and POSTs to a per-session endpoint. This doc names those endpoints and the observable wire pattern. **No payload generation is described.**

## The two request classes

Akamai Bot Manager's client-side surface reduces to two request classes worth naming from a defender's view.

### 1. The script fetch (GET)

A script tag pointing at a same-origin URL under the target's own domain. Two observed URL patterns dominate:

- A path that ends in a `.js` file whose name looks arbitrary (short random-looking string), often served from a rewritten subpath.
- A path that includes a marker fragment characteristic of Akamai's client library (varies per deployment).

**Sizes observed in the wild for current v3 scripts:** roughly 500–560 KB after unpacking, obfuscated. The exact size **rotates per session on many tenants** — the same target served across multiple sessions may return scripts of 526k, 548k, 549k, 551k, or 553k bytes. This rotation is one of the strongest signals that you are looking at an actively-maintained Akamai deployment and not a stale include.

### 2. The sensor POST

Every v3 script POSTs its collected telemetry to a **same-origin path that is not hardcoded**. Two properties characterise it:

- The path is discoverable from the page HTML or the script itself — do not hardcode it in tooling; parse it out.
- The POST body is `text/plain` with a JSON envelope like `{"sensor_data":"…"}`.

Response bodies from a well-formed challenge exchange are short (empty or a small acknowledgement). The **cookie state on the response** is what actually matters — a successful exchange refreshes `_abck` and adjacent cookies. See [`cookies.md`](cookies.md).

## The wire envelope for `sensor_data`

The value inside `sensor_data` is the semicolon-delimited envelope described in [`field-map-v3.md`](field-map-v3.md):

```
3;0;1;0;<cookie_hash>;<b64 key>;<counts>;<ENCRYPTED>
```

The leading `3` is the v3 family marker. Older payloads have different leading structure — see [`v2-vs-v3.md`](v2-vs-v3.md).

## Why the sensor endpoint URL rotates

Two independent rotation mechanisms exist on many tenants:

1. **Per-session sensor endpoint** — the path parsed out of the page changes across sessions. Hardcoded probes against a stale path 404 or 403 without ever touching the challenge layer.
2. **Per-session sensor script** — the script content and size change across sessions, so cached parses go stale immediately.

Both rotations are deliberate. From a defender's viewpoint they mean two things:

- If your monitoring caches Akamai's paths, invalidate on session boundaries.
- If you are an authorised tester documenting a deployment, capture the flow in one session end-to-end; do not compose it from artefacts of different sessions.

## Cross-layer binding

The sensor endpoint binds processing to the client that initiated the session. The observable effect: sending a captured `sensor_data` from a different TLS stack or a different IP than the one that received the initial `bm_sz` typically comes back with `_abck` stuck at `~-1~`, even when everything inside the payload validates.

For defenders this is a useful integrity property — a leaked cookie value cannot be replayed portably. For monitoring this is a diagnostic property — if you see `_abck` never advance past `~-1~` for a specific client fingerprint despite well-formed requests, the client's underlying network layer is the more likely mismatch than the payload contents.

## Where SBSD lives

Deployments running Akamai's SBSD product typically add a second POST class to a distinct path with its own cookie set (`sbsd`, `sbsd_o`). SBSD's short-lived challenge is separate from BM's — a target running both will show both cookie families in a single session. Cataloguing which of the two classes a target uses is part of what `akamaixray identify` is for.

## References

See [`references.md`](references.md).
