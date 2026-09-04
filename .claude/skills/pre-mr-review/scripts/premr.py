#!/usr/bin/env python3
"""Pre-MR checks for gosh.

Runs the mechanical half of the review this repository has learned to
expect. Every check here exists because a real finding got through
without it; the docstring on each says which.

Findings are either CONFIRMED — the script proved it — or REVIEW, where
the script found a candidate a human or model has to judge. The
distinction matters: a REVIEW line is not a defect, it is a question.

Usage:
    python3 .claude/skills/pre-mr-review/scripts/premr.py [--base origin/main]

Exit status is non-zero when there is at least one CONFIRMED finding.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from dataclasses import dataclass, field


@dataclass
class Finding:
    kind: str  # CONFIRMED or REVIEW
    check: str
    where: str
    what: str


@dataclass
class Report:
    findings: list[Finding] = field(default_factory=list)

    def confirmed(self, check: str, where: str, what: str) -> None:
        self.findings.append(Finding("CONFIRMED", check, where, what))

    def review(self, check: str, where: str, what: str) -> None:
        self.findings.append(Finding("REVIEW", check, where, what))


def go_files(*roots: str, tests: bool | None = None) -> list[str]:
    """List .go files under roots. tests=True only _test.go, False excludes them."""
    out = []
    for root in roots:
        for dirpath, _, names in os.walk(root):
            if "/testdata" in dirpath or "/.git" in dirpath:
                continue
            for n in names:
                if not n.endswith(".go"):
                    continue
                is_test = n.endswith("_test.go")
                if tests is True and not is_test:
                    continue
                if tests is False and is_test:
                    continue
                out.append(os.path.join(dirpath, n))
    return sorted(out)


def read(path: str) -> str:
    with open(path, encoding="utf-8") as fh:
        return fh.read()


# --------------------------------------------------------------------
# 1. Wire contracts
# --------------------------------------------------------------------

def check_wire_contracts(rep: Report) -> None:
    """Parameters that never reach the API, or reach it twice.

    net.Encode emits only the keys named in its keys list, so a value
    added under a name absent from that list is dropped without an
    error: the call succeeds, the parameter never arrives, and nothing
    says so.

    Found this way: cloud/db.Get and cloud/db/user.Get sending client_id
    twice plus a bogus api_key; cloud/ssh/user.Update silently ignoring
    ReadOnlyConfig because the keys list named it with an array suffix;
    cloud/db.Add and .Delete sending database twice.
    """
    for path in go_files("pkg", tests=False):
        src = read(path)
        if "net.Encode" not in src:
            continue

        for m in re.finditer(r"\nfunc ", src):
            nxt = src.find("\nfunc ", m.start() + 1)
            body = src[m.start(): nxt if nxt > 0 else len(src)]
            if "net.Encode" not in body:
                continue
            fn = re.search(r"func (?:\([^)]*\) )?(\w+)\(", body)
            fn = fn.group(1) if fn else "?"

            km = re.search(r"keys\s*:?=\s*\[\]string\{(.*?)\n\t\}", body, re.S)
            if not km:
                km = re.search(r"keys\s*:?=\s*\[\]string\{([^}]*)\}", body, re.S)
            if not km:
                continue
            keys = set(re.findall(r'"([^"]+)"', km.group(1)))
            dynamic = False
            for am in re.finditer(r"keys\s*=\s*append\(keys,\s*(.*?)\)", body):
                found = re.findall(r'"([^"]+)"', am.group(1))
                keys |= set(found)
                if not found:
                    dynamic = True

            # Header.Set is not a query parameter. Matching it produced
            # three false CONFIRMED findings on the first run, which is
            # the failure this tool can least afford: a check that cries
            # wolf gets ignored, exactly like one that cannot fail.
            adds = re.findall(
                r"(?<!Header)\b(?<!Header\.)\w+\.(?:Add|Set)\(\s*\"([^\"]+)\"", body)
            adds = [a for a in adds if not re.search(
                r"Header\.(?:Add|Set)\(\s*\"" + re.escape(a) + r"\"", body)]
            where = f"{path}:{fn}()"

            for a in sorted(set(adds)):
                if a in keys or dynamic:
                    continue
                if a + "[]" in keys or a.rstrip("[]") in keys:
                    rep.confirmed(
                        "wire-contract", where,
                        f'adds "{a}" but the keys list names a different spelling; '
                        f"net.Encode drops it and the call still succeeds",
                    )
                else:
                    rep.confirmed(
                        "wire-contract", where,
                        f'adds "{a}", absent from the keys list, so it is never sent',
                    )

            for a in sorted(set(adds)):
                if adds.count(a) > 1:
                    rep.confirmed(
                        "wire-contract", where,
                        f'adds "{a}" {adds.count(a)} times, so it goes on the wire twice',
                    )

            if "req.URL.Query()" in body:
                for cred in ("client_id", "apikey"):
                    if cred in adds:
                        rep.confirmed(
                            "wire-contract", where,
                            f'adds "{cred}" on a GET; NewRequest already put it on the query',
                        )


# --------------------------------------------------------------------
# 2. Empty-collection tolerance
# --------------------------------------------------------------------

EMPTY_SHAPES = ("IsEmptyMapShape", '"[]"', "'['")


def observed_shapes() -> dict[str, set[str]]:
    """Collect the JSON shapes each wire key has actually been seen as.

    Evidence comes from the committed fixtures, which are scrubbed
    recordings: Scrub replaces values but preserves types, so "[]" stays
    "[]" and a quoted integer stays quoted. That makes testdata a
    record of what the API really sends, not of what anyone believed.
    """
    shapes: dict[str, set[str]] = {}

    def walk(o) -> None:
        if isinstance(o, dict):
            for k, v in o.items():
                if v is None:
                    t = "null"
                elif isinstance(v, bool):
                    t = "bool"
                elif isinstance(v, str):
                    t = "string"
                elif isinstance(v, list):
                    t = "empty-list" if not v else "list"
                elif isinstance(v, dict):
                    t = "empty-object" if not v else "object"
                else:
                    t = "number"
                shapes.setdefault(k, set()).add(t)
                walk(v)
        elif isinstance(o, list):
            for x in o:
                walk(x)

    for dirpath, _, names in os.walk("."):
        if "/testdata" not in dirpath or "/.git" in dirpath:
            continue
        for n in names:
            if not n.endswith(".json"):
                continue
            try:
                with open(os.path.join(dirpath, n), encoding="utf-8") as fh:
                    walk(json.load(fh))
            except Exception:
                continue
    return shapes


def check_empty_shapes(rep: Report) -> None:
    """Decoders whose field has been *observed* arriving empty.

    This API is PHP-backed, so an empty map serialises as "[]". A
    decoder handling only the object form fails on it, trading one
    decode failure for another and hiding the API's own message behind
    a JSON type error. shtypes.MaybeBoolMap was added to stop exactly
    that and still rejected "[]".

    The check is grounded in recorded evidence rather than in reading
    the code. An earlier version flagged every custom UnmarshalJSON
    that did not mention "[]", which produced nine findings on a clean
    tree, none of them backed by anything — a bug report built from
    suspicion is the failure this whole tool exists to prevent, and it
    is the failure the tool itself committed first.

    So a finding here requires a fixture in which that field actually
    arrived empty. Where no fixture covers the field at all, the
    honest answer is "no evidence", and it is reported as such rather
    than as a defect.
    """
    seen = observed_shapes()
    if not seen:
        rep.review(
            "empty-shape", "testdata",
            "no committed fixtures found, so nothing here can be checked "
            "against recorded evidence; record a journey with SH_RECORD_DIR",
        )
        return

    # Map each custom-decoded field to the wire key it reads.
    for path in go_files("pkg", tests=False):
        src = read(path)
        pattern = r'(\w+)\s+(?:\[\])?shtypes\.(Maybe\w+)[^`]*`json:"([^",]+)'
        for m in re.finditer(pattern, src):
            field, typ, key = m.group(1), m.group(2), m.group(3)
            line = src[: m.start()].count("\n") + 1
            observed = seen.get(key, set())
            if not observed:
                continue
            if "empty-list" in observed or "empty-object" in observed:
                rep.confirmed(
                    "empty-shape", f"{path}:{line}",
                    f"{field} ({typ}) reads the {key} key, which a committed fixture "
                    f"shows arriving empty "
                    f"({', '.join(sorted(observed))}); confirm the decoder tolerates it",
                )


# --------------------------------------------------------------------
# 3. CHANGELOG claims
# --------------------------------------------------------------------

def check_changelog_claims(rep: Report, base: str) -> None:
    """Symbols the CHANGELOG names that are not in the tree.

    Found this way: an entry for api.SetTransport on a branch that did
    not contain it, and — missed, because the check was a spot-check
    rather than exhaustive — two entries for cloud model changes that
    shipped to main describing code that was not there.
    """
    if not os.path.exists("CHANGELOG.md"):
        return
    text = read("CHANGELOG.md")
    m = re.search(r"^## \[Unreleased\]$(.*?)^## \[", text, re.S | re.M)
    if not m:
        return
    section = m.group(1)

    tree = ""
    for path in go_files("pkg", "examples", "internal", tests=None):
        tree += read(path)

    seen: set[str] = set()
    for sym in re.findall(r"`([A-Za-z_][\w./]*(?:\.[A-Z]\w+)+)`", section):
        leaf = sym.split(".")[-1]
        if leaf in seen or not leaf[:1].isupper():
            continue
        seen.add(leaf)
        if not re.search(r"\b" + re.escape(leaf) + r"\b", tree):
            rep.confirmed(
                "changelog-claim", "CHANGELOG.md",
                f"names `{sym}` but no such symbol is in the tree",
            )

    # Duplicated bullets, which is what appending per review round produces.
    bullets = re.findall(r"^- (.+)$", section, re.M)
    for b in set(bullets):
        if bullets.count(b) > 1:
            rep.confirmed(
                "changelog-claim", "CHANGELOG.md",
                f"duplicate entry: {b[:70]}...",
            )


# --------------------------------------------------------------------
# 4. Historical claims in documentation
# --------------------------------------------------------------------

PAST_FAILURE = re.compile(
    r"(had never (?:worked|decoded|returned)|"
    r"never (?:once )?(?:worked|decoded|shipped)|"
    r"did not work through this SDK|was broken|could not work|"
    r"every \w+ (?:through this|call) \w* ?failed)",
    re.I,
)


def check_historical_claims(rep: Report, base: str) -> None:
    """Comments asserting how the code used to behave.

    A comment describing a past bug is a claim about history, and
    history has a source of truth. Public godoc claiming shipped code
    was broken is a reputational statement that cannot be taken back
    once indexed.

    Found this way: Create's godoc asserting "every provision through
    this method failed", which was false for every released version —
    the bug existed only in an intermediate commit on the branch.
    """
    changed = subprocess.run(
        ["git", "diff", "--name-only", f"{base}...HEAD"],
        capture_output=True, text=True, check=False,
    ).stdout.split()

    for path in changed:
        if not path.endswith(".go") or not os.path.exists(path):
            continue
        for n, line in enumerate(read(path).split("\n"), 1):
            s = line.strip()
            if not s.startswith("//"):
                continue
            if PAST_FAILURE.search(s):
                rep.review(
                    "historical-claim", f"{path}:{n}",
                    f"asserts past behaviour — verify against `git show {base}:{path}` "
                    f"before shipping: {s[:80]}",
                )


# --------------------------------------------------------------------
# 5. Tests that cannot fail
# --------------------------------------------------------------------

def check_unasserted_options(rep: Report, base: str) -> None:
    """Fields set in a test's request but never asserted on the wire.

    A test that sets an option and does not check it reaches the API
    would pass identically if the code stopped sending it. That reads
    as coverage in a diff and provides none, which is worse than the
    original gap because the next reader believes it is pinned.

    Found this way: SortBy and SortDir added to a ListImages call with
    no matching assertion in the handler.
    """
    for path in go_files("pkg", "examples", tests=True):
        src = read(path)
        for m in re.finditer(r"\nfunc (Test\w+)\(", src):
            nxt = src.find("\nfunc ", m.start() + 1)
            body = src[m.start(): nxt if nxt > 0 else len(src)]
            if "httptest.NewServer" not in body:
                continue
            for fm in re.finditer(r"^\s*([A-Z]\w+):\s+\"([^\"]+)\",", body, re.M):
                field_name, value = fm.group(1), fm.group(2)
                snake = re.sub(r"(?<!^)(?=[A-Z])", "_", field_name).lower()
                # Present under any spelling, or the value itself is
                # asserted somewhere, means it is pinned.
                if snake in body or snake.replace("_", "") in body.lower():
                    continue
                if f'"{value}"' in body.replace(fm.group(0), ""):
                    continue
                # Only meaningful for request options, not arbitrary
                # struct literals in a fixture.
                if "Options{" not in body and "Request{" not in body:
                    continue
                rep.review(
                    "unasserted-option", f"{path}:{m.group(1)}",
                    f"sets {field_name} but nothing in the handler looks for "
                    f'"{snake}"; the test would pass if it stopped being sent',
                )


# Suffixes matter: CreateZone, AddRecord and DeleteZone are mutations,
# and requiring an exact method name missed every one of them — so the
# DNS journey, which creates and destroys zones, was never flagged.
MUTATES = re.compile(r"\.(Add|Create|Update|Delete|Restore|Remove|Swap|Set)\w*\(")
OUT_OF_BAND = ("sshRun", "tcpReachable", "waitReachability", "assertBlocked",
               "assertReachable", "net.Dial", "exists(")


def check_control_plane_only(rep: Report) -> None:
    """Journey steps that change something and only ask the API about it.

    Asking the API that performed an action whether it performed it
    proves a record changed, not that anything happened. The control
    plane can report a firewall rule applied while the packet filter
    does nothing, and a restore job can report Completed in ten seconds
    without reverting a disk. Neither failure is visible from the
    control plane at all.

    This is how the gaps in this repository have actually been found:
    run the journey, record what came back, and check the result
    somewhere other than the thing that produced it.

    Reported as REVIEW rather than CONFIRMED because some steps
    legitimately have nowhere else to look — a listing has no
    out-of-band form. The question is whether this one does.
    """
    for path in go_files("examples", tests=False):
        src = read(path)
        # Judged per file, not per function: a step whose mutations sit
        # behind local helpers has no client call in its own body, and
        # checking the body alone missed the address swap entirely —
        # the step with the largest blast radius in the repository.
        if not MUTATES.search(src):
            continue
        if any(tok in src for tok in OUT_OF_BAND):
            continue
        # step[A-Z] only: steps() and stepf() are not journey steps.
        for m in re.finditer(r"\nfunc (step[A-Z]\w+)\(", src):
            rep.review(
                "control-plane-only", f"{path}:{m.group(1)}",
                "mutates and then verifies only through the same API; consider whether "
                "the result can be observed out of band — a socket, a shell in the "
                "guest, a resolver — the way secgroup and snapshot do",
            )


# --------------------------------------------------------------------
# 6. Destructive actions reached through a fallback
# --------------------------------------------------------------------

DESTRUCTIVE = re.compile(r"\.(Delete|Restore|Destroy|Remove)\w*\(")


def check_destructive_fallbacks(rep: Report) -> None:
    """Destructive calls whose target can come from the environment.

    Naming a resource for a read-only step must never be read as
    consent to destroy it.

    Found this way: the delete step fell back to SH_SERVER_A/SH_SERVER_B
    when nothing had been provisioned, and the journey runs cleanup even
    after a failed tour — so a run that failed before provisioning would
    force-delete two servers the operator already owned.
    """
    for path in go_files("examples", tests=False):
        src = read(path)
        if not DESTRUCTIVE.search(src):
            continue
        for m in re.finditer(r"os\.Getenv\(\"(\w+)\"\)", src):
            var = m.group(1)
            if "DELETE" in var or "DESTROY" in var:
                continue  # an explicit opt-in, which is the fix
            line = src[: m.start()].count("\n") + 1
            window = src[max(0, m.start() - 1500): m.start() + 1500]
            if DESTRUCTIVE.search(window):
                rep.review(
                    "destructive-fallback", f"{path}:{line}",
                    f"{var} is read within reach of a destructive call; confirm it cannot "
                    f"become the target of one without an opt-in named for that purpose",
                )


# --------------------------------------------------------------------
# 7. Documented but not read
# --------------------------------------------------------------------

def check_documented_env(rep: Report) -> None:
    """Environment variables named in docs that the program never reads.

    Silently ignoring one is the worst available outcome: someone
    pointing a journey at a sandbox gets production.

    Found this way: SH_BASE_URL documented in two places and read
    nowhere; SH_SSH_KEY_FILE documented as a fallback and unreachable
    from either step that needed it — twice, in two review rounds.
    """
    for root in sorted(
        {os.path.dirname(p) for p in go_files("examples", tests=False)}
    ):
        code = "".join(read(p) for p in go_files(root, tests=False))
        docs = ""
        for name in ("README.md",):
            p = os.path.join(root, name)
            if os.path.exists(p):
                docs += read(p)
        docs += code  # package doc comments live in the code

        documented = set(re.findall(r"\b(SH_[A-Z0-9_]+)\b", docs))
        read_vars = set(re.findall(r"os\.Getenv\(\"(SH_[A-Z0-9_]+)\"\)", code))
        read_vars |= set(re.findall(r"envOr\(\"(SH_[A-Z0-9_]+)\"", code))

        for var in sorted(documented - read_vars):
            rep.confirmed(
                "documented-not-read", root,
                f"{var} is documented but the program never reads it",
            )


# --------------------------------------------------------------------
# 9. Absolute claims in documentation
# --------------------------------------------------------------------

ABSOLUTE = re.compile(
    r"\b(in every configuration|every path|always returns|never returns|"
    r"must carry|must be|it does not return|on every|in all cases|"
    r"unconditionally|without exception)\b", re.I)


def check_absolute_claims(rep: Report, base: str) -> None:
    """Doc comments making a claim the function's own branches may break.

    A doc and the code beneath it get written from the same intention,
    in the same sitting, and neither is checked against the other. That
    is how a comment saying a function returns one type "in every
    configuration" ended up above a branch returning a different one,
    and how a guard documented as requiring a particular status ended up
    skipping the check when the status was absent.

    Absolutes are worth singling out because they are falsifiable by a
    single branch, and because a strengthened doc is often written to
    close a doc-versus-code finding — so getting it wrong reopens the
    finding while looking like the fix.

    Reported for the changed files only, and only where the enclosing
    function has several returns, which is the shape that can
    contradict one.
    """
    changed = subprocess.run(
        ["git", "diff", "--name-only", f"{base}...HEAD"],
        capture_output=True, text=True, check=False,
    ).stdout.split()

    for path in changed:
        if not path.endswith(".go") or path.endswith("_test.go"):
            continue
        if not os.path.exists(path):
            continue
        src = read(path)
        lines = src.split("\n")
        for n, line in enumerate(lines):
            if not line.strip().startswith("//"):
                continue
            m = ABSOLUTE.search(line)
            if not m:
                continue
            # Find the declaration this comment belongs to, and count
            # its returns. One return cannot contradict an absolute.
            j = n
            while j < len(lines) and not lines[j].startswith("func "):
                if lines[j].strip() and not lines[j].strip().startswith("//"):
                    break
                j += 1
            if j >= len(lines) or not lines[j].startswith("func "):
                continue
            end = j + 1
            while end < len(lines) and not lines[end].startswith("}"):
                end += 1
            body = "\n".join(lines[j:end])
            returns = len(re.findall(r"\breturn\b", body))
            if returns < 2:
                continue
            rep.review(
                "absolute-claim", f"{path}:{n + 1}",
                f'claims "{m.group(1)}" above a function with {returns} return paths; '
                f"check the claim against each one rather than against what it was meant to do",
            )


# --------------------------------------------------------------------
# 10. Sibling sites for a changed pattern
# --------------------------------------------------------------------

def check_sibling_sites(rep: Report, base: str) -> None:
    """Methods of the same name this branch changed in one place only.

    The most expensive habit in this repository's review history is
    fixing the site a finding names rather than the class it belongs to.
    A parameter bug fixed in cloud/db.Get sat untouched in
    cloud/db/user.Get. A guard anchored in isThrottled was left loose in
    IsRateLimited, three lines below it in the same file. A doc
    corrected on a constant was left wrong on the test describing it.

    Matching on method name rather than file name, because that is the
    axis the repeats actually fell along: the same endpoint wrapper
    implemented once per namespace.

    It cannot know whether a sibling needs the same change. It can stop
    anyone deciding that by not looking.
    """
    changed = [
        p for p in subprocess.run(
            ["git", "diff", "--name-only", f"{base}...HEAD"],
            capture_output=True, text=True, check=False,
        ).stdout.split()
        if p.endswith(".go") and not p.endswith("_test.go") and os.path.exists(p)
    ]
    if not changed:
        return

    # Methods whose body this branch touched.
    touched: dict[str, str] = {}
    for path in changed:
        diff = subprocess.run(
            ["git", "diff", f"{base}...HEAD", "--", path],
            capture_output=True, text=True, check=False,
        ).stdout
        for m in re.finditer(r"^[+-].*func \([^)]*\) ([A-Z]\w+)\(", diff, re.M):
            touched[m.group(1)] = path
        # A hunk header names the enclosing declaration.
        for m in re.finditer(r"^@@.*@@ func \([^)]*\) ([A-Z]\w+)\(", diff, re.M):
            touched[m.group(1)] = path

    for name, path in sorted(touched.items()):
        siblings = []
        for other in go_files("pkg", tests=False):
            if other in changed:
                continue
            if re.search(r"func \([^)]*\) " + re.escape(name) + r"\(", read(other)):
                siblings.append(other)
        if not siblings:
            continue
        rep.review(
            "sibling-sites", f"{path}:{name}",
            f"{name} is also implemented in {len(siblings)} unchanged file(s): "
            f"{', '.join(siblings[:4])}{' ...' if len(siblings) > 4 else ''} — "
            f"confirm the same change is not needed there, or say why not",
        )


# --------------------------------------------------------------------

CHECKS = [
    ("wire contracts", check_wire_contracts, False),
    ("empty-collection tolerance", check_empty_shapes, False),
    ("CHANGELOG claims", check_changelog_claims, True),
    ("historical claims", check_historical_claims, True),
    ("unasserted options", check_unasserted_options, True),
    ("control-plane-only verification", check_control_plane_only, False),
    ("destructive fallbacks", check_destructive_fallbacks, False),
    ("documented but not read", check_documented_env, False),
    ("absolute claims", check_absolute_claims, True),
    ("sibling sites", check_sibling_sites, True),
]


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--base", default="origin/main")
    ap.add_argument("--json", action="store_true")
    args = ap.parse_args()

    rep = Report()
    for _, fn, needs_base in CHECKS:
        if needs_base:
            fn(rep, args.base)
        else:
            fn(rep)

    if args.json:
        print(json.dumps([f.__dict__ for f in rep.findings], indent=2))
    else:
        confirmed = [f for f in rep.findings if f.kind == "CONFIRMED"]
        review = [f for f in rep.findings if f.kind == "REVIEW"]

        if confirmed:
            print("CONFIRMED — the script proved these:\n")
            for f in confirmed:
                print(f"  [{f.check}] {f.where}")
                print(f"      {f.what}\n")
        if review:
            print("REVIEW — candidates a human has to judge:\n")
            for f in review:
                print(f"  [{f.check}] {f.where}")
                print(f"      {f.what}\n")
        if not rep.findings:
            print("No findings.")
        print(f"{len(confirmed)} confirmed, {len(review)} to review.")

    return 1 if any(f.kind == "CONFIRMED" for f in rep.findings) else 0


if __name__ == "__main__":
    sys.exit(main())
