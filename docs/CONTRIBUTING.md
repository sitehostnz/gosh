## Welcome to GoSH docs contributing guide
Thank you for investing your time in contributing to our project!

Read our [Code of Conduct](./CODE_OF_CONDUCT.md) to keep our community approachable and respectable.

In this guide you will get an overview of the contribution workflow from opening an issue, creating a PR, reviewing, and merging the PR.

### Getting started
Staring by having a look at our [style guide](https://github.com/sitehostnz/go-style-guide/blob/master/style.md). Our linter is closely based on the style guide.

### Issues

#### Create a new issue
If you spot a problem with the docs, [search if an issue already exists](https://github.com/sitehostnz/gosh/issues).
If a related issue doesn't exist, you can open a [new issue](https://github.com/sitehostnz/gosh/issues/new).

#### Solve an issue

Scan through our [existing issues](https://github.com/sitehostnz/gosh/issues) to find one that interests you. As a general rule, we don’t assign issues to anyone. If you find an issue to work on, you are welcome to open a PR with a fix.

### Make Changes

#### Repository Structure
Our GoSH package is closely mimic our [SiteHost public API endpoints](https://docs.sitehost.nz/), therefore the folder structure under `/pkg/api/` should be similar to our endpoints. For example:
- `/pkg/api/job` represents [the job endpoint](https://docs.sitehost.nz/api/v1.2/?path=/job).
- `/pkg/api/server` represents [the server endpoint](https://docs.sitehost.nz/api/v1.2/?path=/server).

When adding a new endpoint, please make sure to follow our API structure. 

#### Make changes locally

1. Fork the repository.
   - Using GitHub Desktop:
       - [Getting started with GitHub Desktop](https://docs.github.com/en/desktop/installing-and-configuring-github-desktop/getting-started-with-github-desktop) will guide you through setting up Desktop.
       - Once Desktop is set up, you can use it to [fork the repo](https://docs.github.com/en/desktop/contributing-and-collaborating-using-github-desktop/cloning-and-forking-repositories-from-github-desktop)!

   - Using the command line:
       - [Fork the repo](https://docs.github.com/en/github/getting-started-with-github/fork-a-repo#fork-an-example-repository) so that you can make your changes without affecting the original project until you're ready to merge them.

2. Install or update to **Golang v1.19**. Or you can use docker image.
3. Create a working branch and start with your changes!

### Commit your update

Commit the changes once you are happy with them. Don't forget to [self-review](/contributing/self-review.md) to speed up the review process:zap:.

## Pull Request

When you're finished with the changes, create a pull request, also known as a PR.
- Fill the "Ready for review" template so that we can review your PR. This template helps reviewers understand your changes as well as the purpose of your pull request.
- Don't forget to [link PR to issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue) if you are solving one.
- Enable the checkbox to [allow maintainer edits](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/allowing-changes-to-a-pull-request-branch-created-from-a-fork) so the branch can be updated for a merge.
  Once you submit your PR, a team member will review your proposal. We may ask questions or request additional information.
- We may ask for changes to be made before a PR can be merged, either using [suggested changes](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/incorporating-feedback-in-your-pull-request) or pull request comments. You can apply suggested changes directly through the UI. You can make any other changes in your fork, then commit them to your branch.
- As you update your PR and apply changes, mark each conversation as [resolved](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/commenting-on-a-pull-request#resolving-conversations).
- If you run into any merge issues, checkout this [git tutorial](https://github.com/skills/resolve-merge-conflicts) to help you resolve merge conflicts and other issues.

### Your PR is merged!

Congratulations :tada::tada: The SiteHost team thanks you :sparkles:.



## Examples share the root module

`examples/` lives in the root module rather than carrying its own
`go.mod`. That makes `golang.org/x/crypto` a direct requirement of
`github.com/sitehostnz/gosh`, because `examples/server` dials SSH to
prove an address swap worked from inside the guests.

This has been raised in review twice, so the reasoning is recorded here
rather than re-argued each time.

### What it actually costs a consumer

Measured, not assumed. A module requiring `gosh` and importing only
`pkg/api` and `pkg/api/server`:

- `golang.org/x/crypto` is **not** in the consumer's `go.mod`;
- it is **not in their `go.sum`**, so it is never downloaded or
  verified and carries no supply-chain or CVE surface for them;
- nothing from it is compiled into their binary.

It appears only in `go mod graph`, as
`github.com/sitehostnz/gosh golang.org/x/crypto`. That is module-graph
pruning (Go 1.17+) working as intended.

So the real cost is dependency-review tooling that reads the module
graph rather than the build graph. `govulncheck` does reachability
analysis and does not flag it.

One trap when re-checking this: `go list -deps` in a consumer reports
several `golang.org/x/crypto/...` paths, which looks conclusive until
you read them — they are `vendor/golang.org/x/crypto/...`, the standard
library's own vendored copy behind `crypto/tls`. Unrelated to this
dependency.

### Why not split the module

The conventional fix is a separate `examples/go.mod`. It does not work
here as stated: Go's `internal` rule is per-module, so an `examples`
module could not import `github.com/sitehostnz/gosh/internal/...`, and
the examples are built on the API recording tooling that lives there.

Making the split work would mean promoting that tooling to exported
API. That may be worth doing on its own merits — a consumer testing
against this API would plausibly want a recorder — but it is a decision
about the SDK's public surface, not a dependency clean-up.

The alternative, dropping the SSH dependency, costs the two steps that
verify a swap from inside the guest, which are the ones that make the
example a check rather than a demonstration.

### The standing decision

Keep `examples/` in the root module. `tools.go` already places
`golangci-lint` and `go-acc` in the same `require` block, so this is
consistent with existing practice rather than a new departure.

Revisit if either of these changes:

- a consumer reports real friction in dependency review, rather than
  theoretical friction; or
- the recording tooling is exported for its own reasons, at which point
  the module split becomes cheap and should be done.

## Examples are the validation layer

`examples/<package>/main.go`, one per package. They are assertion-style
rather than demonstration-style: every check must be able to fail, and
a failed assertion exits non-zero.

Two rules that are easy to get wrong:

- **Never log customer data.** Counts, ids and shapes only.
- **A check that cannot fail is worse than no check**, because it reads
  as coverage. Guard loops with a `len(rows) > 0` precondition, or skip
  explicitly and say so — a loop body that never runs will happily log
  a tick.

Where an example can verify a result somewhere other than the API that
was asked to do the work, it should. Asking the control plane whether
the control plane did what it was told proves a record changed, not
that anything happened: the security-group step opens a TCP connection
to confirm a rule filters, and the snapshot step writes a marker file
into the guest to confirm a restore reverted the disk.
