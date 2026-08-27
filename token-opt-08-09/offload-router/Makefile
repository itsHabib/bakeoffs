# hack3-offload-router — a deterministic quality gate for local-model offload.
PY := .venv/bin/python
PIP := .venv/bin/pip

.PHONY: help venv demo demo-live test report clean

help:
	@echo "make venv       - create .venv and install tiktoken"
	@echo "make demo       - canned, deterministic demo (cassettes, no ollama needed)"
	@echo "make demo-live  - same demo against live ollama qwen2.5:7b"
	@echo "make test       - run the verifier + gate table tests"
	@echo "make report     - print the displacement ledger from the last demo"
	@echo "make clean       - remove venv, ledgers, __pycache__"

venv:
	python3 -m venv .venv
	$(PIP) install -q tiktoken
	@echo "venv ready ($$($(PY) -c 'import tiktoken; print(\"tiktoken\", tiktoken.__version__)'))"

# Deterministic: replays committed cassettes, writes a fresh demo ledger.
demo:
	$(PY) -m hack3_offload.cli demo --ledger .hack3-demo-ledger.jsonl

# Bonus live moment: real qwen2.5:7b. Falls back to REJECT on any bad output.
demo-live:
	$(PY) -m hack3_offload.cli demo --live --ledger .hack3-demo-ledger.jsonl

test:
	$(PY) -m unittest discover -s tests -v

report:
	$(PY) -m hack3_offload.cli report --ledger .hack3-demo-ledger.jsonl

clean:
	rm -rf .venv .hack3-*.jsonl .hack3-ledger.jsonl
	find . -name __pycache__ -type d -prune -exec rm -rf {} +
