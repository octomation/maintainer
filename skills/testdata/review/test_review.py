import re
import subprocess
import contextlib
import io
import datetime
import json
import os
import shutil
import sys
import tempfile
import unittest
import importlib.util
from unittest import mock
from pathlib import Path

SKILL = Path(__file__).resolve().parent.parent.parent / "review"
sys.dont_write_bytecode = True  # keep __pycache__ out of the skill folder
_spec = importlib.util.spec_from_file_location("review", SKILL / "scripts" / "review.py")
review = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(review)

FRONT = """---
uid: 5584fbb3-4e94-4399-997d-fc65767a7d91
timestamp: 2026-11-25T10:15:00Z
model: claude-opus-5.5-medium
---
"""
NAME = "20261125T101500Z.md"

CODE = FRONT + """
# Code review

## Verdict

**BLOCKED** — F1 loses the state file when the disk is full.

| Critical | High | Medium | Low |
| --- | --- | --- | --- |
| 1 | 0 | 1 | 1 |

## Scope

| Field | Value |
| --- | --- |
| Revision | `1848546` on `main`, clean tree |
| Target | `v0.5.0..HEAD` |
| Reviewed | 14 of 14 files |
| Excluded | `docs/`, `tools/` |
| Accepted | [[MAIN-93]] |

## Findings

| ID | Severity | Lens | Location | Finding | Tracked |
| --- | --- | --- | --- | --- | --- |
| F1 | Critical | Data safety | `internal/state/store.go:117` | A failed save leaves an empty state file | — |
| F2 | Medium | Contract | `internal/command/fetch/root.go:104` | A negative `--concurrency` value is accepted | [[MAIN-131]] |
| F3 | Low | Simplicity | `internal/service/status/view.go:125` | The header counter repeats the row filter | — |

### F1. A failed save leaves an empty state file

- **Failure:** The disk fills during `fetch --apply`. `Store.Save` truncates the state file before it writes. The next run sees no records and plans a clone for each checkout.
- **Evidence:** reproduced — a scratch test with a 1 KB file system limit leaves a 0-byte `state.json`.
- **Fix:** A save must replace the state file only after the new content is complete on disk.

### F3. The header counter repeats the row filter

- **Failure:** A change of the filter in one place gives a wrong header count.
- **Evidence:** inspected — `view.go:125` and `view.go:160` hold the same predicate.
- **Fix:** One function must hold the predicate.

## Coverage

| Lens | Checked | Findings | Evidence |
| --- | --- | --- | --- |
| Gates | full | 0 | `make check` exits 0 |
| Correctness | full | 0 | Read the 14 files of the target |
| Data safety | full | 1 | Ran 3 scratch tests on the save path |
| Contract | full | 1 | Compared `--help` output with `docs/content/fetch.md` |
| Tests | full | 0 | Read the tests of the changed packages |
| Simplicity | full | 1 | Read the 14 files of the target |
"""

ISSUE = FRONT + """
# Issue review

## Verdict

**ACTION** — F1 duplicates an open issue.

| Critical | High | Medium | Low |
| --- | --- | --- | --- |
| 0 | 0 | 1 | 0 |

## Scope

| Field | Value |
| --- | --- |
| Revision | `1848546` on `main`, clean tree |
| Target | open issues |
| Reviewed | 2 of 2 issues |
| Excluded | none |
| Accepted | none |

## Findings

| ID | Severity | Lens | Location | Finding | Tracked |
| --- | --- | --- | --- | --- | --- |
| F1 | Medium | Single | [[MAIN-93]], [[MAIN-131\\|#131]] | Two issues ask for the same flag check | — |

### F1. Two issues ask for the same flag check

- **Failure:** The maintainer plans the same work twice.
- **Evidence:** inspected — `.github/issues/MAIN-93.md:12` and `.github/issues/MAIN-131.md:9` state the same criterion.
- **Fix:** Merge [[MAIN-131]] into [[MAIN-93]].

## Coverage

| Lens | Checked | Findings | Evidence |
| --- | --- | --- | --- |
| Gates | full | 0 | `maintainer.py lint` exits 0 |
| Valid | full | 0 | Read the 2 issues |
| Ready | full | 0 | Read the 2 issues |
| Single | full | 1 | Compared the criteria of the 2 issues |
| True | full | 0 | Read the 2 issues |
| Placed | full | 0 | Read the board file |
"""

SECURITY = FRONT + """
# Security review

## Verdict

**BLOCKED** — F1 lets a web page change the state of the hub.

| Critical | High | Medium | Low |
| --- | --- | --- | --- |
| 0 | 1 | 0 | 0 |

## Scope

| Field | Value |
| --- | --- |
| Revision | `1848546` on `main`, clean tree |
| Target | full |
| Reviewed | 20 of 20 files |
| Excluded | none |
| Accepted | none |

## Findings

| ID | Severity | Lens | Location | Finding | Tracked |
| --- | --- | --- | --- | --- | --- |
| F1 | High | Service | `internal/service/hub/server.go:40` | The hub accepts a state change without an `Origin` check | — |

### F1. The hub accepts a state change without an `Origin` check

- **Failure:** A2: The user opens a hostile page. The page sends a `POST` to the hub on loopback.
- **Evidence:** inspected — `server.go:40` reads no `Origin` header.
- **Fix:** The hub must refuse a state change from a foreign origin.

## Coverage

| Lens | Checked | Findings | Evidence |
| --- | --- | --- | --- |
| Gates | partial | 0 | `govulncheck` is absent; the dependency scan did not run |
| Files | full | 0 | Read the path code |
| Processes | full | 0 | Read the calls of `git` |
| Secrets | full | 0 | Read the token code |
| Service | full | 1 | Read the hub handlers |
| Supply chain | full | 0 | Read the workflows |
"""

EMPTY = FRONT + """
# Code review

## Verdict

**CLEAR** — No finding in the target.

| Critical | High | Medium | Low |
| --- | --- | --- | --- |
| 0 | 0 | 0 | 0 |

## Scope

| Field | Value |
| --- | --- |
| Revision | `1848546` on `main`, clean tree |
| Target | `v0.5.0..HEAD` |
| Reviewed | 3 of 3 files |
| Excluded | none |
| Accepted | none |

## Findings

None.

## Coverage

| Lens | Checked | Findings | Evidence |
| --- | --- | --- | --- |
| Gates | full | 0 | `make check` exits 0 |
| Correctness | full | 0 | Read the 3 files |
| Data safety | full | 0 | Read the 3 files |
| Contract | full | 0 | Read the 3 files |
| Tests | full | 0 | Read the 3 files |
| Simplicity | full | 0 | Read the 3 files |
"""

TEMPLATE = """---
uid: <uid>
timestamp: <timestamp>
model: <model>
---
""" + CODE.split("---\n", 2)[2].replace("| `v0.5.0..HEAD` |", "| <target> |")


class Base(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        git_init(self.root)
        issues = self.root / ".github/issues"
        issues.mkdir(parents=True)
        for n in (93, 131):
            (issues / ("MAIN-%d.md" % n)).write_text("---\nnumber: %d\nstate: OPEN\n---\n" % n)
        (issues / "MAIN-7.md").write_text("---\nnumber: 7\nstate: CLOSED\n---\n")

    def tearDown(self):
        self.tmp.cleanup()

    def run_check(self, text, folder=".github/reports"):
        path = self.root / folder / NAME
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text)
        return review.check(path, self.root)

    def rules(self, text, folder=".github/reports"):
        errors, _ = self.run_check(text, folder)
        return {e.split(": ")[1] for e in errors}

    def assertRule(self, text, rule, folder=".github/reports"):
        errors, _ = self.run_check(text, folder)
        self.assertIn(rule, {e.split(": ")[1] for e in errors}, errors)


class Valid(Base):
    def test_code(self):
        errors, out = self.run_check(CODE)
        self.assertEqual(errors, [])
        self.assertEqual(out, [
            "Code review: BLOCKED — Critical 1, High 0, Medium 1, Low 1.",
            "Report: .github/reports/" + NAME,
            "First: F1 A failed save leaves an empty state file.",
            "Coverage: full."])

    def test_issue(self):
        self.assertEqual(self.run_check(ISSUE)[0], [])

    def test_security(self):
        errors, out = self.run_check(SECURITY, ".agents/reports")
        self.assertEqual(errors, [])
        self.assertEqual(out[3], "Coverage: partial: Gates.")
        self.assertEqual(out[4], review.PRIVATE_LINE)

    def test_empty(self):
        errors, out = self.run_check(EMPTY)
        self.assertEqual(errors, [])
        self.assertEqual(out[2], "First: none.")


