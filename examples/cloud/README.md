# examples/cloud

The SiteHost Cloud Container endpoints, as a numbered journey. Same
shape as `examples/server`: every check must be able to fail, and a
failed check exits non-zero.

## The steps

```
10  discover   find a location and product that offer containers
40  read       walk every read path and check the shapes agree
80  probe      provoke rejections deliberately and record them
```

Run with no arguments to print the map. `journey` runs them in order.

Everything here is read-only. The provision, stack, database, sshuser
and delete steps are not written yet, so this journey reads and probes
a container that already exists rather than creating one.

## Why the probe step exists

A hand-written fixture encodes what we believe the API accepts, so a
test built on one can only confirm the belief that produced it. Several
bugs in this SDK survived a green suite that way.

Recorded *rejections* are the half a mock cannot supply, because
obtaining one means being wrong on purpose. They are also free here:
every probe addresses a server that cannot exist, so nothing is
created and nothing needs cleaning up.

Running it is what established that the list filters are optional but
validated, that `cloud/stack/get.json` takes `server` where its
siblings take `server_name`, and that template id `0` is a real
template rather than a null id.

## Environment

| Variable | Default |
|----------|---------|
| `SH_API_KEY` / `SH_CLIENT_ID` | required |
| `SH_LOCATION` | `AKLNCT` |
| `SH_PRODUCT` | `CLDCON4-P` |
| `SH_SERVER` | the first cloud server on the account |
| `SH_BASE_URL` | the public API |
| `SH_RECORD_DIR` | — (see below) |

### Recording

Set `SH_RECORD_DIR` and every call is written there as JSON, rejections
included.

**Those files hold live data.** Secret-bearing fields are blanked on the
way to disk, but that is a reduction rather than a guarantee: the
recordings still contain real server names, addresses, database names,
usernames, home directories and key material.

Point it outside the repository — `SH_RECORD_DIR=$(mktemp -d)` — and run
anything derived from it through `internal/scrubtool` before committing
or sharing. `**/recordings/` is in `.gitignore` as a second line of
defence, not the first.

## Logging discipline

Counts, ids and shapes only. A database name, a username or an address
does not belong in an example's output, and the recordings are where
the real values live.
