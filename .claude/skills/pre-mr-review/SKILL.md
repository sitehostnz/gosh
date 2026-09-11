---
name: pre-mr-review
description: Run before opening or updating a gosh MR. Applies the review checks this repository has learned to expect — wire contracts, claims against evidence, tests that cannot fail, destructive-action safety, and completeness of a change.
---

# Pre-MR review

Run this before opening an MR, and again before pushing changes in
response to review.

The checks here are not general code review — `make lint` covers style.
They exist because 33 findings across two review rounds on #64 fell
into a small number of repeating classes, and almost none of them were
logic bugs. Nothing was miscomputed. Every one was **an artefact
claiming something reality does not support**: the code says X and the
comment says Y; the CHANGELOG describes a type that is not in the tree;
a test sets a field it never checks.

## 1. Run the mechanical checks

```
python3 .claude/skills/pre-mr-review/scripts/premr.py --base origin/main
```

Takes under a second. Output is in two parts:

- **CONFIRMED** — the script proved it. Fix before pushing. Non-zero
  exit.
- **REVIEW** — a candidate the script cannot judge. Not a defect: a
  question. Answer each one, and say why when dismissing.

The line between them is what the script can *prove*, not how likely
it is to be right. A check that establishes half a claim and asks a
human about the other half belongs in REVIEW however good its hit rate
— two shipped as CONFIRMED and both blocked pushes over correct code.

If a CONFIRMED finding is wrong, fix the *script*, not the code. A
check that cries wolf gets ignored, which is the same disease as a
check that cannot fail. The first version of this script reported three
false positives by matching `req.Header.Set` as a query parameter.

**Silence and "checked and clean" must never look the same.** This is
the failure this script is most prone to, because both produce an empty
report:

- Two checks shipped unable to fire for *any* input. `--selftest` runs
  each check against a fixture written to trip it, and fails if one
  stops firing. Run it in CI ahead of the checks themselves.
- Three checks resolved the base ref through git and discarded the exit
  status, so in CI — where a depth-1 checkout creates no
  `refs/remotes/origin/*` — they examined an empty file list and
  reported nothing. The script now refuses to run at all when the base
  does not resolve, and CI checks out with `fetch-depth: 0`.
- A check that gives up on a function it cannot analyse now says so as
  a REVIEW line rather than skipping quietly.

When adding a check, ask what it does when it cannot answer. If the
answer is "the same as when everything is fine", it needs a fixture and
a self-test more than it needs another pattern.

### A fragment is not a value

The sharpest version of that failure, because it produces confident
silence rather than no output. When a check reads a value out of source
text, it has three possible outcomes and only two are usually
considered:

1. it extracted nothing,
2. it extracted the whole value,
3. **it extracted part of the value and cannot tell.**

A quoted literal adjacent to `+` is case 3. `"environments["+name+".env]"`
yields `environments[` and `.env]`, and both the declaring side and the
sending side get mangled identically — so they agree, and agreement
between two identically broken extractions looks exactly like
verification. Worse, finding *some* literals cleared the flag that
existed to say "I cannot read this function", so nothing reported the
gap. Two call sites were reported clean having had none of their real
parameter names read; one of them has ten.

So: when a pattern extracts a literal, check what sits next to it. If
the answer is a concatenation, an index, or a format verb, the check
has a fragment and must say it cannot judge — never compare fragments
and never treat a partial read as a clean one.

### Fixtures have to be adversarial, not illustrative

A fixture built from the bug you already found only proves the check
still catches that bug. The fixtures that matter are the inputs a check
is most likely to be blind to:

- a value built by concatenation rather than written as a literal,
- the same construct written the other way round (`prefix+"[x]"` as
  well as `"x["+suffix`), because an extractor anchored to a leading
  literal cannot see the first form,
- a construct where the check finds *something*, since a partial find
  is what disables the "cannot judge" path.

Every one of those was a real blind spot here, and each was invisible
in the output. When adding a check, write the fixture that would defeat
it before writing the fixture that demonstrates it.

## 1b. Run them again against the fix

After addressing a review round, run the checks a second time with
`--base` pointing at the commit the reviewer last saw, not at `main`:

```
python3 .claude/skills/pre-mr-review/scripts/premr.py --base <last-reviewed-sha>
```

