# context-diet — offline replay harness for tool-result hygiene rules.
# `make demo` is the hands-free judge path: setup, tests, fixtures, corpus.

VENV := .venv
PY   := $(VENV)/bin/python
PIP  := $(VENV)/bin/pip
CORPUS ?= $(HOME)/.claude/projects

.PHONY: demo setup test fixtures corpus sweep html clean

setup: $(VENV)/.stamp
$(VENV)/.stamp:
	python3 -m venv $(VENV)
	$(PIP) install -q -r requirements.txt
	@touch $@

test: setup
	$(PY) tests/run.py

fixtures: setup
	$(PY) fixtures/make_fixtures.py
	$(PY) cli.py --fixtures

corpus: setup
	$(PY) cli.py --corpus "$(CORPUS)"

sweep: setup
	$(PY) cli.py --corpus "$(CORPUS)" --sweep

html: setup
	$(PY) cli.py --corpus "$(CORPUS)" --html report.html

# the canned demo: tests green, then planted-waste fixtures, then the real
# corpus, then the threshold pricing sweep. No live input required.
demo: setup
	@echo "\n=== 1/4  policy-layer tests (every rule + reference detection) ==="
	@$(PY) tests/run.py | tail -1
	@echo "\n=== 2/4  bundled fixtures — known planted waste, asserted in tests ==="
	@$(PY) fixtures/make_fixtures.py >/dev/null && $(PY) cli.py --fixtures
	@echo "\n=== 3/4  real corpus — per-rule savings + computed risk ==="
	@$(PY) cli.py --corpus "$(CORPUS)" --html report.html
	@echo "\n=== 4/4  threshold sweep — the harness pricing each idea ==="
	@$(PY) cli.py --corpus "$(CORPUS)" --sweep

clean:
	rm -rf $(VENV) report.html **/__pycache__ __pycache__