class Invalid(Base):
    def test_section_order(self):
        text = CODE.replace("## Scope", "## Tmp").replace("## Findings", "## Scope").replace("## Tmp", "## Findings")
        self.assertRule(text, "structure")

    def test_clear_with_high(self):
        text = CODE.replace("| F1 | Critical |", "| F1 | High |").replace("**BLOCKED**", "**CLEAR**")
        self.assertRule(text, "verdict")

    def test_clear_with_partial(self):
        text = EMPTY.replace("| Tests | full |", "| Tests | partial |")
        self.assertRule(text, "verdict")

    def test_counts(self):
        self.assertRule(CODE.replace("| 1 | 0 | 1 | 1 |", "| 1 | 0 | 2 | 1 |"), "counts")

    def test_untracked_without_block(self):
        text = CODE.split("### F3.")[0] + "## Coverage" + CODE.split("## Coverage")[1]
        self.assertRule(text, "block")

    def test_tracked_with_block(self):
        block = ("### F2. A negative `--concurrency` value is accepted\n\n"
                 "- **Failure:** A run gets a negative value.\n"
                 "- **Evidence:** inspected — `root.go:104` has no check.\n"
                 "- **Fix:** The value must be positive.\n\n### F3.")
        self.assertRule(CODE.replace("### F3.", block), "block")

    def test_sort_order(self):
        text = CODE.replace("| F2 | Medium |", "| F2 | Low |").replace("| F3 | Low |", "| F3 | Medium |")
        self.assertRule(text, "order")

    def test_evidence_prefix(self):
        self.assertRule(CODE.replace("reproduced — a scratch", "a scratch"), "block")

    def test_budget(self):
        text = CODE.replace("- **Fix:** One function must hold the predicate.",
                            "- **Fix:** One function must hold it. Tests must cover it.")
        self.assertRule(text, "budget")

    def test_banned_word(self):
        self.assertRule(CODE.replace("Ran 3 scratch tests", "Really ran 3 scratch tests"), "language")

    def test_banned_word_in_backticks(self):
        text = CODE.replace("Ran 3 scratch tests", "Ran `really` 3 scratch tests")
        self.assertEqual(self.run_check(text)[0], [])

    def test_security_under_github(self):
        self.assertRule(SECURITY, "privacy", ".github/reports")

    def test_security_published(self):
        path = self.root / ".github/reports" / NAME
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(SECURITY)
        errors, _ = review.check(path, self.root, published=True)
        self.assertEqual({e.split(": ")[1] for e in errors}, {"publication"}, errors)
        path = self.root / "docs" / NAME
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(SECURITY)
        errors, _ = review.check(path, self.root, published=True)
        self.assertIn("privacy", {e.split(": ")[1] for e in errors}, errors)

    def test_home_path(self):
        self.assertRule(CODE.replace("`make check` exits 0", "`/Users/me/x` exits 0"), "privacy")

    def test_generic_home_folder(self):
        text = CODE.replace("`make check` exits 0", "`~/.config/git` holds no token")
        self.assertEqual(self.run_check(text)[0], [])

    def test_code_in_backticks(self):
        text = CODE.replace("`make check` exits 0", '`<App path="/tablo">` and `<T>` render')
        self.assertEqual(self.run_check(text)[0], [])

    def test_placeholder(self):
        self.assertRule(CODE.replace("| Excluded | `docs/`, `tools/` |", "| Excluded | <excluded> |"), "placeholder")

    def test_tracked_without_file(self):
        self.assertRule(CODE.replace("[[MAIN-131]] |", "[[MAIN-9999]] |"), "tracked")


class Regressions(Base):
    HIGH = "| F1 | High | Data safety | `a.go:1` | A save loses the state file | — |"

    def test_none_with_table(self):
        table = "| ID | Severity | Lens | Location | Finding | Tracked |\n| --- | --- | --- | --- | --- | --- |\n"
        errors, out = self.run_check(EMPTY.replace("None.\n", "None.\n\n" + table + self.HIGH + "\n"))
        self.assertIn("findings", {e.split(": ")[1] for e in errors}, errors)

    def test_second_table_before_blocks(self):
        table = "| ID | Severity | Lens | Location | Finding | Tracked |\n| --- | --- | --- | --- | --- | --- |\n"
        low = ("| F1 | Low | Simplicity | `a.go:1` | A name misleads the reader | — |\n\n" + table + self.HIGH +
               "\n\n### F1. A name misleads the reader\n\n- **Failure:** A reader misreads the code.\n"
               "- **Evidence:** inspected — `a.go:1` holds the name.\n- **Fix:** The name must match the role.\n")
        text = EMPTY.replace("None.\n", table + low).replace("| Simplicity | full | 0 |", "| Simplicity | full | 1 |")
        text = text.replace("| 0 | 0 | 0 | 0 |", "| 0 | 0 | 0 | 1 |")
        errors, out = self.run_check(text)
        self.assertIn("findings", {e.split(": ")[1] for e in errors}, errors)
        with contextlib.redirect_stdout(io.StringIO()) as out:
            code = review.main(["check", str(self.root / ".github/reports" / NAME), "--root", str(self.root)])
        self.assertEqual(code, 1)
        self.assertNotIn("CLEAR", out.getvalue())

    def test_second_table_after_blocks(self):
        table = "\n| ID | Severity | Lens | Location | Finding | Tracked |\n| --- | --- | --- | --- | --- | --- |\n"
        self.assertRule(CODE.replace("\n## Coverage", table + self.HIGH + "\n\n## Coverage"), "findings")

    def test_prose_in_findings(self):
        self.assertRule(CODE.replace("### F1.", "Some prose.\n\n### F1."), "findings")

    def test_private_code_report(self):
        errors, out = self.run_check(CODE, ".agents/reports")
        self.assertEqual(errors, [])
        self.assertEqual(out[-1], review.PRIVATE_LINE)

    def test_report_in_docs(self):
        self.assertRule(CODE, "privacy", "docs")

    def test_security_in_docs(self):
        self.assertRule(SECURITY, "privacy", "docs")

    def test_symlinked_folder(self):
        (self.root / "outside/reports").mkdir(parents=True)
        (self.root / ".agents").symlink_to(self.root / "outside")
        self.assertRule(SECURITY, "privacy", ".agents/reports")

    def test_blocked_names_first_three(self):
        rows = "\n".join("| F%d | High | Data safety | `a.go:%d` | A save loses file %d | — |" % (i, i, i)
                         for i in range(1, 6))
        text = CODE.replace(CODE.split("| --- | --- | --- | --- | --- | --- |\n")[1].split("\n\n")[0], rows)
        self.assertIn("the BLOCKED reason must name F2", "\n".join(self.run_check(text)[0]))
        self.assertIn("and 2 more", "\n".join(self.run_check(text)[0]))

    def test_empty_finding_cell(self):
        self.assertRule(CODE.replace("| A negative `--concurrency` value is accepted |", "|  |"), "findings")

    def test_tracked_closed(self):
        self.assertRule(CODE.replace("[[MAIN-131]] |", "[[MAIN-7]] |"), "tracked")

    def test_revision(self):
        self.assertRule(CODE.replace("`1848546` on `main`, clean tree", "1848546 on main"), "scope")

    def test_evidence_cites(self):
        self.assertRule(CODE.replace("inspected — `view.go:125` and `view.go:160` hold the same predicate.",
                                     "inspected — confirmed."), "block")

    def test_image_in_backticks(self):
        text = CODE.replace("Ran 3 scratch tests", "Ran `![alt](x)` and `<b>x</b>` tests")
        self.assertEqual(self.run_check(text)[0], [])


