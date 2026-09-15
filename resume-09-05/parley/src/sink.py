"""Subprocess-owned SQLite effects. Opaque mode intentionally has no query/dedup API."""
import argparse
import json
import sqlite3
import sys
from pathlib import Path

from common import digest, encoded, identity


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("directory", type=Path)
    parser.add_argument("mode", choices=["queryable", "opaque"])
    parser.add_argument("action", choices=["put", "query"])
    args = parser.parse_args()
    request = json.loads(sys.stdin.readline())
    if args.mode == "opaque" and args.action == "query":
        print(json.dumps({"refused": "opaque_query_unsupported"}), flush=True)
        return 2
    args.directory.mkdir(parents=True, exist_ok=True)
    db = sqlite3.connect(args.directory / "private.sqlite", timeout=5)
    db.execute("PRAGMA journal_mode=DELETE")
    db.execute("PRAGMA synchronous=FULL")
    db.execute("CREATE TABLE IF NOT EXISTS effects "
               "(id INTEGER PRIMARY KEY, mode TEXT, op TEXT, digest TEXT, payload TEXT, receipt TEXT)")
    db.commit()
    if args.action == "query":
        row = db.execute("SELECT receipt FROM effects WHERE mode=? AND op=?",
                         (args.mode, request["op"])).fetchone()
        print(json.dumps({"receipt": json.loads(row[0]) if row else None}), flush=True)
        return 0
    payload_digest = identity(request["payload"])
    if request["payload_digest"] != payload_digest:
        print(json.dumps({"refused": "request_payload_digest_mismatch"}), flush=True)
        return 2
    db.execute("BEGIN IMMEDIATE")
    row = None
    if args.mode == "queryable":
        row = db.execute("SELECT digest, receipt FROM effects WHERE mode=? AND op=?",
                         (args.mode, request["op"])).fetchone()
    if row and row[0] != payload_digest:
        db.rollback()
        print(json.dumps({"refused": "conflicting_payload", "receipt": json.loads(row[1])}), flush=True)
        return 2
    receipt = json.loads(row[1]) if row else None
    if not receipt:
        cursor = db.execute("INSERT INTO effects(mode,op,digest,payload,receipt) VALUES(?,?,?,?,?)",
                            (args.mode, request["op"], payload_digest,
                             encoded(request["payload"]).decode(), ""))
        receipt = {"effect_id": cursor.lastrowid, "op": request["op"],
                   "payload_digest": payload_digest}
        db.execute("UPDATE effects SET receipt=? WHERE id=?",
                   (encoded(receipt).decode(), cursor.lastrowid))
    db.commit()
    # This line carries no receipt. The caller must release the withheld reply.
    print(json.dumps({"boundary": "sink_committed_reply_withheld"}), flush=True)
    if sys.stdin.readline().strip() != "release":
        return 0
    print(json.dumps({"receipt": receipt}), flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())
