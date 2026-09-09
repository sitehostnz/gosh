# examples/cloud

The SiteHost Cloud Container endpoints, as a numbered journey. Same
shape as `examples/server`: every check must be able to fail, and a
failed check exits non-zero.

## The steps

```
10  discover           find a location, resolve the smallest container product
20  provision  WRITES  create a container server, prove it answers
30  stack      WRITES  deploy a website container, fetch it over HTTP
40  sshuser    WRITES  add a container SSH user, log into it
50  database   WRITES  create a database, user and grants
60  read               walk every read path and check the shapes agree
80  probe              provoke rejections deliberately and record them
90  delete     WRITES  tear down what this process created
```

Run with no arguments to print the map. `journey` runs them in order.

Steps marked `WRITES` need `SH_EXAMPLE_ALLOW_PROVISION=1`.

## Every write is checked outside the control plane

This is the reason the journey provisions anything. Asking the API
whether the API did what it said proves a record changed, not that
anything happened — so each writing step has a check that does not go
through the control plane at all:

| step | what actually proves it |
| --- | --- |
| 20 provision | a TCP handshake to `tcp/22` completes |
| 30 stack | the compose file round-trips; serving is proven by step 40 |
| 40 sshuser | an SSH session opens as that account, writes a marker into the container, and fetches it back over the Docker network |
| 50 database | the `mysql` client, run inside the host as the created user, names the database, writes a row and reads it back |
| 90 delete | `tcp/22` stops answering |

Completed jobs, `Up` states and returned records are logged too, since a
disagreement between them and reality is worth catching. None of them is
the assertion.

### Public routing works, but a malformed stack can break it

The contract is: `*.<ip>.sth.nz` resolves to the container server, a
proxy on the host serves ports 80 and 443, and it routes by `Host`
header using each container's `VIRTUAL_HOST`. A stack deployed by step
30 was fetched over its public hostname and answered **200**.

It was also seen failing, and the sequence is the useful part. On one
server, in order:

1. A stack deployed **without** the external `infra_default` network —
   container reported `Up`, public hostname returned the proxy's
   **503**.
2. A second stack, this time **with** the network — also **503**
   publicly for over ninety seconds, while answering **200** when
   fetched by stack name over the Docker network from inside the host.
3. The first stack was deleted. The next stack, deployed identically to
   the second, answered **200** publicly straight away.

So the 503 tracked the presence of the malformed stack, not an inability
to publish through the API — which was the first conclusion drawn here,
and it was wrong. A stack the proxy cannot route to appears to cost the
ones alongside it. That is a correlation over three deploys, not a
proven mechanism.

Two things follow. The external `infra_default` network is not optional:
the proxy lives on it, and a container that does not join it is
unreachable. And the journey's own serving check goes through the Docker
network from inside the host rather than the public hostname — the
container answering with no proxy in the path, which is a stronger
check and one that does not fail for a neighbour's mistake.

## Things the API does not tell you

- **A container server is provisioned through the *server* endpoints.**
  There is no cloud create call — it is `server.Create` with a `CLDCON`
  product code, and it is addressed by server name from then on.
- **The label is not the name.** The platform prefixes `ch-` and keeps
  the label's first nine characters: `gosh-cloud-journey` produced
  `ch-gosh-clou`. So a label that differs only in a suffix derives the
  same name — the journey randomises the front of its label for that
  reason.
- **A name freed by a delete is not immediately reusable.**
  Provisioning `ch-gosh-clou` ninety seconds after deleting a server of
  that name failed with "There was a problem creating your server,
  please contact support@sitehost.co.nz" — no indication that the name
  was the problem.
- **An unknown image code fails twenty minutes later, not at the
  request.** `ubuntu-noble.amd64` does not exist (the real code is
  `ubuntu-noble-pvh.amd64`). The provision call accepted it and
  returned a job id; the build then sat in "Configuring server" for
  twenty minutes, failed, and the platform deleted the half-built
  server. Step 10 checks the code against the catalogue first, which
  turns that into a one-second error.
- **Port 80 answers before any container exists.** The host's reverse
  proxy listens from the start and returns its own 503 until something
  is routable, so a port-80 handshake proves the proxy is up and
  nothing about your stack. Step 30 therefore rejects 502, 503 and 504
  and accepts anything else — a 403 from an empty docroot is the
  container answering.
- **The provision response carries no address.** A virtual server
  returns its IPs; a container server returns `ips: null`, so the
  address has to be read back with `server.Get` after the build.
- **`primary` is true for both address families.** A container server
  reports its IPv4 and its IPv6 address both with `primary: true`, so
  selecting on that field alone returns whichever came first. The
  journey filters for IPv4 explicitly.
- **Containers take no image.** The provision parameter is required and
  never validated, and `server.Get` reports `image: null` afterwards.