class ClosedForm(Base):
    def test_prose_after_counts(self):
        self.assertRule(CODE.replace("| 1 | 0 | 1 | 1 |\n", "| 1 | 0 | 1 | 1 |\n\nMore words here.\n"), "structure")

    def test_list_after_counts(self):
        self.assertRule(CODE.replace("| 1 | 0 | 1 | 1 |\n", "| 1 | 0 | 1 | 1 |\n\n- A list item.\n"), "structure")

    def test_prose_after_scope(self):
        self.assertRule(CODE.replace("| Accepted | [[MAIN-93]] |\n", "| Accepted | [[MAIN-93]] |\n\nMore words.\n"),
                        "structure")

    def test_prose_after_coverage(self):
        self.assertRule(CODE + "\nMore words after the table.\n", "structure")

    def test_prose_before_verdict(self):
        self.assertRule(CODE.replace("# Code review\n", "# Code review\n\nAn intro line.\n"), "structure")

    def test_long_span(self):
        span = "`make " + "x" * 90 + "`"
        self.assertRule(CODE.replace("`make check` exits 0", span + " exits 0"), "budget")


STEM = NAME[:-len(".md")]
SOURCE = "Source: [the security review of 2026-11-25](../reports/%s)" % NAME
TRACKED = SECURITY.replace("`Origin` check | — |", "`Origin` check | [[MAIN-131]] |").split("### F1.")[0] \
    + "## Coverage" + SECURITY.split("## Coverage")[1]
NO_FINDING = SECURITY.split("| ID |")[0] + "None.\n\n## Coverage" + SECURITY.split("## Coverage")[1]
NO_FINDING = NO_FINDING.replace("**BLOCKED** — F1 lets a web page change the state of the hub.",
                                "**INCOMPLETE** — The Gates lens did not run in full.")
NO_FINDING = NO_FINDING.replace("| 0 | 1 | 0 | 0 |", "| 0 | 0 | 0 | 0 |").replace("| Service | full | 1 |",
                                                                                   "| Service | full | 0 |")


class Rule(Base):
    """Helpers for the publication rule."""

    def issue(self, n, tail="", state="CLOSED", reason="COMPLETED"):
        text = "---\nnumber: %d\nstate: %s\nstateReason: %s\n---\n\n# An issue\n\n%s\n" % (n, state, reason, tail)
        (self.root / ".github/issues" / ("MAIN-%d.md" % n)).write_text(text)

    def put(self, text=SECURITY, folder=".agents/reports", name=NAME):
        path = self.root / folder / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text)
        return path

    def blockers(self, text=SECURITY):
        """The publication errors of a forced move: the report under .github/reports/."""
        errors, _ = review.check(self.put(text, ".github/reports"), self.root, published=True)
        self.assertEqual({e.split(": ")[1] for e in errors} - {"publication"}, set(), errors)
        return [e.split(": publication: ")[1] for e in errors]

    def run_main(self, *argv):
        with contextlib.redirect_stdout(io.StringIO()) as out:
            code = review.main(list(argv) + ["--root", str(self.root)])
        return code, out.getvalue().splitlines()



