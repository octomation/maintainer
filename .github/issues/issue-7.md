---
id: 7
database_id: 811110070
node_id: MDU6SXNzdWU4MTExMTAwNzA=
status: closed
title: "remove double new lines"
labels: [help wanted]
url: https://github.com/octomation/maintainer/issues/7
created_at: 2021-02-18T13:41:46Z
updated_at: 2021-03-05T19:07:21Z
---

# remove double new lines

input:

```make
include src/common/env.mk

export PATH := $(GOBIN):$(PATH)
```

expected:

```make
make-verbose:
	$(eval AT :=)
	$(eval MAKE := $(MAKE) verbose)
	@echo > /dev/null
.PHONY: make-verbose


export PATH := $(GOBIN):$(PATH)
```

obtained:

```make
make-verbose:
	$(eval AT :=)
	$(eval MAKE := $(MAKE) verbose)
	@echo > /dev/null
.PHONY: make-verbose

export PATH := $(GOBIN):$(PATH)
```
