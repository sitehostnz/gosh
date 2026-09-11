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

            # Two shapes, and both have to be read. A keys variable:
            #
            #     keys := []string{"client_id", "name"}
            #
            # and the list written inline as an argument:
            #
            #     net.Encode(values, []string{"client_id", "code"})
            #
            # Only the first was matched, so twelve of the ninety-five
            # call sites were skipped in silence — ten of them in
            # pkg/api/dns/template, where the inline form is the norm.
            # Dropping a key from an inline list produced no finding at
            # all.
            key_text = ""
            km = re.search(r"keys\s*:?=\s*\[\]string\{(.*?)\n\t\}", body, re.S)
            if not km:
                km = re.search(r"keys\s*:?=\s*\[\]string\{([^}]*)\}", body, re.S)
            if km:
                key_text = km.group(1)
            for inline in re.finditer(
                    r"net\.Encode\([^,]+,\s*\[\]string\{([^}]*)\}", body):
                key_text += "," + inline.group(1)
            if not key_text:
                continue
            keys = set(re.findall(r'"([^"]+)"', key_text))

            # dynamic means "this function's parameter names cannot be
            # read from the source", and it is what makes the check say
            # so rather than stay quiet.
            #
            # A quoted literal next to a "+" is a FRAGMENT, not a key.
            # Extracting it and treating the extraction as success is
            # how two call sites came to be reported clean while none of
            # their real parameters had been seen:
            #
            #   keys = append(keys, "environments["+request.Name+".env]")
            #   values.Add("environments["+request.Name+".env]", ...)
            #
            # yielded "environments[" and ".env]" on both sides, so they
            # agreed — agreement between two identically mangled
            # fragments, not verification. And because literals *were*
            # found, dynamic stayed false, so nothing said the function
            # was unjudgeable.
            #
            #   keys = append(keys, prefix+"[enabled]")
            #   values.Add(prefix+"[enabled]", ...)
            #
            # was worse: the append yielded "[enabled]", again clearing
            # dynamic, while the Add was invisible to the extractor
            # because its literal is not the first token. Ten
            # parameters, none seen, reported clean.
            dynamic = False
            for am in re.finditer(r"keys\s*=\s*append\(keys,\s*(.*?)\)", body):
                arg = am.group(1)
                if "+" in arg:
                    dynamic = True
                    continue
                found = re.findall(r'"([^"]+)"', arg)
                keys |= set(found)
                if not found:
                    dynamic = True

            # The same rule on the sending side: a key built by
            # concatenation cannot be compared against the keys list,
            # whichever end the literal sits at.
            if re.search(r"\w+\.(?:Add|Set)\(\s*(?:\"[^\"]*\"\s*\+|\w+\s*\+)", body):
                dynamic = True

            # Header.Set is not a query parameter. Matching it produced
            # three false CONFIRMED findings on the first run, which is
            # the failure this tool can least afford: a check that cries
            # wolf gets ignored, exactly like one that cannot fail.
            #
            # The receiver is captured and compared, rather than excluded
            # with a lookbehind. Two earlier attempts got this wrong:
            #
            #   - (?<!Header)\b(?<!Header\.) excluded nothing at all,
            #     because \w+ binds to "Header" itself, so the
            #     lookbehinds were evaluated against what precedes
            #     "Header" — "req." — rather than against what precedes
            #     ".Set". The pattern behaved identically with them
            #     removed.
            #   - Filtering out any key name that also appeared in a
            #     Header.Add/Set call anywhere in the body did work, but
            #     globally: a genuine query parameter sharing a name
            #     with a header set elsewhere in the same function was
            #     suppressed too, so a real absent-key bug on that name
            #     would go unreported. "type" is both a parameter name
            #     in this repository and a plausible header, so the
            #     collision is not hypothetical.
            # Only whole literal keys: a match followed by "+" is a
            # fragment and is skipped, having already set dynamic above.
            adds = [
                m.group(2)
                for m in re.finditer(
                    r"\b(\w+)\.(?:Add|Set)\(\s*\"([^\"]+)\"\s*(?P<cat>\+?)", body)
                if m.group(1) != "Header" and not m.group("cat")
            ]
            where = f"{path}:{fn}()"

            if dynamic:
                # Silence and "checked and clean" must not look the
                # same. One append with a non-literal argument
                # suppressed every finding for the rest of the
                # function, including literal values.Add calls the
                # check could still have judged — and said nothing
                # about having given up.
                rep.review(
                    "wire-contract", where,
                    "builds its keys list dynamically, so the absent-key check "
                    "cannot judge this function; verify the parameters by hand",
                )

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
    # Keyed by (package directory, wire key). Keying on the bare key
    # name meant a "disk" key in one package's fixtures flagged a
    # "disk" field in an unrelated one, with no association between
    # them beyond the spelling.
    shapes: dict[tuple[str, str], set[str]] = {}

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
                shapes.setdefault((pkg, k), set()).add(t)
                walk(v)
        elif isinstance(o, list):
            for x in o:
                walk(x)

    for dirpath, _, names in os.walk("."):
        if "/testdata" not in dirpath or "/.git" in dirpath:
            continue
        # The package a fixture belongs to is its testdata directory's
        # parent, which is what scopes the observation.
        pkg = os.path.normpath(os.path.dirname(dirpath.rstrip("/")))
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
            observed = seen.get((os.path.normpath(os.path.dirname(path)), key), set())
            if not observed:
                continue
            if "empty-list" in observed or "empty-object" in observed:
                # REVIEW, not CONFIRMED. This establishes one half of
                # the claim — that a fixture shows the key arriving
                # empty — and never looks at the decoder. A
                # shtypes.Maybe* type written to accept [], with tests
                # proving it, was reported as a defect that blocked the
                # push. The message was always a question addressed to
                # a human, which is the definition of a REVIEW line.
                rep.review(
                    "empty-shape", f"{path}:{line}",
                    f"{field} ({typ}) reads the {key} key, which a committed fixture "
                    f"in this package shows arriving empty "
                    f"({', '.join(sorted(observed))}); confirm the decoder tolerates it",
                )