class Publication(Rule):
    """A report leaves .agents/reports/ only when each finding has an issue
    and each issue that names the report is closed as completed."""

    def test_no_issue(self):
        errors, out = review.check(self.put(), self.root)
        self.assertEqual(errors, [])
        self.assertEqual(out[4], review.PRIVATE_LINE)
        self.assertEqual(self.blockers(), ["F1 has no issue"])

    def test_closed_issue(self):
        self.issue(200, SOURCE + ", F1.")
        errors, out = review.check(self.put(), self.root)
        self.assertEqual(errors, [])
        self.assertEqual(out[4], review.publish_line(NAME, self.root.resolve()))
        self.assertTrue(out[4].startswith("Publish: python3 "), out[4])
        self.assertTrue(out[4].endswith(
            " check --published .agents/reports/%s && mv .agents/reports/%s .github/reports/%s" % (NAME, NAME, NAME)),
            out[4])
        self.assertEqual(self.blockers(), [])

    def publish(self):
        """Run the Publish: line of the report as the owner does, from the root."""
        try:
            subprocess.run(["git", "init", "-q", str(self.root)], check=True, capture_output=True)
        except (OSError, subprocess.CalledProcessError):
            self.skipTest("no git")
        line = review.publish_line(NAME, self.root.resolve())
        return subprocess.run(line[len("Publish: "):], shell=True, cwd=self.root, capture_output=True, text=True)

    def test_publish_line_quotes_paths(self):
        line = review.publish_line("a b.md", self.root.resolve())
        self.assertIn(" check --published '.agents/reports/a b.md' && mv '.agents/reports/a b.md' "
                      "'.github/reports/a b.md'", line)

    def test_publish_line_moves_the_report(self):
        self.issue(200, SOURCE + ", F1.")
        self.put()
        (self.root / ".github/reports").mkdir()
        done = self.publish()
        self.assertEqual(done.returncode, 0, done.stdout + done.stderr)
        self.assertNotIn("Publish:", done.stdout)
        self.assertFalse((self.root / ".agents/reports" / NAME).exists())
        self.assertEqual((self.root / ".github/reports" / NAME).read_text(), SECURITY)
        self.assertEqual(self.run_main("status"),
                         (0, [".github/reports/%s: published, 1 of 1 findings closed." % NAME]))

    def test_stale_publish_line_leaves_the_report_private(self):
        self.issue(200, SOURCE + ", F1.")
        self.put()
        (self.root / ".github/reports").mkdir()
        self.issue(200, SOURCE + ", F1.", "OPEN", "null")
        done = self.publish()
        self.assertEqual(done.returncode, 1, done.stdout + done.stderr)
        self.assertIn("MAIN-200 holds F1 and is not closed as completed", done.stdout)
        self.assertTrue((self.root / ".agents/reports" / NAME).exists())
        self.assertFalse((self.root / ".github/reports" / NAME).exists())

    def test_url_source(self):
        url = SOURCE.replace("../reports/", "https://github.com/octo/repo/blob/main/.github/reports/")
        self.issue(200, url + ", F1.")
        self.assertEqual(self.blockers(), [])
        self.issue(200, url.replace("/.github/reports/", "/docs/") + ", F1.")
        self.assertIn("F1 has no issue", self.blockers())

    def test_repeated_state_key(self):
        for extra in ("state: OPEN\n", "stateReason: null\n", "state : OPEN\n", '"state": OPEN\n',
                      "'stateReason' : null\n"):
            with self.subTest(extra=extra):
                (self.root / ".github/issues/MAIN-200.md").write_text(
                    "---\nnumber: 200\nstate: CLOSED\nstateReason: COMPLETED\n%s---\n\n%s, F1.\n" % (extra, SOURCE))
                self.assertEqual(self.blockers(), ["MAIN-200 repeats a state key"])

    def test_state_line_forms(self):
        for state in ("state:CLOSED", "state:  CLOSED", '"state": CLOSED', "state: CLOSED # done", "state: 'CLOSED'",
                      " state: CLOSED"):
            with self.subTest(state=state):
                (self.root / ".github/issues/MAIN-200.md").write_text(
                    "---\nnumber: 200\n%s\nstateReason: COMPLETED\n---\n\n%s, F1.\n" % (state, SOURCE))
                self.assertEqual(self.blockers(), ["MAIN-200 holds F1 and is not closed as completed"])

    def test_published_code_report_keeps_the_rule(self):
        (self.root / ".github/issues/MAIN-131.md").write_text(
            "---\nnumber: 131\nstate: CLOSED\nstateReason: COMPLETED\n---\n")
        self.issue(200, SOURCE + ", F1, F3.")
        self.put(CODE, ".github/reports")
        self.assertEqual(self.run_main("status"),
                         (0, [".github/reports/%s: published, 3 of 3 findings closed." % NAME]))
        self.issue(200, SOURCE + ", F1, F3.", "OPEN", "null")
        code, out = self.run_main("status")
        self.assertEqual((code, out[0]), (1, ".github/reports/%s: refused, 1 of 3 findings closed." % NAME))
        text = EMPTY.replace("2026-11-25T10:15:00Z", "2026-11-26T10:15:00Z")
        code, out = self.run_main("check", str(self.put(text, ".github/reports", "20261126T101500Z.md")))
        self.assertEqual(out[4:], ["Refused: .github/reports/%s breaks the publication rule; run status." % NAME])

    def test_published_hand_over(self):
        self.issue(200, SOURCE + ", F1.")
        errors, out = review.check(self.put(folder=".github/reports"), self.root, published=True)
        self.assertEqual(errors, [])
        self.assertEqual(out[1], "Report: .github/reports/" + NAME)
        self.assertEqual(len(out), 4)

    def test_open_issue(self):
        self.issue(200, SOURCE + ", F1.", "OPEN", "null")
        self.assertEqual(self.blockers(), ["MAIN-200 holds F1 and is not closed as completed"])
        self.assertEqual(review.check(self.put(), self.root)[1][4], review.PRIVATE_LINE)

    def test_not_planned_issue(self):
        self.issue(200, SOURCE + ", F1.", "CLOSED", "NOT_PLANNED")
        self.assertEqual(self.blockers(), ["MAIN-200 holds F1 and is not closed as completed"])

    def test_closed_without_reason(self):
        (self.root / ".github/issues/MAIN-200.md").write_text(
            "---\nnumber: 200\nstate: CLOSED\n---\n\n%s, F1.\n" % SOURCE)
        self.assertEqual(self.blockers(), ["MAIN-200 holds F1 and is not closed as completed"])

    def test_second_issue_open(self):
        self.issue(200, SOURCE + ", F1.")
        self.issue(201, SOURCE + ", F1.", "OPEN", "null")
        self.assertEqual(self.blockers(), ["MAIN-201 holds F1 and is not closed as completed"])

    def test_source_without_finding(self):
        self.issue(200, SOURCE + ".")
        self.assertEqual(self.blockers(), [
            "the line must be: Source: [<text>](../reports/%s), F<n>." % NAME, "F1 has no issue"])

    def test_source_forms(self):
        for tail in (", F1", ", F1 .", ", f1.", ",F1.", ", F1, .", ", F1. More text."):
            with self.subTest(tail=tail):
                self.issue(200, SOURCE + tail)
                self.assertIn("the line must be: Source: [<text>](../reports/%s), F<n>." % NAME, self.blockers())

    def test_source_of_other_finding(self):
        self.issue(200, SOURCE + ", F1, F2.")
        self.assertEqual(self.blockers(), ["the report has no F2"])

    def test_mention_without_source(self):
        self.issue(200, SOURCE + ", F1.")
        self.issue(201, "See the review %s for the cause." % STEM)
        self.assertEqual(self.blockers(), ["the issue names the report without a Source line"])

    def test_related_issue_open(self):
        self.issue(200, SOURCE + ", F1.")
        self.issue(201, "See the review %s.\n\n%s, F1." % (STEM, SOURCE), "OPEN", "null")
        self.assertEqual(self.blockers(), ["MAIN-201 holds F1 and is not closed as completed"])

    def test_other_report(self):
        self.issue(200, SOURCE + ", F1.")
        self.issue(201, SOURCE.replace(NAME, "20261125T101501Z.md") + ".", "OPEN", "null")
        self.assertEqual(self.blockers(), [])

    def test_tracked_issue(self):
        self.assertEqual(self.blockers(TRACKED), ["MAIN-131 holds F1 and is not closed as completed"])
        self.issue(131)
        self.assertEqual(self.blockers(TRACKED), [])
        self.assertRule(TRACKED, "tracked", ".agents/reports")

    def test_no_finding(self):
        errors, out = review.check(self.put(NO_FINDING), self.root)
        self.assertEqual(errors, [])
        self.assertTrue(out[4].startswith("Publish: "), out)

    def test_private_code_report(self):
        (self.root / ".github/issues/MAIN-131.md").write_text(
            "---\nnumber: 131\nstate: CLOSED\nstateReason: COMPLETED\n---\n")
        self.issue(200, SOURCE + ", F1, F3.")
        errors, out = review.check(self.put(CODE), self.root, published=True)
        self.assertEqual(errors, [])
        self.assertEqual(len(out), 4)
        code, out = self.run_main("status")
        self.assertEqual(out, [".agents/reports/%s: publishable, 3 of 3 findings closed." % NAME,
                               review.publish_line(NAME, self.root.resolve())])

    def test_private_copy_differs(self):
        self.issue(200, SOURCE + ", F1.")
        self.put(SECURITY.replace("Read the workflows", "Read each workflow"))
        self.assertEqual(self.blockers(), ["the report differs from its copy in .agents/reports/"])

    def test_stale_publish_line_keeps_a_differing_public_file(self):
        self.issue(200, SOURCE + ", F1.")
        self.put()
        other = SECURITY.replace("Read the workflows", "Read each workflow")
        self.put(other, ".github/reports")
        done = self.publish()
        self.assertEqual(done.returncode, 1, done.stdout + done.stderr)
        self.assertIn("the report differs from its copy in .github/reports/", done.stdout)
        self.assertEqual((self.root / ".agents/reports" / NAME).read_text(), SECURITY)
        self.assertEqual((self.root / ".github/reports" / NAME).read_text(), other)

    def test_forced_move_fails(self):
        self.issue(200, SOURCE + ", F1.", "OPEN", "null")
        path = self.put(folder=".github/reports")
        code, out = self.run_main("check", str(path), "--published")
        self.assertEqual(code, 1)
        self.assertEqual(out, [".github/issues/MAIN-200.md:3: publication: "
                               "MAIN-200 holds F1 and is not closed as completed"])
        code, out = self.run_main("check", str(path))
        self.assertEqual(code, 1)
        self.assertIn(": privacy: ", out[0])

    def test_status_private(self):
        self.put()
        code, out = self.run_main("status")
        self.assertEqual(code, 0)
        self.assertEqual(out, [".agents/reports/%s: private, 0 of 1 findings closed." % NAME,
                               ".agents/reports/%s:31: publication: F1 has no issue" % NAME])

    def test_status_publishable(self):
        self.put()
        self.issue(200, SOURCE + ", F1.")
        code, out = self.run_main("status")
        self.assertEqual(code, 0)
        self.assertEqual(out, [".agents/reports/%s: publishable, 1 of 1 findings closed." % NAME,
                               review.publish_line(NAME, self.root.resolve())])

    def test_status_published(self):
        self.put(folder=".github/reports")
        self.issue(200, SOURCE + ", F1.")
        self.assertEqual(self.run_main("status"),
                         (0, [".github/reports/%s: published, 1 of 1 findings closed." % NAME]))

    def test_status_refused(self):
        self.put(folder=".github/reports")
        self.issue(200, SOURCE + ", F1.", "OPEN", "null")
        code, out = self.run_main("status")
        self.assertEqual(code, 1)
        self.assertEqual(out[0], ".github/reports/%s: refused, 0 of 1 findings closed." % NAME)
        self.assertEqual(len(out), 2)

    def test_status_duplicate(self):
        self.put()
        self.put(folder=".github/reports")
        self.issue(200, SOURCE + ", F1.")
        self.assertEqual(self.run_main("status"), (0, [
            ".agents/reports/%s: duplicate of the published report; delete it." % NAME,
            ".github/reports/%s: published, 1 of 1 findings closed." % NAME]))

    def test_status_draft(self):
        self.put(SECURITY.replace("| Target | full |", "| Target | <target> |"))
        self.issue(200, SOURCE + ", F1.")
        code, out = self.run_main("status")
        self.assertEqual(code, 0)
        self.assertEqual(out, [".agents/reports/%s: private, 1 of 1 findings closed." % NAME,
                               ".agents/reports/%s:1: publication: the report has 1 other errors; run check" % NAME])

    def test_status_skips_public_code_report(self):
        self.put(CODE, ".github/reports")
        self.assertEqual(self.run_main("status"), (0, ["No private report and no published security report."]))

    def test_hand_over_proposes_other_report(self):
        self.put()
        self.issue(200, SOURCE + ", F1.")
        text = EMPTY.replace("2026-11-25T10:15:00Z", "2026-11-26T10:15:00Z")
        code, out = self.run_main("check", str(self.put(text, ".github/reports", "20261126T101500Z.md")))
        self.assertEqual(code, 0)
        self.assertEqual(out[4:], [review.publish_line(NAME, self.root.resolve())])
        code, out = self.run_main("check", str(self.root / ".agents/reports" / NAME))
        self.assertEqual(out[4:], [review.publish_line(NAME, self.root.resolve())])

    def test_hand_over_names_refused_report(self):
        self.put(folder=".github/reports")
        text = EMPTY.replace("2026-11-25T10:15:00Z", "2026-11-26T10:15:00Z")
        code, out = self.run_main("check", str(self.put(text, ".github/reports", "20261126T101500Z.md")))
        self.assertEqual(code, 0)
        self.assertEqual(out[4:], ["Refused: .github/reports/%s breaks the publication rule; run status." % NAME])


