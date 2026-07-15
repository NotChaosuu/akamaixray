# Script variant identification

The sensor script rotates per session on many current Akamai deployments. This doc explains what a "variant" is, what markers are observable from outside the script, and how `akamaixray` classifies them.

## What "variant" means here

A **variant** is a specific build of the client sensor script served by an Akamai tenant during a specific session. Two requests to the same tenant, minutes apart, can serve variants that differ in:

- **Byte size** (observed range 500–560 KB for current v3, with common sizes clustered around 526k / 548k / 549k / 551k / 553k)
- **Internal string-table layout** (the packed strings the obfuscated script decodes are reordered)
- **Field order** in the plaintext of the encrypted payload (specifically the `din` sub-object's key order)
- **File-hash seed** (the value the script uses as its own cipher seed — different per session)

What **does not** change per session:

- The high-level structure of what the script collects (the field set in [`field-map-v3.md`](field-map-v3.md) is stable across variants of the same generation)
- The transport envelope (`3;0;1;0;…` prefix, POST target being same-origin, `text/plain` body)
- The cookie family the deployment mints (`_abck`, `bm_sz`, `ak_bmsc`, optional `sbsd`)

## Why variants matter

For a defender or authorised tester:

- Stale monitoring assumptions rot fast. Any tool that hard-codes a specific script URL, specific script hash, or a specific parse offset breaks in weeks.
- Comparing script fingerprints across time is a useful drift signal — if a tenant's script size distribution suddenly shifts (e.g. from 548k-ish to 590k-ish), the tenant likely picked up a new Bot Manager update.
- Documenting what version of Bot Manager a specific customer runs is genuinely useful intel; sizes cluster by release cadence.

## Observable markers `akamaixray` records

`akamaixray identify` looks for the following signals in a target's page and script:

| Marker | Where it comes from | What it tells you |
|---|---|---|
| Same-origin `<script src>` that responds with a 500–560 KB obfuscated JS body | Page HTML + a GET to the script URL | Almost certainly the Bot Manager v3 sensor script |
| Presence of the semicolon-envelope prefix `3;0;1;0` anywhere in the response body of a subsequent POST | Sensor POST trace (only relevant if you're testing your own tenant) | Confirms v3 payload family |
| `_abck`, `bm_sz`, `ak_bmsc` set together on the initial GET | Response `Set-Cookie` headers | BM Standard is active on this route |
| `sbsd` and/or `sbsd_o` set alongside | Response `Set-Cookie` headers | SBSD is also active on this route |
| `server-timing` header including `ak_p` or similar Akamai edge markers | Response headers | Traffic is passing Akamai's edge, regardless of BM state |
| `x-akamai-transformed` on responses | Response headers | Akamai edge is actively transforming (compressing, image-modifying, or bot-managing) responses |

## What `akamaixray` deliberately does NOT do

- Does not parse the obfuscated script beyond size and headline markers.
- Does not attempt to compute `file_hash`, decode string tables, or extract cipher constants.
- Does not fetch the sensor POST endpoint or send anything to it.

The identifier is a passive-observation tool for defenders assessing their own or authorised-scope Akamai deployments. It reads what the tenant chooses to expose in HTML and response headers and stops there.

## References

See [`references.md`](references.md).
