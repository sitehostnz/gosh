# server — the whole lifecycle, as a numbered journey

Provision two servers, swap their IP addresses, complete the cutover
inside the guests, and delete everything. Every step asserts something
that can fail.

```sh
go run ./examples/server          # print the journey map, touch nothing
```

That is the safe default and the best place to start: it prints what has
to happen, in what order, and which steps write.

## The numbering is the documentation

```
10  register an SSH key      must precede provisioning
20  discover / preflight     before provisioning, any order
25  inventory                read-only; safe anywhere
30  provision the pair       needs 10 and 20
35  upgrade disk / plan      needs a provisioned server; opt-in
40  firewall / netconfig     after provisioning, any order
45  security group           creates one, proves it filters, removes it
47  snapshot                 take, set lifetime, restore, delete
48  label                    change a label and put it back
50  prestage the guests      needs 10's key; MUST precede 60
60  swap the addresses       needs a provisioned pair
70  reboot via the API       needs 50 and 60
80  verify over SSH          needs 70
90  delete                   always last
```

Steps sharing a number are interchangeable. A higher number needs the
lower ones. These constraints are not written down anywhere in the API,
and they are the part that catches people out — particularly 50 before
60, for which see below.

## Running it

```sh
export SH_API_KEY=...
export SH_CLIENT_ID=...

# the whole thing: provisions two real servers, then deletes them
SH_EXAMPLE_ALLOW_PROVISION=1 go run ./examples/server journey

# one step, against servers you already have
SH_SERVER_A=web1 SH_SERVER_B=web2 go run ./examples/server netconfig
```

Read-only steps need no opt-in. Steps marked `WRITES` in the map refuse
to run without `SH_EXAMPLE_ALLOW_PROVISION=1`.

| Variable | Default |
|----------|---------|
| `SH_LOCATION` | `AKLNCT` |
| `SH_PRODUCT` | `LHPVS1` |
| `SH_IMAGE` | discovered at run time |
| `SH_DISTRO` | `ubuntu-noble` |
| `SH_SSH_USER` | resolved from product family + distro; see `server.LoginUserFor` |
| `SH_LABEL_A` / `SH_LABEL_B` | `gosh-journey-a` / `-b` |
| `SH_SERVER_A` / `SH_SERVER_B` | — (single-step runs) |
| `SH_SSH_KEY_FILE` | — (a private key the servers already trust, for standalone 50 and 80) |
| `SH_DISK_LABEL` | discovered from the server (`upgrade-disk`) |
| `SH_UPGRADE_PLAN` | — (required by `upgrade-plan`) |
| `SH_DELETE_SERVERS` | — (comma-separated; the only way `delete` touches a server this run did not create) |
| `SH_BASE_URL` | the public API |
| `SH_RECORD_DIR` | — (see the warning below) |

### Recording what the API actually said

Set `SH_RECORD_DIR` and every request and response is written there as
JSON, including the rejections. It is how the fixtures under
`pkg/api/**/testdata` were produced, and it is worth turning on while
changing anything here — a hand-written fixture can only confirm the
belief that produced it.

**Those files hold live data.** Secret-bearing fields are blanked on the
way to disk, but that is a reduction rather than a guarantee: the
recordings still contain real server names, addresses, database names,
usernames, home directories and key material, and a field naming a
secret in an unanticipated way will be written in the clear.

So point it outside the repository — `SH_RECORD_DIR=$(mktemp -d)` — and
run anything derived from it through `internal/scrubtool` before
committing or sharing. `**/recordings/` is in `.gitignore` as a second
line of defence, not the first.

## Getting an API key

Keys are created in the Control Panel, not through the API:

1. Sign in to the Control Panel.
2. Go to **Account → API**.
3. Create a key and note the **client ID** shown with it.
4. Restrict the key by IP address if the automation runs from a fixed
   address.

See the [API docs](https://docs.sitehost.nz/api/) and the
[Knowledge Base](https://kb.sitehost.nz/developers/api). Keep the key out
of source control, and prefer a dedicated key per automation so it can be
rotated on its own.

## Why 50 must precede 60

The one ordering rule worth reading before you write your own version of
this.

Changing a server's address does not touch the guest. The platform
generates the guest's network configuration, then hands it over as a
static file — cloud-init provisions the machine on first boot and does
not own the addressing afterwards. So the moment an address moves, the
guest is still configured for an address it no longer has, and is
unreachable.

There is therefore no "log in afterwards and fix it". The configuration
has to be written **before** the swap, while the server can still be
reached, and deliberately left unapplied. Step 70 then reboots each
server **through the API**, which needs no access to the guest at all, so
there is no window to race.

Step 50 does not render that configuration itself. It asks the API for
the config of the *other* server — which is exactly what this server will
need once the addresses move — and writes that. For high-performance
servers the generated netplan matches the interface on a MAC address, so
hand-writing the file and guessing an interface name does not work.

## What the journey records

Behaviour that is not documented elsewhere, each asserted or reported by
the step that meets it:

- **Moving one address at a time is refused** on legacy Xen (LINVPS)
  products when the servers sit in different networks:
  `Im sorry this address space cannot be used here.` The constraint is on
  the *target server's existing addresses*, not on the address — release
  both servers first and the same calls succeed.
- **That refusal is only visible in one window** — after the address is
  freed but before the target is emptied. Probe it earlier and the
  in-use error fires instead.
- **A server may hold zero addresses** in between.
- **High-performance products do not enforce that constraint**, which
  looks like a missing validation. The journey deliberately does *not*
  exercise it there — see `60_swap.go`.
- **Security groups are high-performance only.** Standard products
  reject `server/firewall/get` outright.
- **A server's name is not its label** — truncated, with a digit
  appended on collision. Use the name the provision call returns.
- **A server cannot be deleted while provisioning**, so step 90 retries
  rather than leaking it.
- **The login user is not in the API**, and depends on the product
  family as well as the distro: the same Ubuntu image is `ubuntu` on
  high-performance and `root` on legacy Xen (LINVPS). The
  standard-performance (SVS) tier was not tested. See
  `server.LoginUserFor`.

## Safety

- Read-only steps need no opt-in; every step marked `WRITES` refuses to
  run without `SH_EXAMPLE_ALLOW_PROVISION=1`.
- Step 90 always runs, even when an earlier step fails, and reports
  loudly if a delete did not succeed. It deletes only what this process
  provisioned. To delete a server it did not create, name it explicitly
  in `SH_DELETE_SERVERS` — setting `SH_SERVER_A`/`SH_SERVER_B` for a
  read-only step is never taken as consent to destroy anything.
- The SSH key is generated per run, registered, and deleted at the end.
  The private half never touches disk.
- Host keys are not verified — the servers are created and destroyed
  within the run. Do not copy that into anything longer-lived.
