---
id: 174
database_id: 1996436031
node_id: I_kwDOE2M9Zc52_zI_
status: closed
title: "github: contribution: html markup was changed"
labels: [scope: code, scope: test, type: bug, severity: critical, scope: inventory, impact: high, effort: medium]
url: https://github.com/octomation/maintainer/issues/174
created_at: 2023-11-16T09:29:52Z
updated_at: 2023-11-16T10:37:58Z
---

# github: contribution: html markup was changed

**Details**

```bash
maintainer github contribution lookup 2022-09-18/3

 Day / Week   #37   #38   #39    Date
------------ ----- ----- ----- --------
 Sunday        -     -     -    Sep 25
 Monday        -     -     -    Sep 26
 Tuesday       -     -     -    Sep 27
 Wednesday     -     -     -    Sep 28
 Thursday      -     -     -    Sep 29
 Friday        -     -     -    Sep 30
 Saturday      -     -     -    Oct  1
------------ ----- ----- ----- --------
```

GitHub again changed html structure. The new markup

```html
<td></td>

<td tabindex="0" data-ix="1" aria-selected="false" aria-describedby="contribution-graph-legend-level-0" style="width: 10px" data-date="1986-01-05" data-level="0" id="contribution-day-component-0-1" role="gridcell" data-view-component="true" class="ContributionCalendar-day"></td>

<tool-tip id="tooltip-82615c43-31e0-4660-a8f7-83050a7abf58" for="contribution-day-component-0-1"
 popover="manual" data-direction="n" data-type="label" data-view-component="true" class="sr-only position-absolute">
  No contributions on January 5th.
</tool-tip>
```

Now, contribution counter is located inside new node `<tool-tip/>`, but contribution date is still in `<td/>`.

**Checklist**

- [x] Fix broken logic.
- [x] Add daily integration tests.
- [x] Remove legacy markers about previous formats.
