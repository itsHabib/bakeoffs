"""Check vendored provenance and count the maintained implementation, including adapters."""
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
lock = json.loads((ROOT / "SOURCE-LOCK.json").read_text())
for name, expected in lock["files"].items():
    actual = hashlib.sha256((ROOT / "vendor/parley" / name).read_bytes()).hexdigest()
    assert actual == expected, f"vendored source changed: {name}"

groups = {"ordinary executor, journal, checker and sink": [ROOT / "src" / name for name in ("common.py", "worker.py", "sink.py")],
          "Parley normalization adapter": [ROOT / "src/audit.py"],
          "controller and mutants": list((ROOT / "tests").glob("*.py")),
          "local demo and provenance tools": [ROOT / "tools.py", ROOT / "demo.py"],
          "protocol": [ROOT / "protocol.parley"],
          "reused Haskell compiler and observer": list((ROOT / "vendor/parley/compiler/src").glob("*.hs")) + list((ROOT / "vendor/parley/compiler/app").glob("*.hs")),
          "reused Haskell tests": list((ROOT / "vendor/parley/compiler/test").glob("*.hs")),
          "reused protocol fixtures": list((ROOT / "vendor/parley/protocols").glob("*.parley")),
          "build files": [ROOT / "Makefile", ROOT / "vendor/parley/compiler/Makefile"]}
counts = {group: {"files": len(paths), "lines": sum(len(p.read_text().splitlines()) for p in paths)}
          for group, paths in groups.items()}
print(json.dumps({"pinned_commit": lock["commit"], "verified_vendored_files": len(lock["files"]),
                  "counts_including_blank_and_comment_lines": counts,
                  "total_lines": sum(v["lines"] for v in counts.values())}, indent=2))