This is the highest-value single habit in the file, because the
strongest pattern in this repository's review history is not what the
original branch got wrong. It is what the *fixes* got wrong:

    #65 round 1 — 11 findings
    #65 round 2 — 10 findings, 3 of them introduced by the round-1 fixes
    #65 round 3 —  2 findings, both introduced by the round-2 fixes
    #66 round 1 — 14 findings, 2 of them the recorder leaking the
                  credentials it documented as redacted

A fix is unreviewed code written under time pressure by someone who has
just been told they were wrong, and it lands in the most sensitive part
of the file — the part a reviewer has already looked at once and will
reasonably assume is now settled. Every round above, something got
worse in a place that had just been touched:

- a redaction pass replaced a sentence's tail, left its head, and
  published "the limit is applied after the limit is enforced before
  the request reaches the handler";
- a backoff fix removed a clamp and introduced an overshoot;
- a doc strengthened to close a doc-versus-code finding was made false
  by a branch three lines away;
- adding response redaction turned an empty password into REDACTED and
  made a fixture assert the opposite of the API's real behaviour.

The last of those was caught by a test rather than a reviewer, which is
the outcome to aim for.

So treat the diff since the last round as a change needing the same
scrutiny as the branch itself — because it is one, and because it is
the change most likely to be waved through.

## 1c. Read both comment channels

A review can arrive in two places, and they are separate APIs:

```
gh api repos/<owner>/<repo>/pulls/<n>/comments    # inline, anchored to a line
gh api repos/<owner>/<repo>/issues/<n>/comments   # general, not anchored
```

Check both. A finding about a file outside the diff *cannot* be an
inline thread — GitHub has nowhere to anchor it — so it arrives as a
general comment, and those are exactly the findings about the code a
reviewer had to go looking for.

Two Majors sat unread for two days here because every completeness
check queried only the inline channel, and "all findings answered" was
reported twice on that basis. Both were about files outside the diff.

Also check `reviewDecision`: threads being answered does not clear
`CHANGES_REQUESTED`, which only the reviewer can lift.

## 1d. Search the tree, not the diff

For anything being *removed* — a disclosure, a stale claim, a renamed
term — grep the whole repository, not the changed files. A redaction
scoped to the files under review is not a redaction.

The same rate-limit policy was removed three times here. Twice the
removal was reported as complete, and each time the reason it was not
was a narrower search than the problem: the second attempt softened the
wording and left the substance, the third cleaned the four locations a
reviewer had named and left the same text, in more detail, in two files
outside the diff — including a package doc, which is the landing page
on pkg.go.dev, where indexing is one-way.

The check that finally found it was one line:

```
grep -rniE "reseller|[0-9]+ requests per second" .
```

Write the grep for the thing itself, run it across the tree, and read
every hit before claiming the removal is done.

## 2. Do the checks a script cannot

### Claims against evidence

Every factual statement in a doc comment, the CHANGELOG or the MR body
must be traceable to something you ran or can point at.

- **A comment describing past behaviour is a claim about history**, and
  history has a source of truth. Check it with
  `git show origin/main:<file>` before shipping. Public godoc asserting
  that shipped code was broken is a reputational statement that cannot
  be withdrawn once indexed — and this repository shipped exactly that
  claim, falsely, in a comment saying "every provision through this
  method failed".
- **Figures cited in comments must match committed fixtures.** A claim
  whose weight comes from being a live observation should cite the
  observation that was committed, not one from a session nobody else
  can see.
- **"Verified live" expires.** If the code changed after the run,
  re-run it or drop the claim. `ListUpgrades` was documented as
  corrected while still failing outright, because the fix was made and
  the call was never re-run.

### Completeness of the change

The most expensive class, because it costs a whole review round: a fix
applied to the sites visible in the edit rather than the sites that
exist. Four mechanisms produce it, all observed here more than once.

**The finding names a site; the fix treats that site as the scope.**
`isThrottled` was anchored and `IsRateLimited` — three lines below,
same file, same safety property — was not. A constant's doc was
corrected and the test describing it was left saying the opposite.
The `sibling-sites` check lists other implementations of a changed
method; answer each rather than skipping past.

**Read the rendered doc, not the source.** `go doc ./pkg/api Client.Do`
takes a second and shows what a consumer sees. A redaction pass here
replaced the tail of a sentence and left the head, producing "the limit
is applied after the limit is enforced before the request reaches the
handler" in the published godoc for the SDK's central method — obvious
rendered, invisible in a diff hunk. Do this for every doc comment a
branch touches.

