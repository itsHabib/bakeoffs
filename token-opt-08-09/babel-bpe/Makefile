VENV := .venv
PY := $(VENV)/bin/python
PIP := $(VENV)/bin/pip

.PHONY: demo leaderboard comprehension test setup clean

## demo: the canned 60-second walkthrough (offline-deterministic + live check)
demo: $(VENV)
	@$(PY) src/run.py demo

## leaderboard: token leaderboard only (fully offline, deterministic)
leaderboard: $(VENV)
	@$(PY) src/run.py leaderboard

## comprehension: run the live rig against qwen2.5:7b, refresh the cache
comprehension: $(VENV)
	@$(PY) src/run.py comprehension

## test: encoder round-trips, pinned token counts, grader — all deterministic
test: $(VENV)
	@$(PY) -m unittest discover -s tests

setup: $(VENV)

$(VENV):
	python3 -m venv $(VENV)
	$(PIP) install -q tiktoken

clean:
	rm -rf $(VENV) src/__pycache__ tests/__pycache__