class Case20261125(Rule):
    """The security review of 2026-11-25: five findings, six issues. The
    issues named the report without a finding."""

    MAP = {177: "F1", 183: "F3", 184: "F4", 185: "F5", 186: "F2", 187: "F1"}
    ROW = "| F%d | High | Service | `internal/service/hub/server.go:4%d` | The hub fails in case %d | — |"
    BLOCK = ("### F%d. The hub fails in case %d\n\n- **Failure:** A2: A hostile page sends request %d.\n"
             "- **Evidence:** inspected — `server.go:4%d` has no check.\n- **Fix:** The hub must refuse request %d.\n")

    def report(self):
        rows = "\n".join(self.ROW % (i, i, i) for i in range(1, 6))
        blocks = "\n".join(self.BLOCK % (i, i, i, i, i) for i in range(1, 6))
        head, rest = SECURITY.split("| F1 |")
        text = head + rows + "\n\n" + blocks + "\n## Coverage" + rest.split("## Coverage")[1]
        text = text.replace("**BLOCKED** — F1 lets a web page change the state of the hub.",
                            "**BLOCKED** — F1, F2, F3 and 2 more let a web page change the state of the hub.")
        return text.replace("| 0 | 1 | 0 | 0 |", "| 0 | 5 | 0 | 0 |").replace("| Service | full | 1 |",
                                                                               "| Service | full | 5 |")

    def issues(self, form):
        for n, fid in self.MAP.items():
            self.issue(n, SOURCE + form(fid))

    def test_as_filed(self):
        self.issues(lambda fid: ".")
        got = self.blockers(self.report())
        self.assertEqual(got.count("the line must be: Source: [<text>](../reports/%s), F<n>." % NAME), 6)
        self.assertEqual(got[6:], ["F%d has no issue" % i for i in range(1, 6)])
        self.put(self.report())
        self.assertEqual(self.run_main("status")[0], 1)

    def test_with_finding_ids(self):
        self.issues(lambda fid: ", %s." % fid)
        self.assertEqual(self.blockers(self.report()), [])
        (self.root / ".github/reports" / NAME).unlink()
        self.put(self.report())
        code, out = self.run_main("status")
        self.assertEqual((code, out[0]), (0, ".agents/reports/%s: publishable, 5 of 5 findings closed." % NAME))

    def test_reopened_issue(self):
        self.issues(lambda fid: ", %s." % fid)
        self.issue(187, SOURCE + ", F1.", "OPEN", "null")
        self.assertEqual(self.blockers(self.report()), ["MAIN-187 holds F1 and is not closed as completed"])
        code, out = self.run_main("status")
        self.assertEqual((code, out[0]), (1, ".github/reports/%s: refused, 4 of 5 findings closed." % NAME))



class OldFolders(Rule):
    """The folders of v0.6.x are not read: a repository moves its reports once."""

    def test_check_refuses(self):
        for folder in (".github/reviews", ".agents/reviews"):
            with self.subTest(folder=folder):
                self.assertRule(CODE, "privacy", folder)

    def test_status_skips(self):
        self.put(folder=".agents/reviews")
        self.put(folder=".github/reviews", name="20261124T101500Z.md")
        self.assertEqual(self.run_main("status"), (0, ["No private report and no published security report."]))

    def test_old_source_link(self):
        self.issue(200, SOURCE.replace("../reports/", "../reviews/") + ", F1.")
        self.assertIn("the line must be: Source: [<text>](../reports/%s), F<n>." % NAME, self.blockers())

SKILL_PATHS = ("references/", "templates/", "scripts/")
# The skill serves any project: the words of one project appear only in
# lines that mark them as an example.
PROJECT_WORDS = re.compile(r"MAIN-|maintainer service|\bhub\b|\bfetch\b|internal/service")


class Guards(unittest.TestCase):
    def test_templates(self):
        for mode, (title, _, lenses) in review.MODES.items():
            with self.subTest(mode=mode):
                lines = list(enumerate((SKILL / "templates" / (mode + ".md")).read_text().splitlines(), 1))
                _, body = review.split_front_matter(lines)
                found, sections = review.check_structure(body, lambda *a: None)
                self.assertEqual(found, mode)
                self.assertEqual(list(sections), review.SECTIONS)
                self.assertEqual([r[1][0] for r in review.table(sections["Scope"])[1]], review.SCOPE_FIELDS)
                self.assertEqual([r[1][0] for r in review.table(sections["Coverage"])[1]], lenses)

    def docs(self):
        return [SKILL / "SKILL.md"] + sorted((SKILL / "references").glob("*.md")) \
            + sorted((SKILL / "templates").glob("*.md")) + [SKILL / "scripts/review.py"]

    def test_paths_in_references(self):
        missing = []
        for doc in self.docs():
            for span in re.findall(r"`([^`]+)`", doc.read_text()):
                if span.startswith(SKILL_PATHS) and not any(c in span for c in "<* $"):
                    if not (SKILL / span).exists():
                        missing.append("%s: %s" % (doc.name, span))
        self.assertEqual(missing, [])

    def test_no_project_words(self):
        found = ["%s:%d: %s" % (doc.name, no, line) for doc in self.docs()
                 for no, line in enumerate(doc.read_text().splitlines(), 1)
                 if PROJECT_WORDS.search(line) and "example" not in line.lower()]
        self.assertEqual(found, [])

    def test_links_stay_in_the_skill(self):
        broken = []
        for doc in self.docs()[:-1]:
            for target in re.findall(r"\]\(([^)\s]+)\)", doc.read_text()):
                path = (doc.parent / target.split("#")[0]).resolve()
                if not re.match(r"^[a-z][a-z0-9+.-]*:|#", target, re.I) and \
                        not (path.is_relative_to(SKILL) and path.exists()):
                    broken.append("%s: %s" % (doc.name, target))
        self.assertEqual(broken, [])

    def test_payload(self):
        files = sorted(str(p.relative_to(SKILL)) for p in SKILL.rglob("*")
                       if p.is_file() and "__pycache__" not in p.parts)
        self.assertEqual([f for f in files if not f.startswith(("references/", "templates/"))],
                         ["SKILL.md", "scripts/review.py"])
        self.assertEqual([p for p in SKILL.rglob("*") if p.is_symlink() or p.name.startswith(".")], [])


class Tables(Base):
    SEP = "| ID | Severity | Lens | Location | Finding | Tracked |\n| --- | --- | --- | --- | --- | --- |"

    def test_separator_replaced(self):
        text = CODE.replace(self.SEP, self.SEP.split("\n")[0] + "\n| Praise and process history outside the form |")
        self.assertRule(text, "structure")

    def test_text_after_closing_pipe(self):
        self.assertRule(CODE.replace("value is accepted | [[MAIN-131]] |",
                                     "value is accepted | [[MAIN-131]] | and some praise"), "structure")

    def test_short_separator(self):
        self.assertRule(CODE.replace("| Lens | Checked | Findings | Evidence |\n| --- | --- | --- | --- |",
                                     "| Lens | Checked | Findings | Evidence |\n| --- |"), "structure")

    def test_colon_separator(self):
        text = CODE.replace("| --- | --- | --- | --- |\n| 1 |", "| :-: | :- | -: | - |\n| 1 |")
        self.assertNotEqual(text, CODE)
        self.assertEqual(self.run_check(text)[0], [])


def git_init(root, ignore=".agents\n"):
    """Make root a Git repository that ignores the lines of ignore."""
    try:
        subprocess.run(["git", "init", "-q", str(root)], check=True, capture_output=True)
    except (OSError, subprocess.CalledProcessError):
        raise unittest.SkipTest("no git")
    (Path(root) / ".gitignore").write_text(ignore)


