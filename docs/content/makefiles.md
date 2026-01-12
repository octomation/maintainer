---
title: Makefiles
description: Compile Makefiles with local include directives into distributable files.
---

# Makefiles

Keep build rules in small, focused fragments while you work, then use `makefile build` to hand someone one self-contained Makefile. The command expands `include` and `-include` lines recursively and collapses repeated blank lines.

```makefile
# release.mk
include make/common.mk
-include make/local.mk

release:
	@echo ready
```

```sh
maintainer makefile build release.mk
# Output: dist/release/Makefile
```

Pass several paths to build each separately. With no arguments, the command reads one input path per line from standard input:

```sh
printf 'release.mk\ndocs.mk\n' | maintainer makefile build
```

`include` must resolve to a readable file relative to the working directory or the build fails. A missing `-include` file is skipped. Avoid include cycles: this version has no cycle detection. The command does not run `make`; it writes bundled files under `dist/<input filename without extension>/Makefile`. Inputs with the same base filename use the same output directory, so give them distinct names when building together.
