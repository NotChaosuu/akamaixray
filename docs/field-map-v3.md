# v3 sensor_data field map (semantic)

A defender-oriented reference for the v3 sensor payload plaintext. Names each field, groups by semantic role, and tells you **what the field carries** — not how to synthesise it. No formulas, no keys, no encryption inverses. See [`../README.md`](../README.md) for the reason this line matters.

If you want the encryption/decoding side, [glizzykingdreko/akamai-v3-sensor-data-helper](https://github.com/glizzykingdreko/akamai-v3-sensor-data-helper) is the reference implementation for that ground.

## Payload shape at a glance

The v3 payload rides in the `sensor_data` value POSTed to a per-session sensor endpoint. The outer wire format is a semicolon-delimited envelope:

```
3;0;1;0;<cookie_hash>;<b64 key>;<counts>;<ENCRYPTED>
```

- `3` — payload family version marker (v3)
- `0;1;0` — layout/versioning triplet, stable per script family
- `<cookie_hash>` — session-binding hash read out of `bm_sz`
- `<b64 key>` — key material used by the script's own cipher
- `<counts>` — per-session sensor counters
- `<ENCRYPTED>` — the plaintext object described below, encrypted

Everything below refers to the **plaintext object** — the semantics of what the sensor collects. What Akamai actually decides on lives inside this object.

## Total field count

Current observed variants carry ~52 top-level fields (older variants published in public deobfuscators expose ~30 core fields — the newer variants added a handful of timing / constant fields, and a couple of decoys). Approximately **45 of those 52 fields are effectively static per script + device**, and the remaining ~7 vary per request. That ratio is the story: most of the payload is a device/script fingerprint replayed each time.

## Field categories

Every field falls into one of five semantic buckets:

| Category | What it is | Example fields |
|---|---|---|
| **Script constants** | Values baked into the specific script build; identical across every request on that script variant | `ver`, `jsrf` (the string identifier, not the numeric checksum), `signals`, timing baseline fields |
| **Device profile** | Deterministic derivations of the browser environment (navigator, screen, WebGL, canvas, storage, permissions) | `fpt`, `din.*`, `sde`, `dsi`, `pde`, `per`, `hls`, `wsl` |
| **Session timing** | Values derived from a single per-session start timestamp; stable relative to that clock | `sts`, `hz1`, `dd2`, `ww8`, `hal` |
| **Behavioural telemetry** | Values that reflect user interaction (mouse, keyboard, touch, focus) or its absence | `mev`, `mc`, `kc`, `tc`, `pc`, `tovl`, `mevl`, `tst`, `ssts`, `delt`, `ajr` |
| **Server-validated** | Values Akamai's server checks with logic held server-side; forging these consistently is the "hard" bar | `dvc`, `jsrf` (numeric checksum pair), a subset of the `s###` randoms |

## Field-by-field reference

The purpose of this list is **naming and semantics only**. If you're a defender parsing your own captures, or a researcher trying to understand what Akamai looks at, this tells you what each label means. It does not tell you how each value is computed.

### Script constants

| Field | Semantic |
|---|---|
| `ver` | Script version marker. Hardcoded into the script build. All requests from the same script variant carry the same value. |
| `signals` | Anti-instrumentation summary string. Reports whether the script observed automation markers in the environment. `"0"` is the clean-environment value. |
| `it`, `o9`, `tsd`, `fct`, `srd`, `tid` | Counter and mode fields the script sets to fixed values during a normal run. |
| `hea` | Header-context indicator. Usually empty in normal browser flow. |
| `rval`, `rcfp` | Report-context values, `"-1"` in clean runs. |
| `s002`, `s003`, `s017`, `s127`, `s162`, `s163` | Newer per-script constants added in more recent variants. |

### Device profile — `fpt` and `din`

- `fpt` — a compact, joined string of environment booleans and small values: Brave brand string, plugin count, `sessionStorage`/`localStorage`/`indexedDB` presence, timezone offset, RTC presence, color and pixel depth, `cookieEnabled`, `javaEnabled`, `doNotTrack`. In a clean browser this string is deterministic — same browser fresh session yields the same `fpt`.
- `din` — a nested object of navigator + screen fields. Every subfield below reflects a single web API reading, taken once per session:

| `din.*` field | Web-API source |
|---|---|
| `ua` | `navigator.userAgent` |
| `pha` | `location.protocol` + host derivation |
| `nps` / `nal` / `nap` / `npl` | Plugin summaries |
| `wdr` / `dau` / `asw` / `ash` / `swi` / `she` | Screen dimensions and available area |
| `wiw` / `wih` / `wow` / `woh` | `window.inner*` / `window.outer*` |
| `adp` / `ibr` / `ran` | Browser flags and one random value (the only intentionally-random field in this cluster) |

Every `din.*` reading is deterministic from a real browser; farm-style profiles that mismatch across these fields (e.g. inner > outer, DPR of `1` but a 4K screen) are trivial to score down.

### Fingerprint hashes

| Field | What it hashes |
|---|---|
| `fpc` | Hash of the `fpt` string above |
| `ucs` | Hash of the User-Agent string |
| `xag` | Additional UA-derived signal |
| `nfas` | Navigator feature bitmask (e.g. `25182352`) — a packed representation of which optional web-platform features the browser reports |

These are all stable per (browser, script) — Akamai treats consistency of these hashes across requests as a real-user signal.

### Environment probes

| Field | Purpose |
|---|---|
| `sde` | Selenium / automation probe string; reports whether known automation markers are present. Clean browsers emit a fixed shape like `"0,0,0,0,1,0,0"` |
| `per` | Permissions-API responses summary |
| `dsi` | WebGL introspection summary (vendor, renderer, extensions) |
| `wsl` | Storage-and-quota API summary |
| `hls` | Headless-browser tell aggregator |
| `pde` | `Object.keys(getOwnPropertyDescriptors(navigator))` |
| `fwd`, `eem`, `inf`, `pur`, `mwd` | Additional environment probes (frame context, media element hooks, iframe context, page URL, misc widths) |

### Session-timing family

Derived from a single per-session start timestamp:

| Field | Semantic |
|---|---|
| `sts` | The start timestamp itself |
| `hal` | A halved-timestamp derivation used elsewhere in the payload |
| `hz1`, `dd2`, `ww8` | Coarse-grained derivations of the same clock (base-36 encodings + divisions), used as consistency checks |

The point of these fields is **temporal consistency** — Akamai wants the timing to be internally coherent. Sensors where the timings don't derive from the same clock are trivially caught.

### Behavioural telemetry

| Field | What it summarises |
|---|---|
| `mev` | Serialised mouse-event log (per event: index, type, time, x, y) |
| `mc` / `kc` / `tc` / `pc` | Mouse / keyboard / touch / pointer event counters |
| `tovl` | Aggregate over the mouse event log; roughly a coherence checksum |
| `mevl` | Combined event log length signal |
| `tst`, `ssts`, `delt` | Cumulative and last-event timestamps |
| `ajr` | A three-int summary of the event log where `ajr[2]` is the `tovl` aggregate |

Sensors from bot farms with no mouse interaction default these to their empty-flow values, which is a strong signal on its own.

### Server-validated fields — the hard bar

- `dvc` — a server-validated token minted by an obfuscated routine using session inputs (start timestamp, user-agent slice, per-sensor counters) run through a hash and a filtered alphabet. Akamai holds the validating logic server-side; forged `dvc` values inconsistent with the rest of the payload are the primary "we know this is fake" gate.
- `jsrf` numeric pair — a checksum-shaped pair derived from the script's own view of the payload; content-covering, so if you modify anything the script computed over, this stops matching.
- A handful of `s###` fields (e.g. `s148`, `s150`, `s151`, `s153`) — random-looking timing fields that are nevertheless server-validated for plausibility (monotonic, sensible magnitude).

## Cross-layer consistency is the real story

Every field above matters less than the **cross-field coherence** Akamai computes across all of them:

- `din.ua` must match the User-Agent the request actually carries.
- `fpt` and the derived `fpc` must match a real browser profile — mismatches between plugins, screen, timezone, and Brave brand string get caught.
- `din.wow` / `din.woh` / `din.wiw` / `din.wih` must be internally consistent (inner ≤ outer, DPR-corrected).
- Timing fields must all derive from the same clock.
- The TLS and HTTP/2 fingerprint of the actual request (which is not IN the payload) must match what the User-Agent claims. See [`tls-thesis.md`](tls-thesis.md).

Any single field flipping is not the giveaway. The **combination not matching a real browser** is.

## What this map does NOT contain

Deliberately absent from this repo:

- No formulas for computing any of the constants, hashes, or timers.
- No encryption or decryption code.
- No `dvc` reversal, `jsrf` checksum reconstruction, or `file_hash` recovery.
- No script deobfuscation output.

The purpose here is to name the surface so a defender or authorised tester can read a captured payload and know what they're looking at. Building the payload is out of scope on purpose.

## References

See [`references.md`](references.md).