class RealTemplates(Base):
    def test_each_placeholder_fails(self):
        git_init(self.root)
        now = datetime.datetime(2026, 11, 25, 10, 15, tzinfo=datetime.timezone.utc)
        for mode in review.MODES:
            with self.subTest(mode=mode):
                path = review.new(mode, "m", self.root, now, SKILL / "templates")
                errors, _ = review.check(path, self.root)
                flagged = {int(e.split(":")[1]) for e in errors if ": placeholder: " in e}
                lines = path.read_text().splitlines()
                for i, line in enumerate(lines, 1):
                    for m in re.finditer(r"<[^<>]+>", line):
                        self.assertIn(i, flagged, (mode, i, m.group(0)))
                path.unlink()


class Roadmap(Base):
    TEXT = ISSUE.replace("[[MAIN-93]], [[MAIN-131\\|#131]] |", "[[MAIN-93]], [[MAIN-M4]] |")

    def test_milestone_location(self):
        (self.root / ".github/roadmap").mkdir()
        (self.root / ".github/roadmap/MAIN-M4.md").write_text("milestone\n")
        self.assertEqual(self.run_check(self.TEXT)[0], [])

    def test_milestone_location_without_file(self):
        self.assertRule(self.TEXT, "findings")


class New(Base):
    def test_new_fails_only_for_placeholders(self):
        templates = self.root / "templates"
        templates.mkdir()
        (templates / "code.md").write_text(TEMPLATE)
        now = datetime.datetime(2026, 11, 25, 10, 15, tzinfo=datetime.timezone.utc)
        path = review.new("code", "claude-opus-5.5-medium", self.root, now, templates)
        self.assertEqual(path, self.root / ".github/reports" / NAME)
        errors, _ = review.check(path, self.root)
        self.assertTrue(errors)
        self.assertEqual({e.split(": ")[1] for e in errors}, {"placeholder"}, errors)
        with self.assertRaises(FileExistsError):
            review.new("code", "m", self.root, now, templates)


class Ignored(Base):
    """`new security` writes only to a folder that Git ignores."""

    NOW = datetime.datetime(2026, 11, 25, 10, 15, tzinfo=datetime.timezone.utc)

    def new(self):
        return review.new("security", "m", self.root, self.NOW, SKILL / "templates")

    def test_ignored(self):
        for ignore in (".agents\n", ".agents/\n", "/.agents/reports/\n", "*.md\n"):
            with self.subTest(ignore=ignore):
                git_init(self.root, ignore)
                path = self.new()
                self.assertEqual(path, self.root / ".agents/reports" / NAME)
                path.unlink()

    def test_not_ignored(self):
        git_init(self.root, "/dist/\n")
        with self.assertRaises(review.Refused) as e:
            self.new()
        self.assertIn(".agents/reports/%s is not ignored" % NAME, str(e.exception))
        self.assertFalse((self.root / ".agents").exists())

    def test_negated_report_name(self):
        git_init(self.root, "*.md\n!2026*.md\n")
        with self.assertRaises(review.Refused):
            self.new()

    def test_outside_git(self):
        shutil.rmtree(self.root / ".git")
        with self.assertRaises(review.Refused) as e:
            self.new()
        self.assertIn("not in a Git repository", str(e.exception))
        self.assertFalse((self.root / ".agents").exists())

    def test_main_refuses(self):
        git_init(self.root, "")
        err = io.StringIO()
        with contextlib.redirect_stderr(err):
            code = review.main(["new", "security", "--model", "m", "--root", str(self.root)])
        self.assertEqual(code, 1)
        self.assertIn("is not ignored by Git", err.getvalue())

    def test_check_not_ignored(self):
        git_init(self.root, "/dist/\n")
        errors, out = self.run_check(SECURITY, ".agents/reports")
        self.assertIn("privacy: .agents/reports/%s is not ignored by Git" % NAME, "\n".join(errors))
        self.assertEqual(len(errors), 1, errors)
        self.assertEqual(out[-1], review.PRIVATE_LINE)

    def test_check_outside_git(self):
        shutil.rmtree(self.root / ".git")
        self.assertRule(CODE, "privacy", ".agents/reports")
        self.assertEqual(self.run_check(CODE)[0], [])

    def test_status_not_ignored(self):
        git_init(self.root, "")
        self.run_check(SECURITY, ".agents/reports")
        with contextlib.redirect_stdout(io.StringIO()) as out:
            review.main(["status", "--root", str(self.root)])
        self.assertEqual(out.getvalue().splitlines()[0], ".agents/reports/%s: private, 0 of 1 findings closed." % NAME)
        self.assertIn("the report has 1 other errors; run check", out.getvalue())

    def test_code_needs_no_git(self):
        self.assertTrue(review.new("code", "m", self.root, self.NOW, SKILL / "templates").is_file())


class Prefix(Base):
    """The issue codes come from the file names of the mirror."""

    def setUp(self):
        super().setUp()
        issues = self.root / ".github/issues"
        for path in issues.glob("*.md"):
            path.rename(issues / path.name.replace("MAIN-", "ACME-"))

    def test_detected_prefix(self):
        text = CODE.replace("[[MAIN-", "[[ACME-")
        self.assertEqual(self.run_check(text)[0], [])
        self.assertEqual(self.run_check(ISSUE.replace("[[MAIN-", "[[ACME-"))[0], [])

    def test_other_prefix(self):
        errors, _ = self.run_check(CODE.replace("[[MAIN-93]]", "[[ACME-93]]"))
        self.assertIn("Tracked must be — or one issue link: [[ACME-N]]", "\n".join(errors))

    def test_roadmap_location(self):
        (self.root / ".github/roadmap").mkdir()
        (self.root / ".github/roadmap/ACME-M4.md").write_text("milestone\n")
        text = ISSUE.replace("[[MAIN-93]], [[MAIN-131\\|#131]] |", "[[ACME-93]], [[ACME-M4]] |")
        self.assertEqual(self.run_check(text.replace("[[MAIN-", "[[ACME-"))[0], [])

    def test_publication(self):
        (self.root / ".github/issues/ACME-200.md").write_text(
            "---\nnumber: 200\nstate: CLOSED\nstateReason: COMPLETED\n---\n\n%s, F1.\n" % SOURCE)
        path = self.root / ".agents/reports" / NAME
        path.parent.mkdir(parents=True)
        path.write_text(SECURITY)
        self.assertEqual(review.status(self.root)[0], [
            ".agents/reports/%s: publishable, 1 of 1 findings closed." % NAME,
            review.publish_line(NAME, self.root.resolve())])

    def test_any_prefix(self):
        for name in ("acme-1.md", "TEAM-API-1.md", "X_Y-1.md"):
            with self.subTest(name=name):
                folder = self.root / name
                (folder / ".github/issues").mkdir(parents=True)
                (folder / ".github/issues" / name).write_text("---\nstate: OPEN\n---\n")
                issues = review.project_issues(folder)
                self.assertIsInstance(issues, review.Mirror)
                link = "[[%s]]" % name[:-len(".md")]
                self.assertRegex(link, "^%s$" % issues.link)
                self.assertIsNone(issues.missing(name[:-len(".md")]))

    def test_mixed_names(self):
        (self.root / ".github/issues/README.md").write_text("notes\n")
        self.assertIsInstance(review.project_issues(self.root), review.Mirror)


FAKE_GH = """#!%s
import json, os, sys, time
data = json.load(open(os.environ["FAKE_GH"]))
open(os.environ["FAKE_GH"] + ".log", "a").write(" ".join(sys.argv[1:]) + "\\n")
time.sleep(data.get("sleep") or 0)
if data.get("fail"):
    sys.stderr.write(data["fail"] + "\\n")
    sys.exit(4)
args = sys.argv[1:]
repo = args[args.index("-R") + 1] if "-R" in args else ""
if args[:2] == ["issue", "view"]:
    issue = data["issues"].get(repo + "#" + args[2])
    if issue is None:
        sys.stderr.write("GraphQL: Could not resolve to an issue or pull request with the number of %%s.\\n" %% args[2])
        sys.exit(1)
    print(json.dumps({"state": issue["state"], "stateReason": issue.get("stateReason", "")}))
elif args[:2] == ["issue", "list"]:
    found = []
    for key, issue in data["issues"].items():
        comments = [{"body": c} for c in issue.get("comments", [])]
        text = issue.get("body", "") + "".join(c["body"] for c in comments)
        if key.startswith("#"):
            found.append({"number": int(key[1:]), "state": issue["state"],
                          "stateReason": issue.get("stateReason", ""), "body": issue.get("body", ""),
                          "comments": comments})
    print(json.dumps(found))
else:
    sys.exit(2)
""" % sys.executable