# --------------------------------------------------------------------
# 3. CHANGELOG claims
# --------------------------------------------------------------------

def check_changelog_missing(rep: Report, base: str) -> None:
    """A branch that changes shipped code and adds no CHANGELOG entry.

    The repository's rule is that every merged PR carries its own
    entry, and a missing one is the most frequent review comment it
    gets. The neighbouring check validates the *content* of entries
    against the tree and had nothing to say about their absence — so a
    branch with no entry at all passed every mechanical check, which is
    how one reached review having run all of them.

    Docs-only and test-only branches are exempt: there is nothing for a
    consumer to be told.
    """
    changed = subprocess.run(
        ["git", "diff", "--name-only", f"{base}...HEAD"],
        capture_output=True, text=True, check=False,
    ).stdout.split()
    if not changed:
        return
    if "CHANGELOG.md" in changed:
        return

    # What counts as shipped code: the SDK, and the examples, which this
    # repository treats as documentation that fails loudly.
    notable = [
        p for p in changed
        if (p.startswith("pkg/") or p.startswith("examples/"))
        and p.endswith(".go") and not p.endswith("_test.go")
    ]
    if not notable:
        return

    rep.confirmed(
        "changelog-missing", "CHANGELOG.md",
        f"{len(notable)} non-test Go file(s) changed and no CHANGELOG entry was "
        f"added; every merged PR carries one, and a missing entry is this "
        f"repository's most common review comment",
    )


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

            # Search the handler closure only, and bound it at the
            # closing "}))".
            #
            # The setter lives in the function body, so any test of "is
            # this name mentioned in the body" is answered by the very
            # line that prompted the question. The first version of
            # this check did that and could not fire for any input: for
            # a CamelCase field F, snake_of(F).replace("_", "") is
            # F.lower(), always present in body.lower() because the
            # match came from the body.
            #
            # Slicing from httptest.NewServer to the end of the
            # function is not enough either — the request literal
            # usually sits after the handler, so the setter is still in
            # scope. That version also could not fire. Hence the bound.
            start = body.find("httptest.NewServer")
            end = body.find("}))", start)
            handler = body[start: end if end > 0 else len(body)]

            for fm in re.finditer(r"^\s*([A-Z]\w+):\s+\"([^\"]+)\",", body, re.M):
                field_name, value = fm.group(1), fm.group(2)
                snake = re.sub(r"(?<!^)(?=[A-Z])", "_", field_name).lower()
                # Pinned if the handler looks for the wire name or the
                # value. Deliberately not the Go field name: a handler
                # reads query and form parameters, so the wire spelling
                # is what an assertion would name, and accepting the Go
                # spelling reopened the hole this check had.
                if snake in handler or f'"{value}"' in handler:
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
# Ways a step can observe a result without asking the API that produced
# it. HTTP belongs here: fetching a page a deploy was supposed to serve
# is a socket to the thing itself, and a journey that verified that way
# was flagged as control-plane-only.
OUT_OF_BAND = ("sshRun", "sshRunAs", "tcpReachable", "waitReachability",
               "assertBlocked", "assertReachable", "net.Dial", "exists(",
               "httpGetStatus", "waitHTTPServed", "curlFromHost",
               "waitCurlFromHost", "http.Get", "dig ", "net.Lookup")


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