**The test you write bounds the fix.** A table row asserting the case
you just corrected passes, which makes the fix look complete. The
backoff fix honoured the caller's value on the first attempt and still
discarded it on every later one, and the table asserted attempt 1 and
stopped.

It happened twice on the same function. The replacement table then
sampled two bases that were both *structurally immune* to the next
defect — one doubled exactly onto the ceiling, the other made the
ceiling equal to itself — so the overshoot it was written to catch
could not appear in either. The samples chosen were the two numbers
most in mind: the default, and the value from the previous finding.

Where a behaviour has a property, assert the property over a range
rather than picking examples. "Never decreases, never exceeds the
ceiling, always positive, first wait is the configured value" is what a
backoff *is*, and it does not depend on choosing the right sample —
which is the judgement that failed both times.

**A doc and the code beneath it get written from the same intention and
neither is checked against the other.** A comment saying a function
returns one type "in every configuration" sat above a branch returning
another; a guard documented as requiring a status skipped the check
when the status was absent. This is most dangerous when strengthening a
doc to close a doc-versus-code finding, because getting it wrong
reopens the finding while looking like the fix. The `absolute-claim`
check flags these; walk every return path.

**A behaviour change is not analysed for what it breaks.** Wiring a
previously-discarded context into the request turned `Do(nil, ...)`
from harmless into a panic. For any change to an exported function, ask
what the old contract permitted — `git show origin/main:<file>` — and
say so in the CHANGELOG.

- For any renamed term or corrected claim, **grep for every instance**
  and resolve each. A half-applied rename is worse than none — a reader
  who trusts a corrected site and then meets an uncorrected one cannot
  tell which is current.
- For any bug fixed in one file, **look for siblings with the same
  shape**. `cloud/db.Get` was fixed while `cloud/db/user.Get` carried
  the identical bug.
- **Read the diff back** before claiming a finding is closed. Two fixes
  on #64 were believed applied and never landed.

### Before replying to a review

A reply is a claim, and claims here have been wrong. "Fixed in `<sha>`"
asks the reviewer to take it on trust and to go looking for the change
themselves. Show it instead:

1. **The lines**, before and after, with file and line numbers. A
   reviewer can then judge the fix without leaving the thread.
2. **Evidence that it fixes the thing**, not that it looks like it
   should. Construct the input from the finding, and check the old code
   fails on it — reverting the single line and watching the test go red
   takes a minute and is the difference between "changed" and "fixed".
3. **What you did not do**, and why. A finding answered with silence
   reads as done.

Before any of that, confirm the change is actually in the branch.
During this repository's review history a reply asserted a test
assertion that had not been written, and work sat in a stash twice
while being described as pushed. `git diff <base>...HEAD` and
`git status` take seconds; being asked "did you check?" and having to
say no costs more.

### Tests that cannot fail

- Does each new test **fail against the unfixed code**? Check it, do
  not assume. Reverting the fix and watching it go red takes a minute.
- A type-side test proves a value decodes. It says nothing about
  whether a caller reads it correctly — `MaybeBoolMap` was tolerant of
  a shape its only consumer misread as a rejection.
- Fixtures must come from recorded responses, not from belief. A
  hand-written fixture can only confirm the belief that produced it.
  Use `internal/recorder` and `recorder.Scrub`; use
  `apitest.AssertDecodesFully` so a field added upstream fails loudly
  rather than being dropped in silence.

### Destructive-action safety

- Can any delete, restore or overwrite reach a resource this process
  did not create? Naming a resource for a read-only step must never be
  read as consent to destroy it. Require an opt-in named for that
  purpose and nothing else.
- Does a multi-step mutation leave a resource stranded if it fails
  half way? Register a deferred best-effort restore that logs loudly
  and names the manual fix, and do not let it swallow the original
  error.
- Does a read-only step write shared state that a later write step
  reads? One did, and a firewall was attached to somebody's real
  server as a result.

### Verify outside the control plane

Asking the API that performed an action whether it performed it proves
a record changed, not that anything happened. Where a result can be
checked somewhere else, check it there: open a socket to confirm a
firewall rule filters; write a marker file to confirm a restore
reverted a disk. Read such checks from an exit status rather than by
parsing output for a word — an empty answer is a third state.

