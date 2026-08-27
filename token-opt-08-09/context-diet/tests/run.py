#!/usr/bin/env python3
"""Zero-dependency test runner. pytest is not on this machine and isn't
unavoidable, so we discover `test_*` functions across tests/test_*.py and
run them, handing any function that takes an argument a fresh temp dir
(the one pytest fixture we use). Plain `assert` gives readable failures.
"""

import importlib.util
import inspect
import sys
import tempfile
import traceback
from pathlib import Path

HERE = Path(__file__).parent


def _load(path):
    spec = importlib.util.spec_from_file_location(path.stem, path)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def main():
    sys.path.insert(0, str(HERE.parent))
    files = sorted(HERE.glob("test_*.py"))
    passed = failed = 0
    failures = []
    for f in files:
        mod = _load(f)
        for name, fn in sorted(vars(mod).items()):
            if not (name.startswith("test_") and inspect.isfunction(fn)):
                continue
            try:
                if inspect.signature(fn).parameters:
                    with tempfile.TemporaryDirectory() as td:
                        fn(Path(td))
                else:
                    fn()
                passed += 1
                print(f"  ok   {f.stem}.{name}")
            except Exception:  # noqa: BLE001 — a test runner reports all
                failed += 1
                failures.append((f.stem, name, traceback.format_exc()))
                print(f"  FAIL {f.stem}.{name}")

    print(f"\n{passed} passed, {failed} failed")
    for stem, name, tb in failures:
        print(f"\n--- {stem}.{name} ---\n{tb}")
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
