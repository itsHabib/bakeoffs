# hack3-spend-audit — no build step, just convenience targets.
# Prefer the venv python (tiktoken) if it exists; else fall back to system.
PY := $(shell [ -x .venv/bin/python ] && echo .venv/bin/python || echo python3)

.PHONY: demo real test venv vet clean

## demo: canned, hands-free — run the bundled fixtures, open the dashboard.
demo:
	$(PY) audit.py --root fixtures --out out --open

## real: the money shot — run against the real ~/.claude/projects corpus.
real:
	$(PY) audit.py --out out-real --open

## test: exact-number table tests on the deterministic core.
test:
	$(PY) -m unittest test_audit -v

## vet: byte-compile everything (the stdlib's "does it parse" check).
vet:
	$(PY) -m py_compile audit.py test_audit.py fixtures/_gen.py

## venv: optional — install tiktoken for precise tool-result token counts.
venv:
	python3 -m venv .venv && .venv/bin/pip install -q tiktoken
	@echo "venv ready — 'make real' now uses tiktoken o200k_base."

clean:
	rm -rf out out-real
