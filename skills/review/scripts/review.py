#!/usr/bin/env python3
"""Create and check review reports: `review.py new`, `check` and `status`.
Read the security alerts of the repository on GitHub: `review.py alerts`."""

import argparse
import datetime
import json
import os
import re
import shlex
import shutil
import subprocess
import sys
import uuid
from pathlib import Path

MODES = {
    "code": ("Code review", "files",
             ["Gates", "Correctness", "Data safety", "Contract", "Tests", "Simplicity"]),
    "issue": ("Issue review", "issues",
              ["Gates", "Valid", "Ready", "Single", "True", "Placed"]),
    "security": ("Security review", "files",
                 ["Gates", "Files", "Processes", "Secrets", "Service", "Supply chain"]),
}
SECTIONS = ["Verdict", "Scope", "Findings", "Coverage"]
SEVERITIES = ["Critical", "High", "Medium", "Low"]
VERDICTS = ["BLOCKED", "INCOMPLETE", "ACTION", "CLEAR"]
CHECKED = ["full", "partial", "none"]
SCOPE_FIELDS = ["Revision", "Target", "Reviewed", "Excluded", "Accepted"]
FINDING_COLUMNS = ["ID", "Severity", "Lens", "Location", "Finding", "Tracked"]
COVERAGE_COLUMNS = ["Lens", "Checked", "Findings", "Evidence"]
BLOCK_FIELDS = ["Failure", "Evidence", "Fix"]
EVIDENCE_PREFIXES = ("reproduced — ", "inspected — ")
FIX_VERBS = ["Close", "Rewrite", "Split", "Merge", "Reopen", "Untick", "Move", "Sync"]
BANNED = ["seems", "appears", "might", "could", "possibly", "probably", "perhaps",
          "very", "really", "quite", "extremely", "comprehensive", "robust", "crucial",
          "overall", "basically", "simply", "note that", "in order to", "it is worth"]
PRIVATE_MARKERS = ["/Users/", "/home/", ".github/notes"]
HTML_TAGS = {"a", "b", "i", "p", "br", "hr", "div", "span", "img", "code", "pre", "em",
             "strong", "sub", "sup", "kbd", "details", "summary", "table", "tr", "td", "th",
             "ul", "ol", "li", "script", "style"}
PRIVATE_LINE = "Private: do not publish this report before the fix ships."
PUBLIC, PRIVATE = ".github/reports", ".agents/reports"
ISSUES, ROADMAP = ".github/issues", ".github/roadmap"

