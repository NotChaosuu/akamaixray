# References

Every claim in `docs/` is grounded in one of the sources below. Direct citations use inline hyperlinks in each doc; this file consolidates them and adds context.

## Primary Akamai references

- **Akamai's own blog: "Bots Tampering with TLS to Avoid Detection"** — [akamai.com/blog/security/bots-tampering-with-tls-to-avoid-detection](https://www.akamai.com/blog/security/bots-tampering-with-tls-to-avoid-detection). Akamai publicly acknowledging TLS-layer signals as part of their detection stack.
- **Akamai Bot Manager product page** — [akamai.com/products/bot-manager](https://www.akamai.com/products/bot-manager). Vendor-provided feature summary.

## Third-party research repos

- **[Edioff/akamai-analysis](https://github.com/Edioff/akamai-analysis)** — the reference deep-dive on Bot Manager **v2**. Contains signal categories, detection pipeline, cookie lifecycle, and a sanitized sensor structure doc. `akamaixray` cites this heavily for v2 material and does not attempt to duplicate its ground.
- **[glizzykingdreko/akamai-v3-sensor-data-helper](https://github.com/glizzykingdreko/akamai-v3-sensor-data-helper)** — the reference implementation for the v3 sensor encoding and decoding side. Different scope from this repo (helper library vs defender atlas).
- **[DalphanDev/akamai-sensor](https://github.com/DalphanDev/akamai-sensor)** — sensor reverse-engineering helper repo.

## Blog and article references

- **[Scrapfly: Understanding _abck Cookies and Sensor Data](https://scrapfly.io/blog/posts/akamai-bot-manager-understanding-abck-cookies-and-sensor-data)** — the cleanest public summary of the cookie lifecycle and cross-layer-mismatch framing. Direct source for the "cross-layer mismatch rather than one isolated signal" wording.
- **[Kameleo glossary entry on the _abck cookie](https://kameleo.io/glossary/akamai-abck-cookie)** — short defender-oriented summary of what `_abck` is.
- **[glizzykingdreko: Akamai v3 Sensor Data Deep Dive](https://medium.com/@glizzykingdreko/akamai-v3-sensor-data-deep-dive-into-encryption-decryption-and-bypass-tools-da0adad2a784)** — Medium article covering the v3 encryption/decryption side in depth.
- **xkiian's 2026 posts on the TLS-first thesis** — surfaced in current web search under titles like "Bypassing Akamai v3 sensor_data with TLS in 2026 — why the deobfuscator is a trap." Source of the TLS-first framing this repo builds on in [`tls-thesis.md`](tls-thesis.md).

## Related tools

- **[wafprobe](https://github.com/NotChaosuu/wafprobe)** — companion repo for TLS-fingerprint diversity and multi-vendor WAF/antibot detection. Where `akamaixray` reads Akamai's response-side, `wafprobe` explores the request-side TLS fingerprint dimension.
- **[apkxray](https://github.com/NotChaosuu/apkxray)** — Android APK triage; also documents the Akamai BMP mobile SDK on the Android side (via `ignassew/bapka`-style analysis).

## Cited public research on bot detection fingerprinting

- Antoine Vastel's academic work on browser fingerprinting and bot detection — the theoretical basis for a lot of what modern Bot Manager products do at the client level.
- Sannier / Vastel / Rouvoy series on canvas / audio / WebGL fingerprinting entropy.

## Disclosure

This repo is research and education. It cites its sources so readers can verify claims and go deeper themselves. If you spot a factual error, PRs and issues are welcome; see the repo README for the contribution note.