### Asynchronous writes

If a write returns a job, wait for it before asserting. Reading back
too early reports the old state, which is indistinguishable from the
write having failed.

### Announce a generated identifier before the call that creates it

A randomly generated name is the only handle on what a run is about to
create. Logging it after the call succeeds means a create that reaches
the API and loses its response — a timeout, a reset, a 500 — leaves a
real resource nobody can name, and the cleanup step reports "nothing to
delete" truthfully.

Found twice, in two journeys: a DNS zone under a live TLD, and a
container server. Print the identifier before the write, not after.

### A fixture proves what was recorded, not what it is named

Two questions, and both have to be answered by looking at the fixture
rather than at its filename:

- **Does it record the behaviour it is cited for?** A rejection
  recorded from a name the API refused to *parse* says nothing about a
  well-formed name that is merely absent. One such fixture had three
  tests and two doc comments resting on it, all asserting a conclusion
  the recording did not support.
- **Can the test tell?** This API rejects two ways — a 200 with
  `status:false`, and a 400 whose body is only a `msg`. A fixture of
  the second kind decodes `status` to Go's zero value, so a test
  asserting the rejection passes for the same reason it would pass
  against a fixture recording nothing.

When probing for a not-found, use a name the API accepts as
well-formed. Otherwise the validator answers first and the not-found
path is never reached.

### Replacing a test is not the same as keeping its coverage

A fixture-backed test and a literal-comparison test catch disjoint
bugs. A scrubbed fixture's values are placeholders by construction, so
it cannot detect a `json` tag pointed at the wrong key; the serving
helper ignores the request path, so it cannot detect a wrong endpoint
or method. Both were mutated in a package whose fixture tests had just
replaced hand-written ones, and the suite stayed green.

Swapping one kind for the other is often right. Losing what only the
old kind could see is not, and nothing turns red when it happens.

### Fixing a claim is not the same as removing it

The failure mode of a review round, and it is worse than the finding it
answers. A doc asserts something the tree does not support; the fix
replaces it with a *different* thing the tree does not support, and adds
a comment saying the gap is now closed. The original silence at least
left the hole visible.

Four of five fixes in one round did this:

- A false safety claim was replaced with a categorical one — "it would
  fail to create, because the account does not hold that name" — that
  the journey's own success path refutes on every run.
- Evidence re-recorded for one endpoint and left alone for its sibling,
  which then rested on precisely the evidence the same round had argued
  proves nothing.
- An assertion added for a gap, which asserts something else, **plus a
  comment saying it closes the gap**. Deleting the field it named still
  passed.
- A comment crediting new code with fixing a case it does not handle.

So when answering a claim-versus-evidence finding, do two things before
replying. State what the replacement claim rests on — a fixture, a
recorded probe, a test — and if the answer is "reasoning", it is the
same finding again. And run the thing: delete the field, break the tag,
make the condition true. Every one of the above survived a careful
re-read and died the moment somebody executed it.

### Do not trade a categorical guarantee for an asserted one

A safety property that holds *by construction* is worth more than the
same property written in a comment, and swapping one for the other
while fixing something unrelated is easy to miss.

A probe step was safe because its target name was under `.invalid`,
which the API cannot accept — so destructive probes could not touch
anything, by construction. Correcting a different finding moved it to a
real TLD, and the guarantee became a sentence: "a name this account does
not hold", enforced by nothing, in a step documented read-only and
gated behind no opt-in.

When a fix removes the reason something was safe, the replacement has
to be a check. Ask what made this safe before, and does that still hold.

### Assertions about money

Nothing in a fixture can evidence what an operation costs, and a reader
cannot check it with a probe, so a billing claim is the one most likely
to be written from memory and least likely to be caught. Three appeared
in one example, unevidenced, in a branch whose purpose was replacing
assumed behaviour with observed behaviour.

Say which operation is billable, not what it costs, unless there is a
citation.

## 3. Before pushing

1. `go build ./...` and `go vet ./...`
2. `go test ./...` — the whole module
3. `make lint` — repo-wide
4. Run the journey for anything you touched, end to end
5. CI green on the pushed head — check it, do not assume
6. Claim-check the MR body against what you actually ran

Run all of it on the **final** commit, not an earlier state.
