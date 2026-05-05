# `bandwidth/get_ip_list.json` mangles IPv6 addresses

## Symptom

`bandwidth/get_ip_list.json` returns IPv6 addresses with `:` replaced by
`.` in both the `ip_addr` field and the response map's key. Double colons
(`::`) become double dots (`..`).

| API output | What it should be |
|---|---|
| `2403.7000.8000.300..ce/128` | `2403:7000:8000:300::ce/128` |
| `2403.7000.8000.c00..9b/128` | `2403:7000:8000:c00::9b/128` |

The address itself is otherwise correct — same hextets, same prefix
length — but the separator substitution makes the string an invalid
IPv6 representation.

## Round-trip impact

The most natural workflow is broken end-to-end: read an IP from
`get_ip_list`, pass it to any IP-input endpoint. The input endpoint
validates strictly and rejects the mangled form:

```
GET /bandwidth/get_usage_by_month.json?ip_addr=2403.7000.8000.300..ce/128
=> 400 Please specify a valid IP address.
```

Same break would apply to any other endpoint that accepts an IPv6
address as input (RDNS, firewall rules, etc.) when the address comes
from `get_ip_list`'s output.

## Scope (verified live)

The bug is **localised to this one endpoint**. Other endpoints that
emit IPv6 addresses are unaffected:

| Endpoint | IPv6 representation |
|---|---|
| `bandwidth/get_ip_list.json` | **Mangled** (`:` → `.`) |
| `bandwidth/get_usage_summary.json` | Canonical (colons preserved, even in map keys) |
| `server/list_servers.json` | Canonical |

So the bug is almost certainly a single buggy serialisation step in
the `get_ip_list` handler — not a shared API output layer. Fixing it
server-side would be a one-place change.

## Workaround

For now, callers passing addresses from `models.IPAddress.IP` to other
endpoints should either filter to IPv4 or canonicalise the address
(replace `..` with `::`, then remaining `.` between hextets with `:`).
Examples in this repo (e.g. `examples/bandwidth/main.go`) filter to
IPv4 to avoid the issue.

## Future SDK fix

A targeted normaliser on `models.IPAddress.IP` (custom `UnmarshalJSON`
that detects the mangled form and corrects it) would make the field
safe to round-trip. Worth doing once the upstream API bug is either
filed or fixed — the SDK layer should not silently mask a server-side
bug without it being tracked.