class NoMirror(Rule):
    """Without a mirror, Tracked takes a GitHub issue, the plain check stays
    offline, and the publication rule reads GitHub through gh."""

    TRACKED_GH = TRACKED.replace("[[MAIN-131]]", "#131")

    def setUp(self):
        super().setUp()
        for path in (self.root / ".github/issues").glob("*.md"):
            path.unlink()
        self.bin = self.root / "bin"
        self.bin.mkdir()
        self.data = self.root / "gh.json"
        self.gh({})
        git = shutil.which("git")
        if git:
            (self.bin / "git").symlink_to(git)
        self.env = mock.patch.dict(os.environ, {"PATH": str(self.bin), "FAKE_GH": str(self.data)})
        self.env.start()
        self.addCleanup(self.env.stop)

    def gh(self, issues, fail=None, installed=True, sleep=0):
        self.data.write_text(json.dumps({"issues": issues, "fail": fail, "sleep": sleep}))
        script = self.bin / "gh"
        if installed:
            script.write_text(FAKE_GH)
            script.chmod(0o755)
        elif script.exists():
            script.unlink()

    def calls(self):
        log = Path(str(self.data) + ".log")
        return log.read_text().splitlines() if log.exists() else []

    def closed(self, body="", reason="COMPLETED"):
        return {"state": "CLOSED", "stateReason": reason, "body": body}

    def test_project_issues(self):
        self.assertIsInstance(review.project_issues(self.root), review.GitHub)

    def test_github_links(self):
        for cell in ("#131", "octo/repo#131"):
            with self.subTest(cell=cell):
                errors, out = self.run_check(CODE.replace("[[MAIN-131]]", cell))
                self.assertEqual(errors, [])
        self.assertEqual(self.calls(), [])

    def test_wiki_link_refused(self):
        errors, _ = self.run_check(CODE)
        self.assertIn("Tracked must be — or one issue link: #N or owner/repo#N", "\n".join(errors))

    def test_issue_location(self):
        text = ISSUE.replace("[[MAIN-93]], [[MAIN-131\\|#131]]", "#93, octo/repo#131") \
            .replace("[[MAIN-131]] into [[MAIN-93]]", "#131 into #93")
        self.assertEqual(self.run_check(text)[0], [])

    def test_plain_check_offline(self):
        self.gh({"#200": self.closed(SOURCE + ", F1.")})
        errors, out = review.check(self.put(), self.root)
        self.assertEqual(errors, [])
        self.assertEqual(out[4], review.PRIVATE_LINE)
        code, out = self.run_main("check", str(self.put(EMPTY, ".github/reports", "20261126T101500Z.md")))
        self.assertEqual(code, 1)  # the timestamp differs from the name; only the call count matters
        self.assertEqual(self.calls(), [])

    def test_closed_issue(self):
        self.gh({"#200": self.closed(SOURCE + ", F1.")})
        self.put()
        self.assertEqual(self.run_main("status"), (0, [
            ".agents/reports/%s: publishable, 1 of 1 findings closed." % NAME,
            review.publish_line(NAME, self.root.resolve())]))
        code, out = self.run_main("check", "--published", str(self.root / ".agents/reports" / NAME))
        self.assertEqual(code, 0, out)

    def test_open_issue(self):
        self.gh({"#200": {"state": "OPEN", "stateReason": "", "body": SOURCE + ", F1."}})
        self.put()
        self.assertEqual(self.run_main("status"), (0, [
            ".agents/reports/%s: private, 0 of 1 findings closed." % NAME,
            "#200:1: publication: #200 holds F1 and is not closed as completed"]))

    def test_not_planned(self):
        self.gh({"#200": self.closed(SOURCE + ", F1.", "NOT_PLANNED")})
        self.assertEqual(self.blockers(), ["#200 holds F1 and is not closed as completed"])

    def test_url_source(self):
        url = SOURCE.replace("../reports/", "https://github.com/octo/repo/blob/release/v1/.github/reports/")
        self.gh({"#200": self.closed(url + ", F1.")})
        self.assertEqual(self.blockers(), [])

    def test_one_listing(self):
        self.gh({"#131": self.closed(), "#200": self.closed(SOURCE + ", F1.")})
        self.put()
        self.put(CODE.replace("[[MAIN-131]]", "#131"), ".github/reports", "20261124T101500Z.md")
        self.run_main("status")
        self.assertEqual(sum(1 for c in self.calls() if c.startswith("issue list")), 1, self.calls())

    def test_source_in_comment(self):
        self.gh({"#200": dict(self.closed("A fix."), comments=["Done.\n\n" + SOURCE + ", F1."])})
        self.assertEqual(self.blockers(), [])

    def test_tracked_issue(self):
        self.gh({"#131": {"state": "OPEN", "stateReason": ""}})
        self.assertEqual(self.blockers(self.TRACKED_GH), ["#131 holds F1 and is not closed as completed"])
        self.gh({"#131": self.closed()})
        self.assertEqual(self.blockers(self.TRACKED_GH), [])

    def test_tracked_other_repository(self):
        text = self.TRACKED_GH.replace("| #131 |", "| octo/repo#131 |")
        self.gh({"octo/repo#131": self.closed()})
        self.assertEqual(self.blockers(text), [])
        self.assertIn("issue view 131 -R octo/repo --json state,stateReason", self.calls())

    def test_missing_gh(self):
        self.gh({}, installed=False)
        self.assertEqual(self.blockers(), ["cannot read the issues: gh is not on PATH"])
        (self.root / ".github/reports" / NAME).unlink()
        self.put()
        code, out = self.run_main("status")
        self.assertEqual(out[0], ".agents/reports/%s: private, 0 of 1 findings closed." % NAME)

    def test_gh_error(self):
        self.gh({"#200": self.closed(SOURCE + ", F1.")},
                fail="To get started with GitHub CLI, please run:  gh auth login")
        self.assertEqual(self.blockers(), [
            "cannot read the issues: gh exits 4: To get started with GitHub CLI, please run:  gh auth login"])

    def test_unreadable_issue(self):
        self.gh({})
        got = self.blockers(self.TRACKED_GH)
        self.assertEqual(len(got), 1, got)
        self.assertTrue(got[0].startswith("cannot read the issues: gh exits 1: GraphQL: Could not resolve"), got)

    def test_timeout(self):
        self.gh({"#200": self.closed(SOURCE + ", F1.")}, sleep=5)
        with mock.patch.object(review, "GH_TIMEOUT", 0.5):
            self.assertEqual(self.blockers(), ["cannot read the issues: gh does not answer in 0.5 seconds"])

    def test_failure_once_per_run(self):
        self.gh({}, fail="gh: no network")
        self.put()
        self.put(name="20261124T101500Z.md")
        self.put(CODE.replace("[[MAIN-131]]", "#131"), ".github/reports", "20261123T101500Z.md")
        code, out = self.run_main("status")
        self.assertEqual(code, 1, out)
        self.assertEqual(len(self.calls()), 1, self.calls())

    def test_too_many_issues(self):
        self.gh({"#200": self.closed(SOURCE + ", F1."), "#201": self.closed()})
        with mock.patch.object(review, "GH_LIMIT", 2):
            self.assertEqual(self.blockers(), ["cannot read the issues: gh returns no list of issues, or 2 issues and more"])

    def test_bad_json(self):
        (self.bin / "gh").write_text("#!/bin/sh\necho not json\n")
        self.assertEqual(self.blockers(), ["cannot read the issues: gh returns no JSON"])

    def test_status_public_reports(self):
        self.put(CODE.replace("[[MAIN-131]]", "#131"), ".github/reports")
        self.gh({})
        self.assertEqual(self.run_main("status"), (0, ["No private report and no published security report."]))
        self.gh({}, fail="gh: no network")
        code, out = self.run_main("status")
        self.assertEqual((code, out), (1, [
            ".github/reports/%s:1: publication: cannot read the issues: gh exits 4: gh: no network" % NAME]))

    def test_status_named_public_report(self):
        self.put(CODE.replace("[[MAIN-131]]", "#131"), ".github/reports")
        self.gh({"#131": self.closed(), "#200": self.closed(SOURCE + ", F1, F3.")})
        self.assertEqual(self.run_main("status"),
                         (0, [".github/reports/%s: published, 3 of 3 findings closed." % NAME]))