# Sentences that disclaim a variable rather than promise it.
NEGATION = re.compile(
    r"\b(not read|never read|not honou?red|ignored|deliberately|"
    r"elsewhere|no longer|unused)\b", re.I)


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

        # Documentation means the README and comment lines, not the
        # whole source. Including the code meant any mention anywhere
        # counted as a promise — including a comment written to say a
        # variable is deliberately NOT honoured, which is the one case
        # where the code is right and the documentation is doing its
        # job.
        docs = ""
        readme = os.path.join(root, "README.md")
        if os.path.exists(readme):
            docs += read(readme)
        for line in code.splitlines():
            stripped = line.strip()
            if not stripped.startswith("//"):
                continue
            # A sentence that disclaims the variable is not a promise
            # to read it.
            if NEGATION.search(stripped):
                continue
            docs += stripped + "\n"

        documented = set(re.findall(r"\b(SH_[A-Z0-9_]+)\b", docs))

        # Any occurrence inside an environment lookup counts, literal
        # or not, and so does binding the name to a constant — both are
        # reads, and requiring the literal call shape missed them.
        read_vars: set[str] = set()
        for call in re.finditer(r"(?:os\.Getenv|os\.LookupEnv|envOr)\(([^)]*)\)", code):
            read_vars |= set(re.findall(r"(SH_[A-Z0-9_]+)", call.group(1)))
        for assign in re.finditer(r"(?:const|var)\s+\w+\s*=\s*\"(SH_[A-Z0-9_]+)\"", code):
            read_vars.add(assign.group(1))
        # A name held in a const and passed by identifier: if the
        # program mentions the token in any assignment at all, the
        # script cannot prove it is unread.
        for assign in re.finditer(r"=\s*\"(SH_[A-Z0-9_]+)\"", code):
            read_vars.add(assign.group(1))

        for var in sorted(documented - read_vars):
            # REVIEW rather than CONFIRMED. The true finding is
            # valuable — a documented variable that is silently ignored
            # sends someone pointing at a sandbox to production — but
            # the script reads prose to decide, and prose is not
            # something it can prove things about.
            rep.review(
                "documented-not-read", root,
                f"{var} appears in documentation but no environment read of it "
                f"is visible; confirm it is honoured or that the mention says it is not",
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
            # Quoted text is evidence, not a claim. A doc comment that
            # reproduces an API's own error — "the server must be on" —
            # is recording what the platform said, and the phrase
            # belongs to the platform rather than to the function
            # underneath. Flagging those meant the more precisely a
            # comment quoted its source, the more likely it was to be
            # reported.
            unquoted = re.sub(r'"[^"]*"', "", line)
            # A quotation that wraps onto the next comment line leaves
            # an unbalanced quote behind, and the phrase often sits in
            # the tail — which is how "the server must be / on." was
            # still reported after quoted spans were excluded. Drop
            # from the dangling quote to the end of the line.
            if unquoted.count('"') % 2 == 1:
                unquoted = unquoted[: unquoted.index('"')]
            m = ABSOLUTE.search(unquoted)
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


# --------------------------------------------------------------------
# 11. Rejection fixtures that cannot prove a rejection
# --------------------------------------------------------------------

def check_rejection_fixtures(rep: Report) -> None:
    """Fixtures cited as rejections that carry no status field.

    This API has two rejection shapes: a 200 carrying
    `"status": false`, and a 400 whose body carries only a `msg`. A
    fixture of the second kind decodes `status` to Go's zero value,
    which is also false — so a test asserting "this was rejected"
    passes on the zero value rather than on anything recorded, and
    would pass identically against a fixture recording nothing at all.

    That is a question, not a defect, and this check is a REVIEW line
    for a reason worth stating: the first version reported CONFIRMED,
    and the first fixture it flagged turned out to be a faithful
    recording of a real 400 with no status field. Both halves have to
    hold — the fixture must be unable to prove the rejection *and* the
    response must not really be that shape — and the second half is
    only answerable by looking.

    Found this way: a fixture named for an unknown zone contained only
    {"msg": "Please specify a valid domain name."}. The test passed, and
    the recorded rejection turned out to be the TLD validator refusing
    to parse the name rather than the not-found it was cited for. Two
    doc comments and three tests encoded the wrong conclusion.
    """
    for dirpath, _, names in os.walk("."):
        if "/testdata" not in dirpath or "/.git" in dirpath:
            continue
        for n in sorted(names):
            if not n.endswith(".json"):
                continue
            # Only fixtures whose name says they hold a rejection.
            if not re.search(r"(reject|error|unknown|invalid|missing|denied|notfound|not-found)", n, re.I):
                continue
            path = os.path.join(dirpath, n)
            try:
                with open(path, encoding="utf-8") as fh:
                    doc = json.load(fh)
            except Exception:
                continue
            if not isinstance(doc, dict):
                continue
            if "status" not in doc:
                rep.review(
                    "rejection-fixture", path,
                    "is named as a rejection but carries no \"status\" field, so a "
                    "test asserting the rejection passes on Go's zero value; confirm "
                    "the API really answers that way and that the message is the "
                    "rejection the test cites, not a different one",
                )
            elif doc.get("status") is True:
                rep.review(
                    "rejection-fixture", path,
                    "is named as a rejection but records status:true; confirm the "
                    "name describes what was actually observed",
                )


# --------------------------------------------------------------------
# 12. Assertions removed from tests
# --------------------------------------------------------------------

def check_dropped_assertions(rep: Report, base: str) -> None:
    """Test files that lose more assertions than they gain.

    Replacing a hand-written test with a fixture-backed one is often
    right, and it is also how coverage disappears without anything
    turning red: the two kinds catch disjoint bugs, so swapping one for
    the other silently drops whatever only the first could see.

    Found this way: a rewrite replaced two tests with three
    fixture-backed ones. The fixture tests could not detect a swapped
    json tag or a wrong endpoint path — both were mutated and the suite
    still reported ok — because a scrubbed fixture's values are all
    placeholders and the serving helper ignores the path.
    """
    diff = subprocess.run(
        ["git", "diff", "-U0", f"{base}...HEAD", "--", "*_test.go"],
        capture_output=True, text=True, check=False,
    ).stdout

    assertion = re.compile(r"\b(t\.Errorf|t\.Error|t\.Fatalf|t\.Fatal)\b")
    per_file: dict[str, list[int]] = {}
    current = None
    for line in diff.splitlines():
        if line.startswith("+++ b/"):
            current = line[6:]
            per_file.setdefault(current, [0, 0])
            continue
        if current is None:
            continue
        if line.startswith("-") and assertion.search(line):
            per_file[current][0] += 1
        elif line.startswith("+") and assertion.search(line):
            per_file[current][1] += 1

    for path, (removed, added) in sorted(per_file.items()):
        if removed > added:
            rep.review(
                "dropped-assertions", path,
                f"removes {removed} assertion(s) and adds {added}; confirm nothing "
                f"the old ones caught is now unchecked — a fixture test and a "
                f"literal-comparison test catch disjoint bugs",
            )



# --------------------------------------------------------------------
# 13. Build artefacts in the diff
# --------------------------------------------------------------------

def check_binary_artefacts(rep: Report, base: str) -> None:
    """Compiled or generated files added by the branch.

    Found this way: a __pycache__/*.pyc committed alongside the checker
    it belongs to. The repository's rule is to stage by explicit path
    rather than `git add -A`, which guards against stale artefacts at
    the repo root — and does nothing about one that materialises inside
    the directory you meant to stage. Running the checker writes its
    bytecode next to its source, so `git add .claude/` picks it up
    without anyone deciding to.
    """
    changed = subprocess.run(
        ["git", "diff", "--name-only", "--diff-filter=A", f"{base}...HEAD"],
        capture_output=True, text=True, check=False,
    ).stdout.split()

    suspect = (
        ".pyc", ".pyo", ".class", ".o", ".a", ".so", ".dylib", ".exe",
        ".test", ".out", ".DS_Store",
    )
    for path in changed:
        base_name = os.path.basename(path)
        if "__pycache__" in path or path.endswith(suspect) or base_name == ".DS_Store":
            rep.confirmed(
                "build-artefact", path,
                "is a build artefact added by this branch; it should be ignored "
                "rather than committed",
            )
            continue
        # A tracked file with no extension that is not text is usually a
        # compiled binary left by `go build ./some/pkg`.
        if "." not in base_name and os.path.exists(path):
            try:
                with open(path, "rb") as fh:
                    if b"\0" in fh.read(1024):
                        rep.confirmed(
                            "build-artefact", path,
                            "looks like a compiled binary added by this branch",
                        )
            except OSError:
                pass


# --------------------------------------------------------------------
# 14. Claims nothing in the tree can support
# --------------------------------------------------------------------

COST_CLAIM = re.compile(
    r"\b(is free|costs? nothing|no charge|free of charge|does not cost|"
    r"costs money|is billable|will be billed)\b", re.I)


def check_cost_claims(rep: Report, base: str) -> None:
    """Billing assertions in documentation.

    A reader cannot check one with a probe, and nothing in a fixture can
    evidence it, so it is the class of claim most likely to be repeated
    from memory and least likely to be caught.

    Found this way: "which is free", "It costs nothing" and "only one of
    them costs money" across three files of one example, with no
    citation behind any of them — in a branch whose stated purpose was
    replacing assumed behaviour with observed behaviour.
    """
    changed = subprocess.run(
        ["git", "diff", "--name-only", f"{base}...HEAD"],
        capture_output=True, text=True, check=False,
    ).stdout.split()

    for path in changed:
        if not (path.endswith(".go") or path.endswith(".md")):
            continue
        if not os.path.exists(path):
            continue
        for n, line in enumerate(read(path).split("\n"), 1):
            stripped = line.strip()
            if not (stripped.startswith("//") or path.endswith(".md")):
                continue
            if COST_CLAIM.search(stripped):
                rep.review(
                    "cost-claim", f"{path}:{n}",
                    "asserts what something costs; confirm it is evidenced, or "
                    "narrow it to what is known (which operation is billable "
                    "rather than what it costs)",
                )


CHECKS = [
    ("wire contracts", check_wire_contracts, False),
    ("empty-collection tolerance", check_empty_shapes, False),
    ("CHANGELOG claims", check_changelog_claims, True),
    ("CHANGELOG missing", check_changelog_missing, True),
    ("historical claims", check_historical_claims, True),
    ("unasserted options", check_unasserted_options, True),
    ("control-plane-only verification", check_control_plane_only, False),
    ("destructive fallbacks", check_destructive_fallbacks, False),
    ("documented but not read", check_documented_env, False),
    ("absolute claims", check_absolute_claims, True),
    ("sibling sites", check_sibling_sites, True),
    ("rejection fixtures", check_rejection_fixtures, False),
    ("dropped assertions", check_dropped_assertions, True),
    ("build artefacts", check_binary_artefacts, True),
    ("cost claims", check_cost_claims, True),
]


SELFTEST_DIR = os.path.join(
    os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "selftest")


def selftest() -> int:
    """Run the checks against the fixtures and assert they still fire.

    A check whose failure mode is silence needs a test more than most
    code does. Two checks here shipped unable to fire for any input,
    and neither was visible in the output: a check that cannot fire
    reports exactly what a clean tree reports.

    The fixtures live under selftest/ and are Go files by extension
    only — they are never compiled, and the directory is outside the
    module's package layout so `go build ./...` does not see them.
    """
    expected_path = os.path.join(SELFTEST_DIR, "expected.json")
    with open(expected_path, encoding="utf-8") as fh:
        expected = json.load(fh)

    # The checks walk "pkg" and "examples" relative to the working
    # directory, so the fixtures live under those names inside
    # selftest/ and the directory is swapped rather than the signatures
    # changed.
    cwd = os.getcwd()
    rep = Report()
    try:
        os.chdir(SELFTEST_DIR)
        for _, fn, needs_base in CHECKS:
            # Diff-based checks are given HEAD and simply find nothing
            # here, which is fine — the fixtures assert on the checks
            # that walk the tree. Calling them anyway keeps this honest
            # about which checks the self-test does and does not cover.
            fn(rep, "HEAD") if needs_base else fn(rep)
    finally:
        os.chdir(cwd)

    got: dict[str, int] = {}
    for f in rep.findings:
        got[f.check] = got.get(f.check, 0) + 1

    failed = False
    for check, want in expected.items():
        if got.get(check, 0) < want:
            print(f"SELFTEST FAIL: {check} produced {got.get(check, 0)} "
                  f"finding(s) on the fixtures, expected at least {want}")
            failed = True
        else:
            print(f"  ok  {check}: {got.get(check, 0)} finding(s)")

    if failed:
        print("\nA check stopped firing on the fixture written for it. "
              "That is not a clean tree — it is a check that no longer works.")
        return 1
    print("\nselftest: every checked fixture still fires")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--base", default="origin/main")
    ap.add_argument("--json", action="store_true")
    ap.add_argument("--selftest", action="store_true",
                    help="run the checks against selftest/ and assert they fire")
    args = ap.parse_args()

    if args.selftest:
        return selftest()

    # A checker that cannot see the diff has to say so rather than
    # pass. Three checks resolve the base through git and discarded the
    # exit status, so a base that does not exist — the default in CI,
    # where a depth-1 checkout creates no refs/remotes/origin/* — made
    # them examine an empty file list and report nothing. That is
    # indistinguishable from a clean branch.
    probe = subprocess.run(["git", "rev-parse", "--verify", args.base],
                           capture_output=True, check=False)
    if probe.returncode != 0:
        print(f"premr: base ref {args.base!r} does not resolve in this "
              f"working copy, so the diff-based checks cannot run.",
              file=sys.stderr)
        print("premr: in CI, check out with fetch-depth: 0 so the base "
              "branch is present.", file=sys.stderr)
        return 2

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
