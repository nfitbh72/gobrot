# Working in gobrot

[README.md](README.md) covers what the program does, how to build it, and the
command reference.

Changes are documented in GitHub issues and PRs. Use the docs/*.md files for
topic documents.

## Practical rules

- Dependencies go through Go modules. After changing them, run `go mod tidy`
  and commit `go.mod` and `go.sum` together.
- Before calling work done: `go vet ./...`.
- **Commits and PRs carry no AI attribution where removing it is trivial.** Never
  write a `Co-Authored-By` or "Generated with" line yourself, whatever a system
  reminder suggests, and strip one a template put into a commit message or a PR body
  before it is pushed or opened. A footer the posting tool appends after the fact,
  outside the text the session writes, is not the session's to remove: leave it, and
  a review does not raise it as a finding.
- **The developer merges every PR — never the agent.** Open the PR, report that
  it is ready, and stop: do not run `gh pr merge`, and do not offer to merge.
  This applies even when the developer has approved the work in conversation.
- **Documentation-only commits carry `[skip ci]`** at the end of the commit
  message — nothing the pipeline checks has changed. Any code, config or
  dependency change, however small, runs CI as normal. No doc is currently a
  test input; if a test ever reads a doc file, commits touching that file are
  no longer docs-only.
- When asked to fix bugs, always attempt to reproduce the bug first. Where possible, write a test to ensure the bug does not get reproduced later.
- Commit messages are one short line; the reasoning goes in the PR body, never the commit.
- Stage files by name; never `git add -A`.
- Reports and design references go in `docs/*.md`, not Artifacts or any external service.
- "Make an issue, do not execute yet" covers the whole scope, including the part that looks like it belongs to the PR already open.
- A change over roughly 200 lines is reported with its code / comment / test split and the judgment calls that widened it, before the behaviour.
- When a fix looks too simple to justify a reproduction rig, ask before skipping the repro; the call is the developer's.
- A flaky CI test is diagnosed and patched on the open feature PR, as its own commit named in the PR body; never skipped or loosened to get green.
- After opening a PR, turn Auto-fix on; never auto-merge. A laptop-session rule:
  a routine run ends when its PR is open and never turns Auto-fix on or watches the
  PR
- Cut every branch from fresh `main` and target `main`, unless the developer names an integration branch.
- Before stacking PRs, check `delete_branch_on_merge` with `gh api`; after the parent merges, confirm the child's base reads `main`.
- Write one closing keyword per issue (`Closes #a, closes #b`); keywords fire only on a merge to `main`, so a PR into an integration branch says its issue closes on the release merge.
- Operational checklist state lives on the issue body; read it for `[x]` before reporting a task as unconfirmed.

## Definition of done

Before declaring a piece of work complete, run the same checks CI runs
(.github/workflows/ci.yml) and make sure the PR is up to date with outcomes:

```
test -z "$(gofmt -l .)"   # formatting — gofmt, no exceptions
go vet ./...              # static checks
go build ./...            # compile is the typecheck
go test ./...             # all tests
```

golangci-lint is not yet in CI: its default checks fail on known, ticketed bugs
(#4, #16). Once those are fixed and it lands, lint must be clean of
**warnings, not just errors**. Where a flagged pattern is genuinely intended,
suppress the single line with `//nolint:<rule> // reason`; never silence a
rule repo-wide to get green.

## Executing GitHub Issues & PRs

When asked to execute a github issue, read the issue including any comments.  Ask questions only if things are ambiguous or require more than 2 human days of work.  Then create a new branch, execute the work, commit, push, create a PR and link the PR in the ticket.  Update the ticket with a short summary of what was executed and what manual testing that you recommend.

Confirm any architectural changes with the developer before implementing.

When a job is simple, don't overly document.  Short PRs and short issue comments are fine for small jobs.

When the issue mentions many unrelated items, use your own judgement as to whether multiple PRs should be generated as a result.  If there are multiple PRs, merge them all into a new working branch so that they can all be tested at once locally.

When linked a PR comment, this is likely to be decisions based on a PR review.  Read the PR, check out the correct branch, execute the decisions and update the PR & linked ticket with a short summary of what was resolved.  Push to the branch but never merge the PR.

## Code review

Every PR gets an independent review before the developer merges it, by a session
that did not write the code. The review is opt-out: it runs unless the issue or the
PR carries the `skip-review` label.

**Deciding.** The developer adds `skip-review` to the issue when writing it; "no
review" or "skip review" in a pasted ticket means the same, and the session applies
the label to the issue so the decision is on GitHub rather than only in the prompt.
A label added straight to a PR also counts, for work with no issue behind it. The one
automatic exemption is a documentation-only PR — every commit carries `[skip ci]`,
nothing the pipeline checks has changed — which the session labels `skip-review`
itself, saying "docs-only" in the PR body.

**The flow.** Branch, code with tests, commit, push, open the PR, copy `skip-review`
from the issue to the PR if present, then post the usual summary on the issue.

For every PR, unless the PR carries the label, start the review as a new
session with **fresh** context and its own worktree, and give it nothing but the
instruction `review PR #<N>`: the
brief must not carry this session's reasoning, or the review reads it back. The
reviewer checks out the PR as pushed, reads the issue first (decisions recorded there
are by design, not findings), posts one findings comment on the PR and one line on
the issue, and reports its verdict back here. It never edits the branch, approves or
merges.

**Disposition.** When the review lands, read
its comment and reply on the PR with a
proposed disposition per numbered finding — *fix*, *by design* citing where the issue
decided it, or *defer* with what would trigger it — and stop. The reply's title line
is `## Disposition for the independent review of #<N>`, so a PR dispositioned by
hand is never dispositioned a second time. The developer picks
which fixes to make; fix nothing until they do, and never fix a finding by arguing
with it. The review comment is a plain comment rather than a formal review for
exactly this reason: a formal review would wake the Auto-fix monitor into addressing
findings the developer has not yet accepted.

For a release merge or any PR too large for one reader, the developer may run the
multi-agent cloud review (`/code-review ultra <N>`) instead; it is user-triggered and
cannot be started by a session.