FAKE_GH_API = """#!%s
import json, os, sys
data = json.load(open(os.environ["FAKE_GH"]))
open(os.environ["FAKE_GH"] + ".log", "a").write(" ".join(sys.argv[1:]) + "\\n")
args = sys.argv[1:]
if args[0] != "api":
    sys.exit(2)
answer = data.get(args[1], {"status": 404, "message": "Not Found"})
if "status" in answer:
    print(json.dumps({"message": answer["message"], "status": str(answer["status"])}))
    sys.stderr.write("gh: %%s (HTTP %%d)\\n" %% (answer["message"], answer["status"]))
    sys.exit(1)
if "--jq" in args:
    print(answer["jq"])
else:
    sys.stdout.write("".join(json.dumps(page) for page in answer["pages"]))
""" % sys.executable

REPO = "repos/octo/repo/"
DEPENDABOT = REPO + "dependabot/alerts?state=open,dismissed&per_page=100"
CODE_SCANNING = REPO + "code-scanning/alerts?state=open&per_page=100"
SECRET_SCANNING = REPO + "secret-scanning/alerts?state=open&per_page=100"
ADVISORIES = REPO + "security-advisories?per_page=100"


def dependabot_alert(n, state, severity="high", reason=None, comment=None):
    return {"number": n, "state": state, "dismissed_reason": reason, "dismissed_comment": comment,
            "security_advisory": {"severity": severity},
            "dependency": {"package": {"ecosystem": "npm", "name": "zod"}, "manifest_path": "docs/package-lock.json"}}


class Alerts(unittest.TestCase):
    """`alerts` reads the security alerts of the repository with `gh api`."""

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.bin = self.root / "bin"
        self.bin.mkdir()
        self.data = self.root / "gh.json"
        (self.bin / "gh").write_text(FAKE_GH_API)
        (self.bin / "gh").chmod(0o755)
        env = mock.patch.dict(os.environ, {"PATH": str(self.bin), "FAKE_GH": str(self.data)})
        env.start()
        self.addCleanup(env.stop)

    def answers(self, **changes):
        answers = {
            "repos/{owner}/{repo}": {"jq": "octo/repo"},
            "repos/octo/repo": {"jq": "true"},
            DEPENDABOT: {"pages": [
                [dependabot_alert(97, "dismissed", reason="tolerable_risk", comment="Docs site only.\nNot shipped."),
                 dependabot_alert(98, "open", "medium")],
                [dependabot_alert(99, "dismissed"), dependabot_alert(12, "fixed")]]},
            CODE_SCANNING: {"pages": [[{"number": 3, "state": "open",
                                        "rule": {"id": "go/path-injection", "security_severity_level": "high"},
                                        "most_recent_instance": {"location": {"path": "cmd/x.go", "start_line": 7}}}]]},
            SECRET_SCANNING: {"pages": [[{"number": 1, "state": "open",
                                          "secret_type_display_name": "GitHub Personal Access Token"}]]},
            ADVISORIES: {"pages": [[{"ghsa_id": "GHSA-aaaa-bbbb-cccc", "state": "draft", "severity": "critical",
                                     "vulnerabilities": [{"package": {"ecosystem": "go", "name": "go.acme/tool"}}]},
                                    {"ghsa_id": "GHSA-dddd-eeee-ffff", "state": "published", "severity": "low"}]]},
        }
        answers.update(changes)
        self.data.write_text(json.dumps(answers))

    def run_main(self, *argv):
        with contextlib.redirect_stdout(io.StringIO()) as out:
            code = review.main(["alerts", "--root", str(self.root)] + list(argv))
        return code, out.getvalue().splitlines()

    def calls(self):
        log = Path(str(self.data) + ".log")
        return log.read_text().splitlines() if log.exists() else []

    def test_lines(self):
        self.answers()
        self.assertEqual(self.run_main(), (0, [
            "Repository: octo/repo.",
            "dependabot #97: high, npm:zod, docs/package-lock.json, dismissed: tolerable_risk — Docs site only. Not shipped.",
            "dependabot #98: medium, npm:zod, docs/package-lock.json, open",
            "dependabot #99: high, npm:zod, docs/package-lock.json, dismissed: no reason — no comment",
            "dependabot: 1 open, 2 dismissed.",
            "code-scanning #3: high, go/path-injection, cmd/x.go:7, open",
            "code-scanning: 1 open.",
            "secret-scanning #1: —, GitHub Personal Access Token, —, open",
            "secret-scanning: 1 open.",
            "advisories GHSA-aaaa-bbbb-cccc: critical, go:go.acme/tool, —, draft",
            "advisories: 1 draft, 0 triage.",
        ]))

    def test_advisories_need_admin(self):
        self.answers(**{"repos/octo/repo": {"jq": "false"}})
        code, out = self.run_main()
        self.assertEqual(code, 1)
        self.assertIn("advisories: error: draft advisories need admin access to octo/repo", out)
        self.assertIn("dependabot: 1 open, 2 dismissed.", out)
        self.assertNotIn("api " + ADVISORIES, self.calls())

    def test_only_api_reads(self):
        self.answers()
        self.run_main()
        calls = self.calls()
        self.assertEqual(len(calls), 6, calls)
        for call in calls:
            self.assertTrue(call.startswith("api "), call)
            self.assertFalse(re.search(r"(^| )(-X|--method|-f|-F|--field|--raw-field|--input)( |$)", call), call)

    def test_off(self):
        self.answers(**{
            DEPENDABOT: {"status": 403, "message": "Dependabot alerts are disabled for this repository."},
            CODE_SCANNING: {"status": 404, "message": "no analysis found"},
            SECRET_SCANNING: {"status": 404, "message": "Secret scanning is disabled on this repository."}})
        code, out = self.run_main()
        self.assertEqual((code, out[1:4]), (0, [
            "dependabot: off — Dependabot alerts are disabled for this repository.",
            "code-scanning: off — no analysis found",
            "secret-scanning: off — Secret scanning is disabled on this repository."]))

    def test_no_access(self):
        self.answers(**{CODE_SCANNING: {"status": 403, "message": "You are not authorized to read code scanning alerts."},
                        SECRET_SCANNING: {"status": 404, "message": "Not Found"}})
        code, out = self.run_main()
        self.assertEqual(code, 1)
        self.assertIn("code-scanning: error: gh exits 1: gh: You are not authorized to read code scanning alerts. "
                      "(HTTP 403)", out)
        self.assertIn("secret-scanning: error: gh exits 1: gh: Not Found (HTTP 404)", out)
        self.assertIn("dependabot: 1 open, 2 dismissed.", out)

    def test_no_gh(self):
        self.answers()
        (self.bin / "gh").unlink()
        self.assertEqual(self.run_main(), (1, ["repository: error: gh is not on PATH"]))
        self.assertEqual(self.run_main("--repo", "octo/repo")[0], 1)

    def test_no_github_remote(self):
        self.answers(**{"repos/{owner}/{repo}": {"status": 404, "message": "Not Found"}})
        self.assertEqual(self.run_main(), (1, ["repository: error: gh exits 1: gh: Not Found (HTTP 404)"]))

    def test_repo_flag(self):
        self.answers()
        code, out = self.run_main("--repo", "octo/repo")
        self.assertEqual((code, out[0]), (0, "Repository: octo/repo."))
        self.assertNotIn("api repos/{owner}/{repo} --jq .full_name", self.calls())
        with contextlib.redirect_stderr(io.StringIO()), self.assertRaises(SystemExit):
            self.run_main("--repo", "octo/repo/../x")

    def test_bad_json(self):
        self.answers(**{ADVISORIES: {"pages": [{"message": "a dict"}]}})
        code, out = self.run_main()
        self.assertEqual((code, out[-1]), (1, "advisories: error: gh returns no JSON list"))

    def test_timeout(self):
        self.answers()
        (self.bin / "gh").write_text("#!%s\nimport time\ntime.sleep(5)\n" % sys.executable)
        with mock.patch.object(review, "GH_TIMEOUT", 0.5):
            self.assertEqual(self.run_main(), (1, ["repository: error: gh does not answer in 0.5 seconds"]))

if __name__ == "__main__":
    unittest.main()