ISSUE_FILE_RE = re.compile(r"^([A-Za-z][A-Za-z0-9_-]*)-\d+\.md$")
GH_LINK = r"((?:[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)?#\d+)"
GH_REF_RE = re.compile(r"^(?:([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+))?#(\d+)$")
GH_LIMIT = 10000
GH_TIMEOUT = 60
CODE_SPAN = re.compile(r"`[^`]*`")
NAME_RE = re.compile(r"^(\d{8}T\d{6})Z\.md$")
UUID4_RE = re.compile(r"^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
VERDICT_RE = re.compile(r"^\*\*([A-Z]+)\*\* — (.+\.)$")
REVISION_RE = re.compile(r"^`[0-9a-f]{7,12}` on `[^`]+`, (clean tree|dirty tree: \d+ files)$")
REVIEWED_RE = re.compile(r"^(\d+) of (\d+) (files|issues)$")
BLOCK_TITLE_RE = re.compile(r"^### (F\d+)\. (.*)$")
FIELD_RE = re.compile(r"^- \*\*([A-Za-z]+):\*\* ?(.*)$")
SEPARATOR_CELL_RE = re.compile(r"^:?-+:?$")
SOURCE_RE = re.compile(r"^Source: \[[^\]]+\]\((?:\.\./reports/|https://github\.com/[^)\s]+/\.github/reports/)"
                       r"([^)/]+)\)((?:, F\d+)+)\.$")
SECURITY_H1_RE = re.compile(r"^# *security review *$", re.I | re.M)
STATE_KEY_RE = re.compile(r"""^["']?(state|stateReason)["']?\s*:""")
STATE_LINE_RE = re.compile(r"^(?:state|stateReason): ([A-Za-z_]+)$")
SPAN_LIMIT = 90
HR_RE = re.compile(r"^ {0,3}([-*_])( *\1){2,} *$")
REPO_RE = re.compile(r"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$")
HTTP_RE = re.compile(r"\(HTTP (\d{3})\)")
# The message of a scanner that is off: GitHub answers 403 or 404 with it.
OFF_RE = re.compile(r"disabled|not enabled|must be enabled|no analysis found", re.I)


# Text helpers

def outside_code(text):
    """The text with each backtick span replaced by a space."""
    return CODE_SPAN.sub(" ", text)


def prose(text):
    """Authored prose: without backtick spans and double-quoted text."""
    return re.sub(r'"[^"]*"|“[^”]*”', " ", outside_code(text))


def tokens(text):
    """The text with each backtick span, wiki link and quote as one word."""
    return re.sub(r'`[^`]*`|\[\[[^\]]*\]\]|"[^"]*"|“[^”]*”', "W", text)


def word_count(text):
    return sum(1 for w in tokens(text).split() if any(c.isalnum() for c in w))


def sentences(text):
    parts = re.split(r"[.?!](?:\s+|$)", tokens(text).strip())
    return [p for p in parts if p.strip()]


def split_cells(line):
    cells, cell, code, prev = [], "", False, ""
    for ch in line.strip():
        if ch == "`":
            code = not code
        if ch == "|" and not code and prev != "\\":
            cells.append(cell.strip())
            cell = ""
        else:
            cell += ch
        prev = ch
    cells.append(cell.strip())
    return cells[1:-1] if len(cells) > 2 else cells


def table_lines(lines):
    """The first table in the lines: (line number, text) pairs."""
    rows = []
    for no, text in lines:
        if text.startswith("|"):
            rows.append((no, text))
        elif rows:
            break
    return rows


def table(lines):
    """The first table in the lines: header cells and (line number, cells) rows."""
    rows = [(no, split_cells(text)) for no, text in table_lines(lines)]
    if len(rows) < 2:
        return None, []
    return rows[0][1], rows[2:]


def content(lines):
    return [(no, t) for no, t in lines if t.strip()]


# Report parsing

def split_front_matter(lines):
    if not lines or lines[0][1] != "---":
        return None, lines
    for i in range(1, len(lines)):
        if lines[i][1] == "---":
            return lines[1:i], lines[i + 1:]
    return None, lines


def check_front_matter(fm, name, err):
    if fm is None:
        err(1, "frontmatter", "no front matter")
        return
    keys = {}
    for no, text in fm:
        key, _, value = text.partition(":")
        keys[key.strip()] = (no, value.strip())
    if list(keys) != ["uid", "timestamp", "model"] or len(fm) != 3:
        err(1, "frontmatter", "keys must be exactly uid, timestamp, model")
    no, uid = keys.get("uid", (1, ""))
    if not UUID4_RE.match(uid):
        err(no, "frontmatter", "uid is not a uuid4")
    no, ts = keys.get("timestamp", (1, ""))
    m = NAME_RE.match(name)
    if not m:
        err(1, "frontmatter", "file name is not <UTC>Z.md")
    else:
        stamp = datetime.datetime.strptime(m.group(1), "%Y%m%dT%H%M%S")
        if ts != stamp.strftime("%Y-%m-%dT%H:%M:%SZ"):
            err(no, "frontmatter", "timestamp does not equal the file name")
    no, model = keys.get("model", (1, ""))
    if not model:
        err(no, "frontmatter", "model is empty")


def check_structure(body, err):
    """Check the skeleton; return the mode and the lines of each H2 section."""
    mode, sections, current, h2 = None, {}, None, []
    for no, text in body:
        if text.startswith("```") or text.startswith("~~~"):
            err(no, "structure", "fenced code block")
        if "![" in outside_code(text):
            err(no, "structure", "image")
        if HR_RE.match(text):
            err(no, "structure", "horizontal rule")
        heading = re.match(r"^(#+) (.*)$", text)
        level = len(heading.group(1)) if heading else 0
        if level == 1:
            titles = {v[0]: k for k, v in MODES.items()}
            if mode or h2 or heading.group(2) not in titles:
                err(no, "structure", "H1 must be one of: " + ", ".join(titles))
            mode = mode or titles.get(heading.group(2))
        elif level == 2:
            current = heading.group(2)
            h2.append(current)
            sections[current] = []
        elif level == 3 and current != "Findings":
            err(no, "structure", "H3 outside Findings")
        elif level >= 4:
            err(no, "structure", "H4 or deeper heading")
        elif current and level != 2:
            sections[current].append((no, text))
        elif not current and text.strip() and level == 0:
            err(no, "structure", "text outside a section: only the H1 comes before ## Verdict")
    if h2 != SECTIONS:
        err(1, "structure", "H2 sections must be " + ", ".join(SECTIONS) + " in order")
    if mode is None:
        err(1, "structure", "no valid H1")
    return mode, sections


def check_tokens(body, err):
    token = re.compile(r"<([^<>\n]+)>")
    for no, text in body:
        if any(m in text for m in PRIVATE_MARKERS):
            err(no, "privacy", "home path or private note link")
        free = outside_code(text)
        if re.search(r"\bTODO\b", free):
            err(no, "placeholder", "TODO left")
        for m in token.finditer(free):
            name = re.match(r"/?([a-zA-Z][a-zA-Z0-9]*)", m.group(1))
            if name and name.group(1).lower() in HTML_TAGS:
                err(no, "structure", "HTML tag " + m.group(0))
            else:
                err(no, "placeholder", "placeholder " + m.group(0) + " left")
        # A code span is a leftover only if it holds nothing but placeholders:
        # `<path>:<line>` fails, `tool --config <path>` passes.
        for span in CODE_SPAN.findall(text):
            if len(span) - 2 > SPAN_LIMIT:
                err(no, "budget", "backtick span above %d characters" % SPAN_LIMIT)
            names = token.findall(span)
            if names and all(re.fullmatch(r"[a-z ]+", n) for n in names) \
                    and not re.search(r"[A-Za-z0-9]", token.sub("", span)):
                err(no, "placeholder", "placeholder " + span + " left")


def check_shape(section, lines, lead, rows, err):
    """Allow `lead` lines, then one table: header, separator and `rows` rows, nothing else."""
    body = content(lines)
    for i, (no, text) in enumerate(body):
        in_table = lead <= i < lead + 2 + rows and text.startswith("|") and (
            i == lead or body[i - 1][0] == no - 1)
        if i < lead or in_table:
            continue
        err(no, "structure", "%s: the section holds only its fixed lines and one table" % section)


def check_table(section, lines, columns, err):
    """One table: pipes at both ends, a separator second, equal cell counts, the header."""
    rows = table_lines(lines)
    if not rows:
        return
    for no, text in rows:
        if not text.rstrip().endswith("|"):
            err(no, "structure", "%s: text after the closing pipe" % section)
        if len(split_cells(text)) != len(columns):
            err(no, "structure", "%s: a table line must have %d cells" % (section, len(columns)))
    if split_cells(rows[0][1]) != columns:
        err(rows[0][0], "structure", "%s: the header must be %s" % (section, " | ".join(columns)))
    if len(rows) < 2 or not all(SEPARATOR_CELL_RE.match(c) for c in split_cells(rows[1][1])):
        err(rows[min(1, len(rows) - 1)][0], "structure", "%s: the second table line must be a separator" % section)


def check_language(no, text, err):
    for s in sentences(text):
        if word_count(s) > 25:
            err(no, "language", "sentence above 25 words")
    clean = prose(text)
    if re.search(r"\bmay\b", clean):
        err(no, "language", "banned word 'may'")
    for word in BANNED:
        if re.search(r"\b" + word.replace(" ", r"\s+") + r"\b", clean, re.I):
            err(no, "language", "banned word '" + word + "'")


# Sections

def parse_findings(lines, mode, issues, err, published=False):
    body = content(lines)
    if body and body[0][1] == "None.":
        if len(body) > 1:
            err(body[1][0], "findings", "None. must be the whole section")
        return []
    header, rows = table(lines)
    if not body or not body[0][1].startswith("|") or header != FINDING_COLUMNS:
        err(body[0][0] if body else 1, "findings", "table must have columns " + ", ".join(FINDING_COLUMNS))
        return []
    lenses = MODES[mode][2]
    loc = issues.location if mode == "issue" else r"`[^`\s]+:\d+`"
    loc_re = re.compile(r"^%s(, %s){0,2}( \+\d+ more)?$" % (loc, loc))
    findings, prev = [], None
    for i, (no, cells) in enumerate(rows, 1):
        if len(cells) != 6:
            err(no, "findings", "row must have 6 cells")
            continue
        fid, sev, lens, where, text, tracked = cells
        if fid != "F%d" % i:
            err(no, "findings", "ID must be F%d" % i)
        if sev not in SEVERITIES:
            err(no, "findings", "unknown severity " + sev)
        if lens not in lenses:
            err(no, "findings", "unknown lens " + lens)
        if not loc_re.match(where):
            err(no, "findings", "location format of the mode")
        for key in re.findall(issues.location, where) if mode == "issue" else []:
            missing = issues.missing(key)
            if missing:
                err(no, "findings", "no file " + missing)
        if not text:
            err(no, "findings", "empty Finding cell")
        if word_count(text) > 15 or text.endswith("."):
            err(no, "findings", "finding text: at most 15 words, no period at the end")
        link = re.fullmatch(issues.link, tracked)
        if tracked != "—" and not link:
            err(no, "tracked", "Tracked must be — or one issue link: " + issues.form)
        elif link:
            if issues.missing(link.group(1)):
                err(no, "tracked", "no issue file for " + tracked)
            elif not published and issues.offline_state(link.group(1)) not in (None, "OPEN"):
                err(no, "tracked", "the issue is not open: " + tracked)
        key = (SEVERITIES.index(sev) if sev in SEVERITIES else 9,
               lenses.index(lens) if lens in lenses else 9, where)
        if prev and key < prev:
            err(no, "order", "rows must sort by severity, lens order, location")
        prev = key
        check_language(no, text, err)
        findings.append({"no": no, "id": fid, "sev": sev, "lens": lens, "text": text,
                         "tracked": tracked != "—",
                         "issue": link.group(1) if link else None})
    return findings


def issue_state(issue):
    """The state and the state reason of an issue file, and the line of its
    state key. The state is "repeated" if the front matter holds one of the
    two keys more than once, in any spelling of the key. A value counts only
    on a line of the form `key: VALUE`; another form gives an empty value."""
    fm, _ = split_front_matter(list(enumerate(issue.read_text(encoding="utf-8").splitlines(), 1)))
    keys = {"state": [], "stateReason": []}
    for no, text in fm or []:
        m = STATE_KEY_RE.match(text)
        if m:
            value = STATE_LINE_RE.match(text)
            keys[m.group(1)].append((no, value.group(1) if value else ""))
    for found in keys.values():
        if len(found) > 1:
            return "repeated", "", found[1][0]
    no, state = (keys["state"] or [(1, "")])[0]
    return state, (keys["stateReason"] or [(1, "")])[0][1], no


class GhError(Exception):
    """gh is absent, not signed in, or did not return an issue."""


class Mirror:
    """The issues of the project from its mirror: `.github/issues/<PREFIX>-<n>.md`.
    A key is the stem of an issue file, such as `ACME-131`."""

    def __init__(self, root, prefixes):
        self.root = Path(root)
        alt = "(?:%s)" % "|".join(re.escape(p) for p in prefixes)
        self.link = r"\[\[(%s-\d+)(?:\\?\|[^\]]+)?\]\]" % alt
        self.location = r"\[\[(%s-(?:\d+|[MP]\d+))(?:\\?\|[^\]]+)?\]\]" % alt
        self.form = "[[%s-N]]" % prefixes[0]

    def file(self, key):
        folder = ISSUES if re.search(r"-\d+$", key) else ROADMAP
        return "%s/%s.md" % (folder, key)

    def missing(self, key):
        """The path of the file of the key, if the file does not exist."""
        path = self.file(key)
        return None if (self.root / path).is_file() else path

    def offline_state(self, key):
        return issue_state(self.root / self.file(key))[0]

    def state(self, key):
        """(state, reason, label, line) of the issue; state None if no file."""
        label = self.file(key)
        if not (self.root / label).is_file():
            return None, "", label, 1
        state, reason, no = issue_state(self.root / label)
        return state, reason, label, no

    def naming(self, stem):
        """(key, label, hits) for each issue whose text holds stem."""
        for issue in sorted((self.root / ISSUES).glob("*.md")):
            hits = [(no, text) for no, text in
                    enumerate(issue.read_text(encoding="utf-8").splitlines(), 1) if stem in text]
            if hits:
                yield issue.stem, "%s/%s" % (ISSUES, issue.name), hits

    def sources(self):
        """The file names of the reports that an issue names by finding ID."""
        names = set()
        for issue in sorted((self.root / ISSUES).glob("*.md")):
            for text in issue.read_text(encoding="utf-8").splitlines():
                m = SOURCE_RE.match(text)
                if m:
                    names.add(m.group(1))
        return names


class GitHub:
    """The issues of the project from GitHub through `gh`, for a project
    without a mirror. A key is `#N` or `owner/repo#N`. Each failure of `gh`
    raises GhError: the caller keeps the report private."""

    link = location = GH_LINK
    form = "#N or owner/repo#N"

    def __init__(self, root):
        self.root = Path(root)
        self.states, self.listed, self.failure = {}, None, None

    def missing(self, key):
        return None

    def offline_state(self, key):
        return None

    def gh(self, *args):
        # A failure that holds for each call holds for the whole run.
        if self.failure:
            raise self.failure
        if not shutil.which("gh"):
            self.failure = GhError("gh is not on PATH")
            raise self.failure
        try:
            done = subprocess.run(["gh"] + list(args), cwd=self.root, capture_output=True, text=True,
                                  timeout=GH_TIMEOUT)
        except subprocess.TimeoutExpired:
            self.failure = GhError("gh does not answer in %g seconds" % GH_TIMEOUT)
            raise self.failure
        except OSError as e:
            self.failure = GhError("gh does not run: %s" % e)
            raise self.failure
        if done.returncode:
            lines = [t.strip() for t in done.stderr.splitlines() if t.strip()]
            raise GhError("gh exits %d: %s" % (done.returncode, lines[0] if lines else "no message"))
        try:
            return json.loads(done.stdout)
        except ValueError:
            raise GhError("gh returns no JSON")

    def issues(self):
        """Each issue of the repository with its text: the body, then each
        comment. The search of GitHub does not find a report name inside a
        URL, so the list holds each issue."""
        if self.listed is not None:
            return self.listed
        try:
            return self.list_issues()
        except GhError as e:
            self.failure = self.failure or e
            raise

    def list_issues(self):
        found = self.gh("issue", "list", "--state", "all", "--limit", str(GH_LIMIT),
                        "--json", "number,state,stateReason,body,comments")
        if not isinstance(found, list) or len(found) >= GH_LIMIT:
            raise GhError("gh returns no list of issues, or %d issues and more" % GH_LIMIT)
        issues = []
        for item in found:
            try:
                key = "#%d" % item["number"]
                self.states[key] = (item["state"], item.get("stateReason") or "")
                text = [item.get("body") or ""] + [c.get("body") or "" for c in item.get("comments") or []]
            except (KeyError, TypeError, AttributeError):
                raise GhError("gh returns an issue without number, state or text")
            issues.append((key, "\n".join(text).splitlines()))
        self.listed = issues
        return issues

    def state(self, key):
        if key not in self.states:
            m = GH_REF_RE.match(key)
            repo = ["-R", m.group(1)] if m.group(1) else []
            got = self.gh("issue", "view", m.group(2), *repo, "--json", "state,stateReason")
            if not isinstance(got, dict) or "state" not in got:
                raise GhError("gh returns no state for " + key)
            self.states[key] = (got["state"], got.get("stateReason") or "")
        state, reason = self.states[key]
        return state, reason, key, 1

    def naming(self, stem):
        for key, lines in self.issues():
            hits = [(no, text) for no, text in enumerate(lines, 1) if stem in text]
            if hits:
                yield key, key, hits

    def sources(self):
        return {m.group(1) for _, lines in self.issues() for m in map(SOURCE_RE.match, lines) if m}


def project_issues(root):
    """The issues of the project: the mirror if `.github/issues/` holds a
    `<PREFIX>-<n>.md` file, otherwise GitHub."""
    names = sorted((Path(root) / ISSUES).glob("*.md"))
    prefixes = sorted({m.group(1) for m in (ISSUE_FILE_RE.match(p.name) for p in names) if m})
    return Mirror(root, prefixes) if prefixes else GitHub(root)


def publication(name, findings, issues, add):
    """Apply the publication rule to the report with this file name. Call
    add(label, line, message) for each blocker; label None is the report.
    Return the number of closed findings. A failure to read the issues is
    one blocker, and no finding is closed."""
    try:
        return rule(name, findings, issues, add)
    except GhError as e:
        add(None, 1, "cannot read the issues: %s" % e)
        return 0


def rule(name, findings, issues, add):
    stem = name[:-len(".md")]
    holders = {f["id"]: set() for f in findings}
    for f in findings:
        if f["issue"]:
            holders[f["id"]].add(f["issue"])
    related = {f["issue"] for f in findings if f["issue"]}
    for key, label, hits in issues.naming(stem):
        related.add(key)
        sources = [(no, text) for no, text in hits if text.startswith("Source:")]
        if not sources:
            add(label, hits[0][0], "the issue names the report without a Source line")
        for no, text in sources:
            m = SOURCE_RE.match(text)
            if not m or m.group(1) != name:
                add(label, no, "the line must be: Source: [<text>](../reports/%s), F<n>." % name)
                continue
            for fid in m.group(2).split(", ")[1:]:
                if fid in holders:
                    holders[fid].add(key)
                else:
                    add(label, no, "the report has no " + fid)
    done, blockers = {}, []
    for key in sorted(related):
        state, reason, label, no = issues.state(key)
        done[key] = state == "CLOSED" and reason == "COMPLETED"
        held = [fid for fid in holders if key in holders[fid]]
        if state == "repeated":
            blockers.append((label, no, "%s repeats a state key" % key))
        elif state is not None and not done[key]:
            blockers.append((label, no, "%s %sis not closed as completed" % (
                key, "holds %s and " % ", ".join(held) if held else "")))
    for blocker in blockers:
        add(*blocker)
    closed = 0
    for f in findings:
        if not holders[f["id"]]:
            add(None, f["no"], f["id"] + " has no issue")
        elif all(done[key] for key in holders[f["id"]]):
            closed += 1
    return closed


def check_blocks(lines, mode, findings, err):
    blocks, current, body = [], None, content(lines)
    start = 1 if body and body[0][1] == "None." else 0
    first = start  # the one table: contiguous lines, no blank line inside
    while start < len(body) and body[start][1].startswith("|") and (
            start == first or body[start][0] == body[start - 1][0] + 1):
        start += 1
    for no, text in body[start:]:
        m = BLOCK_TITLE_RE.match(text)
        if text.startswith("### "):
            current = {"no": no, "id": m.group(1) if m else "", "title": m.group(2) if m else "",
                       "fields": []}
            blocks.append(current)
        elif current:
            f = FIELD_RE.match(text)
            if f:
                current["fields"].append((no, f.group(1), f.group(2).strip()))
            else:
                err(no, "findings", "a line that is not a block field")
        else:
            err(no, "findings", "Findings holds one table, then the blocks")
    rows = {f["id"]: f for f in findings}
    expected = [f["id"] for f in findings if not f["tracked"]]
    for b in blocks:
        row = rows.get(b["id"])
        if row is None or row["tracked"]:
            err(b["no"], "block", "no block for a tracked or missing row: " + (b["id"] or "?"))
    got = [b["id"] for b in blocks if b["id"] in expected]
    for fid in expected:
        if fid not in got:
            err(rows[fid]["no"], "block", fid + " has no block")
    if got != [i for i in expected if i in got]:
        err(blocks[0]["no"], "block", "blocks must follow the table order")
    for b in blocks:
        row = rows.get(b["id"])
        if row is None:
            continue
        if b["title"] != row["text"]:
            err(b["no"], "block", "title must equal the Finding cell")
        if [n for _, n, _ in b["fields"]] != BLOCK_FIELDS:
            err(b["no"], "block", "fields must be Failure, Evidence, Fix")
        for no, name, value in b["fields"]:
            check_field(no, name, value, mode, row["sev"] == "Low", err)


def check_field(no, name, value, mode, low, err):
    if not value:
        err(no, "block", name + " is empty")
        return
    if name == "Evidence" and not value.startswith(EVIDENCE_PREFIXES):
        err(no, "block", "Evidence must start with 'reproduced — ' or 'inspected — '")
    elif name == "Evidence" and not CODE_SPAN.search(value):
        err(no, "block", "Evidence must cite a command or a path:line in backticks")
    if name == "Failure" and mode == "security" and not re.match(r"^A[1-4]: ", value):
        err(no, "block", "Failure must start with the attacker code A1: to A4:")
    if name == "Fix" and mode == "issue" and value.split()[0] not in FIX_VERBS:
        err(no, "block", "Fix must start with one of: " + ", ".join(FIX_VERBS))
    count, words = len(sentences(value)), word_count(value)
    if (low and (count > 1 or words > 25)) or count > 3 or words > 50:
        limit = "1 sentence, 25 words" if low else "3 sentences, 50 words"
        err(no, "budget", "%s above the budget of %s" % (name, limit))
    check_language(no, value, err)


def check_coverage(lines, mode, findings, err):
    header, rows = table(lines)
    if header != COVERAGE_COLUMNS:
        err(lines[0][0] if lines else 1, "coverage", "table must have columns " + ", ".join(COVERAGE_COLUMNS))
        return {}
    lenses, checked = MODES[mode][2], {}
    if [r[1][0] for r in rows] != lenses:
        err(rows[0][0] if rows else 1, "coverage", "rows must be the lenses " + ", ".join(lenses))
    for no, cells in rows:
        if len(cells) != 4:
            err(no, "coverage", "row must have 4 cells")
            continue
        lens, state, count, evidence = cells
        checked[lens] = state
        if state not in CHECKED:
            err(no, "coverage", "Checked must be full, partial or none")
        if count != str(sum(1 for f in findings if f["lens"] == lens)):
            err(no, "coverage", "Findings must equal the table count for " + lens)
        if not evidence or word_count(evidence) > 25:
            err(no, "coverage", "Evidence must be non-empty, at most 25 words")
        check_language(no, evidence, err)
    return checked


def check_scope(lines, mode, not_full, err):
    header, rows = table(lines)
    if header != ["Field", "Value"] or [r[1][0] for r in rows] != SCOPE_FIELDS:
        err(lines[0][0] if lines else 1, "scope", "rows must be " + ", ".join(SCOPE_FIELDS))
    for no, cells in rows:
        value = cells[1] if len(cells) == 2 else ""
        if not value:
            err(no, "scope", "empty value")
            continue
        check_language(no, value, err)
        if cells[0] == "Revision" and not REVISION_RE.match(value):
            err(no, "scope", "Revision must be `<sha>` on `<branch>`, clean tree or dirty tree: N files")
        if cells[0] != "Reviewed":
            continue
        m = REVIEWED_RE.match(value)
        if not m or m.group(3) != MODES[mode][1]:
            err(no, "scope", "Reviewed must be 'N of M %s'" % MODES[mode][1])
        elif int(m.group(1)) > int(m.group(2)):
            err(no, "scope", "Reviewed: N above M")
        elif int(m.group(1)) < int(m.group(2)) and not not_full:
            err(no, "scope", "N < M, but each lens is full")


def check_verdict(lines, findings, not_full, err):
    body = content(lines)
    if not body:
        err(1, "verdict", "no verdict line")
        return None
    no, line = body[0]
    counts = {s: sum(1 for f in findings if f["sev"] == s) for s in SEVERITIES}
    blocking = [f["id"] for f in findings if f["sev"] in ("Critical", "High")]
    if blocking:
        derived = "BLOCKED"
    elif not_full:
        derived = "INCOMPLETE"
    else:
        derived = "ACTION" if counts["Medium"] else "CLEAR"
    m = VERDICT_RE.match(line)
    if not m or m.group(1) not in VERDICTS:
        err(no, "verdict", "line must be **<VERDICT>** — <reason>.")
    else:
        reason = m.group(2)
        if m.group(1) != derived:
            err(no, "verdict", "verdict must be %s by the tables" % derived)
        for fid in blocking[:3] if m.group(1) == "BLOCKED" else []:
            if not re.search(r"\b%s\b" % fid, reason):
                err(no, "verdict", "the BLOCKED reason must name " + fid)
        more = "and %d more" % (len(blocking) - 3)
        if m.group(1) == "BLOCKED" and len(blocking) > 3 and more not in reason:
            err(no, "verdict", "the BLOCKED reason must hold '%s'" % more)
        for lens in not_full if derived == "INCOMPLETE" else []:
            if lens.lower() not in reason.lower():
                err(no, "verdict", "the reason must name the lens " + lens)
        if word_count(reason) > 25:
            err(no, "verdict", "reason above 25 words")
        check_language(no, reason, err)
    header, rows = table(lines)
    row = [str(counts[s]) for s in SEVERITIES]
    if header != SEVERITIES or len(rows) != 1 or rows[0][1] != row:
        err(rows[0][0] if rows else no, "counts", "counts row must be " + " | ".join(row))
    return derived


# Verbs

def check(path, root, published=False):
    """Return (errors, hand-over lines) for the report at path. published
    applies the publication rule: it accepts a security report in
    .github/reports/ only if each of its findings is closed. Without a
    mirror, only published reads GitHub."""
    return examine(path, root, published)[:2]


def examine(path, root, published=False, label=None, issues=None):
    """Return (errors, hand-over lines, closed findings, findings) for the
    report at path. label replaces the path in the error lines."""
    path = Path(path)
    label = label or str(path)
    errors = []

    def err(no, rule, message):
        errors.append("%s:%d: %s: %s" % (label, no, rule, message))

    lines = list(enumerate(path.read_text(encoding="utf-8").splitlines(), 1))
    fm, body = split_front_matter(lines)
    check_front_matter(fm, path.name, err)
    mode, sections = check_structure(body, err)
    check_tokens(body, err)
    root = Path(root).resolve()
    issues = issues or project_issues(root)
    folders = {root / PUBLIC: PUBLIC, root / PRIVATE: PRIVATE}
    folder = folders.get(path.resolve().parent, "")
    private = folder == PRIVATE
    if mode == "security" and not private and not (published and folder):
        err(1, "privacy", "a security report must be in .agents/reports/ of the root")
    elif not folder:
        err(1, "privacy", "a report must be in .github/reports/ or .agents/reports/ of the root")
    if private:
        reason = not_ignored(root, PRIVATE, path.name)
        if reason:
            err(1, "privacy", reason)
    twin = root / (PRIVATE if folder == PUBLIC else PUBLIC) / path.name
    if published and folder and twin.is_file() and twin.read_bytes() != path.read_bytes():
        err(1, "publication", "the report differs from its copy in %s/" % (PRIVATE if folder == PUBLIC else PUBLIC))
    if mode is None or any(s not in sections for s in SECTIONS):
        return errors, [], 0, 0
    for section, lead, rows in (("Verdict", 1, 1), ("Scope", 0, 5), ("Coverage", 0, 6)):
        check_shape(section, sections[section], lead, rows, err)
    tables = {"Verdict": SEVERITIES, "Scope": ["Field", "Value"], "Coverage": COVERAGE_COLUMNS}
    if [t for _, t in content(sections["Findings"])][:1] != ["None."]:
        tables["Findings"] = FINDING_COLUMNS
    for section, columns in tables.items():
        check_table(section, sections[section], columns, err)
    findings = parse_findings(sections["Findings"], mode, issues, err, published)
    check_blocks(sections["Findings"], mode, findings, err)
    checked = check_coverage(sections["Coverage"], mode, findings, err)
    not_full = [lens for lens in MODES[mode][2] if checked.get(lens, "full") != "full"]
    check_scope(sections["Scope"], mode, not_full, err)
    verdict = check_verdict(sections["Verdict"], findings, not_full, err)
    first = [f["id"] + " " + f["text"] for f in findings if f["sev"] in ("Critical", "High")][:3]
    counts = ", ".join("%s %d" % (s, sum(1 for f in findings if f["sev"] == s)) for s in SEVERITIES)
    out = ["%s: %s — %s." % (MODES[mode][0], verdict, counts),
           "Report: %s/%s" % (folder, path.name),
           "First: " + ("; ".join(first) if first else "none") + ".",
           "Coverage: " + ("partial: " + ", ".join(not_full) if not_full else "full") + "."]
    blockers = []

    def add(where, no, message):
        blockers.append("%s:%d: publication: %s" % (where or label, no, message))

    # Without a mirror, the rule needs GitHub: the plain check stays offline
    # and keeps a private report private.
    offline = isinstance(issues, GitHub) and not published
    closed = publication(path.name, findings, issues, add) if (published or private) and not offline else 0
    if published:
        errors.extend(blockers)
    if private and not published:
        out.append(PRIVATE_LINE if errors or blockers or offline else publish_line(path.name, root))
    return errors, out, closed, len(findings)


def publish_line(name, root):
    """The command that publishes a private report: the owner runs it. It
    tests the rule first, so a stale proposal leaves the report private."""
    script = os.path.abspath(__file__)
    inside = os.path.relpath(script, root)
    if not inside.startswith(os.pardir):
        script = inside
    private, public = (shlex.quote("%s/%s" % (folder, name)) for folder in (PRIVATE, PUBLIC))
    return "Publish: python3 %s check --published %s && mv %s %s" % (
        shlex.quote(script), private, private, public)


def status(root):
    """Return (lines, failed): the state of each private report and of each
    published report that the rule applies to: a security report, a report
    with a private copy, a report that an issue names by finding ID. failed
    means that a report in .github/reports/ breaks the publication rule."""
    root = Path(root).resolve()
    lines, failed, issues = [], False, project_issues(root)
    for path in sorted((root / PRIVATE).glob("*.md")):
        label = "%s/%s" % (PRIVATE, path.name)
        twin = root / PUBLIC / path.name
        if twin.is_file():
            if twin.read_bytes() == path.read_bytes():
                lines.append("%s: duplicate of the published report; delete it." % label)
            continue
        errors, _, closed, total = examine(path, root, True, label, issues)
        state = "private" if errors else "publishable"
        lines.append("%s: %s, %d of %d findings closed." % (label, state, closed, total))
        form = [e for e in errors if ": publication: " not in e]
        lines += [e for e in errors if e not in form]
        if form:
            lines.append("%s:1: publication: the report has %d other errors; run check" % (label, len(form)))
        if not errors:
            lines.append(publish_line(path.name, root))
    for path in sorted((root / PUBLIC).glob("*.md")):
        label = "%s/%s" % (PUBLIC, path.name)
        if not ((root / PRIVATE / path.name).is_file()
                or SECURITY_H1_RE.search(path.read_text(encoding="utf-8"))):
            try:
                if path.name not in issues.sources():
                    continue
            except GhError as e:
                lines.append("%s:1: publication: cannot read the issues: %s" % (label, e))
                failed = True
                continue
        errors, _, closed, total = examine(path, root, True, label, issues)
        state = "refused" if errors else "published"
        lines.append("%s: %s, %d of %d findings closed." % (label, state, closed, total))
        lines += errors
        failed = failed or bool(errors)
    return lines, failed


def proposals(root, shown):
    """The hand-over lines for the other reports: each publishable private
    report and each refused public report. Without a mirror there is none:
    the plain check does not read GitHub."""
    lines = []
    for line in status(root)[0] if isinstance(project_issues(root), Mirror) else []:
        m = re.match(r"^(%s/\S+): refused, " % re.escape(PUBLIC), line)
        if line.startswith("Publish: ") and line not in shown:
            lines.append(line)
        elif m:
            lines.append("Refused: %s breaks the publication rule; run status." % m.group(1))
    return lines


def new(mode, model, root, now=None, templates=None):
    """Write the template of the mode as a new report; return its path."""
    now = now or datetime.datetime.now(datetime.timezone.utc)
    templates = Path(templates or Path(__file__).resolve().parent.parent / "templates")
    text = (templates / (mode + ".md")).read_text(encoding="utf-8")
    for token, value in (("<uid>", str(uuid.uuid4())),
                         ("<timestamp>", now.strftime("%Y-%m-%dT%H:%M:%SZ")),
                         ("<model>", model)):
        text = text.replace(token, value, 1)
    folder = Path(root) / (PRIVATE if mode == "security" else PUBLIC)
    name = now.strftime("%Y%m%dT%H%M%S") + "Z.md"
    if mode == "security":
        ignored(root, PRIVATE, name)
    folder.mkdir(parents=True, exist_ok=True)
    path = folder / name
    with open(path, "x", encoding="utf-8") as f:
        f.write(text)
    return path


class Refused(Exception):
    """`new` does not write the report."""


# Alerts

class ScannerOff(Exception):
    """GitHub says that the scanner of a source is off for the repository."""


def gh_api(root, path, jq=None):
    """The answer of `gh api` for the path: each page of a list joined into
    one list. Raise ScannerOff for a scanner that is off and GhError for each
    other failure."""
    if not shutil.which("gh"):
        raise GhError("gh is not on PATH")
    args = ["gh", "api", path] + (["--jq", jq] if jq else ["--paginate"])
    try:
        done = subprocess.run(args, cwd=root, capture_output=True, text=True, timeout=GH_TIMEOUT)
    except subprocess.TimeoutExpired:
        raise GhError("gh does not answer in %g seconds" % GH_TIMEOUT)
    except OSError as e:
        raise GhError("gh does not run: %s" % e)
    if done.returncode:
        lines = [t.strip() for t in done.stderr.splitlines() if t.strip()]
        message = lines[0] if lines else "no message"
        http = HTTP_RE.search(message)
        if http and http.group(1) in ("403", "404") and OFF_RE.search(message):
            raise ScannerOff(re.sub(r"^gh: | \(HTTP \d+\)$", "", message))
        raise GhError("gh exits %d: %s" % (done.returncode, message))
    if jq:
        return done.stdout.strip()
    # --paginate prints one JSON array for each page.
    found, text, at, decoder = [], done.stdout.strip(), 0, json.JSONDecoder()
    try:
        while at < len(text):
            page, at = decoder.raw_decode(text, at)
            if not isinstance(page, list):
                raise ValueError
            found += page
            while at < len(text) and text[at].isspace():
                at += 1
    except ValueError:
        raise GhError("gh returns no JSON list")
    return found


def field(item, *keys):
    """The value at the keys of nested dicts, or "" if one is absent."""
    for key in keys:
        item = item.get(key) if isinstance(item, dict) else None
    return item if item not in (None, "") else ""


def one_line(text):
    return " ".join(str(text).split())


def dependabot(a):
    pkg = field(a, "dependency", "package")
    line = ["#%s" % field(a, "number"),
            field(a, "security_advisory", "severity") or field(a, "security_vulnerability", "severity"),
            "%s:%s" % (field(pkg, "ecosystem"), field(pkg, "name")),
            field(a, "dependency", "manifest_path"), field(a, "state")]
    if field(a, "state") == "dismissed":
        line[-1] += ": %s — %s" % (field(a, "dismissed_reason") or "no reason",
                                   one_line(field(a, "dismissed_comment")) or "no comment")
    return line


def code_scanning(a):
    where = field(a, "most_recent_instance", "location", "path")
    start = field(a, "most_recent_instance", "location", "start_line")
    return ["#%s" % field(a, "number"),
            field(a, "rule", "security_severity_level") or field(a, "rule", "severity"),
            field(a, "rule", "id"), "%s:%s" % (where, start) if start else where, field(a, "state")]


def secret_scanning(a):
    return ["#%s" % field(a, "number"), "—",
            field(a, "secret_type_display_name") or field(a, "secret_type"), "—", field(a, "state")]


def advisory(a):
    names = ", ".join("%s:%s" % (field(v, "package", "ecosystem"), field(v, "package", "name"))
                      for v in a.get("vulnerabilities") or [] if isinstance(v, dict))
    return [field(a, "ghsa_id"), field(a, "severity"), names, "—", field(a, "state")]


# Each source: name, the path under the repository, the states that it
# lists, the form of one alert.
ALERT_SOURCES = [
    ("dependabot", "dependabot/alerts?state=open,dismissed&per_page=100", ["open", "dismissed"], dependabot),
    ("code-scanning", "code-scanning/alerts?state=open&per_page=100", ["open"], code_scanning),
    ("secret-scanning", "secret-scanning/alerts?state=open&per_page=100", ["open"], secret_scanning),
    ("advisories", "security-advisories?per_page=100", ["draft", "triage"], advisory),
]


def alerts(root, repo=None):
    """Return (lines, failed): one line for each alert of each source and one
    summary line for each source. A scanner that is off is `off`. failed
    means that a source that is on, or the repository, is not readable."""
    try:
        repo = repo or gh_api(root, "repos/{owner}/{repo}", ".full_name")
    except (GhError, ScannerOff) as e:
        return ["repository: error: %s" % e], True
    if not REPO_RE.match(repo):
        return ["repository: error: %r is not owner/repo" % repo], True
    lines, failed = ["Repository: %s." % repo], False
    for name, path, states, form in ALERT_SOURCES:
        if name == "advisories":
            # GitHub lists draft advisories only to an admin of the repository,
            # and answers an empty list to anyone else: no list is no proof.
            try:
                admin = gh_api(root, "repos/%s" % repo, ".permissions.admin")
            except (GhError, ScannerOff) as e:
                admin = "error: %s" % e
            if admin != "true":
                lines.append("%s: error: draft advisories need admin access to %s" % (name, repo))
                failed = True
                continue
        try:
            found = gh_api(root, "repos/%s/%s" % (repo, path))
        except ScannerOff as e:
            lines.append("%s: off — %s" % (name, e))
            continue
        except GhError as e:
            lines.append("%s: error: %s" % (name, e))
            failed = True
            continue
        found = [a for a in found if isinstance(a, dict) and a.get("state") in states]
        for cells in map(form, found):
            lines.append("%s %s: %s" % (name, cells[0], ", ".join(str(c) or "—" for c in cells[1:])))
        lines.append("%s: %s." % (name, ", ".join(
            "%d %s" % (sum(1 for a in found if a["state"] == st), st) for st in states)))
    return lines, failed


def not_ignored(root, folder, name):
    """Why Git can commit the file name in the folder of the root, or None
    when Git ignores it. Offline: `git check-ignore` reads local rules only."""
    probe = "%s/%s" % (folder, name)
    try:
        done = subprocess.run(["git", "-C", str(root), "check-ignore", "-q", "--", probe],
                              capture_output=True, text=True)
    except OSError as e:
        return "git does not run: %s" % e
    if done.returncode == 1:
        return ("%s is not ignored by Git; a private report there can reach a commit. "
                "Add %s/ to .gitignore or .git/info/exclude" % (probe, folder))
    if done.returncode:
        return "%s is not in a Git repository; a private report needs an ignored %s/" % (root, folder)
    return None


def ignored(root, folder, name):
    """Raise Refused unless Git ignores the file name in the folder of the root."""
    reason = not_ignored(root, folder, name)
    if reason:
        raise Refused(reason)


def find_root(start):
    try:
        out = subprocess.run(["git", "-C", str(start), "rev-parse", "--show-toplevel"],
                             capture_output=True, text=True, check=True)
    except (OSError, subprocess.CalledProcessError):
        print("review.py: no git repository; pass --root", file=sys.stderr)
        sys.exit(2)
    return out.stdout.strip()


def main(argv=None):
    parser = argparse.ArgumentParser(prog="review.py")
    verbs = parser.add_subparsers(dest="verb", required=True)
    p_new = verbs.add_parser("new")
    p_new.add_argument("mode", choices=list(MODES))
    p_new.add_argument("--model", required=True)
    p_new.add_argument("--root")
    p_check = verbs.add_parser("check")
    p_check.add_argument("file")
    p_check.add_argument("--root")
    p_check.add_argument("--published", action="store_true")
    p_status = verbs.add_parser("status")
    p_status.add_argument("--root")
    p_alerts = verbs.add_parser("alerts")
    p_alerts.add_argument("--repo", help="owner/repo; default: the GitHub repository of the root")
    p_alerts.add_argument("--root")
    args = parser.parse_args(argv)
    if args.verb == "alerts":
        if args.repo and not REPO_RE.match(args.repo):
            parser.error("--repo must be owner/repo")
        lines, failed = alerts(args.root or (os.getcwd() if args.repo else find_root(os.getcwd())), args.repo)
        print("\n".join(lines))
        return 1 if failed else 0
    if args.verb == "status":
        lines, failed = status(args.root or find_root(os.getcwd()))
        print("\n".join(lines or ["No private report and no published security report."]))
        return 1 if failed else 0
    if args.verb == "new":
        root = args.root or find_root(os.getcwd())
        try:
            print(new(args.mode, args.model, root))
        except FileExistsError as e:
            print("review.py: %s exists" % e.filename, file=sys.stderr)
            return 1
        except Refused as e:
            print("review.py: %s" % e, file=sys.stderr)
            return 1
        return 0
    if not os.path.isfile(args.file):
        parser.error("no file " + args.file)
    root = args.root or find_root(Path(args.file).resolve().parent)
    errors, handover = check(args.file, root, args.published)
    own = [publish_line(Path(args.file).name, Path(root).resolve())] if args.published else []
    print("\n".join(errors or handover + proposals(root, handover + own)))
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
