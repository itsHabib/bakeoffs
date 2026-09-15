.PHONY: demo test build
build:
	python3 tools.py
	$(MAKE) -C vendor/parley/compiler build
	mkdir -p artifacts/contracts
	vendor/parley/compiler/bin/parleyc compile protocol.parley artifacts/contracts

test: build
	$(MAKE) -C vendor/parley/compiler test
	python3 tests/run.py

demo:
	python3 demo.py