- **Database stacks have their own endpoints, and the SDK was missing
  them.** `MySQLHost` on the `cloud/db` calls is the name of a database
  stack on the same server — `mariadb1108`, `mysql57`. Those stacks are
  managed by `cloud/stack/integrated/list_all.json` and
  `cloud/stack/integrated/add.json`, now wrapped as
  `pkg/api/cloud/stack/integrated`.

  A freshly provisioned container server has none, so step 50 creates
  one. The name is the image code, and nothing else is configurable —
  the platform generates the compose itself, including the environment
  file holding the root password and the loopback port bindings.

  The reason this is worth spelling out is that every path you find
  first is a dead end that does not mention the right one:
  `cloud/stack/add.json` refuses those names as reserved and lists them,
  which reads as "you cannot create these"; the same image under an
  unreserved name gets past that and collides on the port it publishes;
  and prefix-matching `cloud/stack/list_all.json` for `mysql`/`mariadb`
  finds them on some servers and not others. Reserved *for* the
  integrated endpoints, not reserved from you.
- **Integrated stacks are not reliably in the general stack listing.**
  On an established server they appear alongside the website stacks,
  which makes prefix-matching that listing look sound —
  `examples/cloud-db-compare` does exactly that. On another server the
  listing omitted a database stack that was demonstrably running and
  accepting databases. Use `cloud/stack/integrated/list_all.json`.
- **A `ports` mapping on a web container does nothing, silently.** Web
  and application containers cannot have their ports changed — 80 and
  443 are open by default and that is the whole of it. Deploying a web
  stack with `ports: - '8931:80/tcp'` was accepted, returned a job that
  completed, kept the mapping when read back through
  `cloud/stack/get.json`, and brought the container up `Up` — while the
  port stayed closed from the internet before *and* after
  `cloud/stack/restart.json` completed. The container's own `ports`
  field reads `null` throughout, which is the only clue the API offers.

  **Service containers** are the ones that publish — Redis, Postgres,
  MongoDB, Elasticsearch and so on ship exposed-but-unpublished.
  Publishing is subject to rules: below 1024 is reserved, `3306`–`3310`
  and `8080` are refused outright (which catches the NodeJS images,
  since 8080 is the port they expose), and a published port must be
  unique across the server's containers. Security groups are not
  applied to Cloud Containers, so there is no second lever.

  Rules: <https://kb.sitehost.nz/cloud-containers/containers/ports>

  Worth knowing because the use cases are real and are not web traffic
  — MQTT, a game server, a mail service. WebSockets are the exception
  that misleads: they upgrade from HTTP, so the proxy carries them and
  no published port is needed.
- **Delete the account SSH key before the server, and wait for the SSH
  user first.** Removing a key enumerates every Cloud SSH user that
  references it and updates each one. Two consequences, both hit in a
  single teardown:

  - If any referencing user belongs to a server that has since been
    deleted, the call fails with an **empty HTTP 500** and the key
    cannot be removed at all — not by retrying, and not later. Two keys
    were stranded this way before the ordering was fixed.
  - If a referencing user has a job in flight — its own delete, for
    instance — the call is refused with "there is a job already running
    on this user".

  So the order is: delete the SSH user, **wait for its job**, remove the
  key, then delete the server. Every delete in `cloud/*` returns a job,
  and the platform holds short locks between them; not waiting is what
  produced the second failure above.

  Worth knowing outside this journey: create a container server, add an
  SSH user with a key, delete the server, then try to delete the key,
  and that key is stuck on the account permanently.
- **Grants are lowercase, and there are nineteen of them.** `SELECT` is
  rejected with "one or more of the supplied grants are invalid", which
  does not say which or why — the comparison is case-sensitive. The
  accepted set is `select, insert, update, delete, create, drop, alter,
  index, create view, show view, lock tables, create temporary tables,
  references, execute, create routine, alter routine, event, trigger`.
- **Passing every grant is not the same as passing all but one.** If
  the number of grants supplied equals the size of that set, the
  platform substitutes `ALL PRIVILEGES` rather than enumerating them.
  Worth knowing before asserting on what a user was granted.
- **`grant.Add` is for a different database, not for re-asserting.**
  `cloud/db/user/add.json` already writes the user's grant, so an Add
  against the same user and database fails on a unique constraint.
  `grant.Update` is what changes grants on a database the user already
  has.
- **`params[ssh_keys][]` means different things on different
  endpoints.** On `server.Create` it carries public key *content*. On
  `cloud/ssh/user` it carries the *id* of a key registered through
  `ssh/key/add.json`. Register the key first and pass the id.

  The two cloud calls disagree about how badly they take the mistake,
  and the reason is worth knowing: `add.json` treats each entry as an
  integer id with no coercion, so key *content* raises a type error
  that is not the not-found case it handles — it escapes as a bare HTTP
  500 with an empty body. `update.json` coerces first and so reports
  "One or more of the given SSH keys could not be found or you do not
  have access to them" for the identical input. A 500 from `add.json`
  with an empty body is the signature of passing content where an id
  was wanted.
- **An SSH user needs a container or a volume**, and the container has
  to exist first — hence stack at 30, SSH user at 40.
- **`MySQLHost` is a stack name, not a hostname** — the database stack
  on the container server, resolvable only inside its Docker network. A
  fresh container server has none, so step 50 deploys one.
- **Product codes should be discovered, not hardcoded.** AKLNCT offers
  21 `CLDCON` codes and the set varies by location;
  `server/products.json` takes `location` as a bare parameter, not
  `filters[location]`.

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
