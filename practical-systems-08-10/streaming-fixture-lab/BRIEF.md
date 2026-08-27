# Streaming Fixture Lab — frozen brief

## Bet

Large newline-delimited or chunked event streams can be captured, normalized,
redacted, fingerprinted, and replayed as deterministic fixtures without loading
the whole stream into memory or requiring a broker.

## One job

Turn an input stream into a bounded-memory fixture manifest plus replayable
chunks, preserving order and explicit truncation/error markers.

## Hard case

A sensitive field straddles chunk boundaries immediately before a malformed
record. Redaction must still occur, the malformed record must remain observable,
and replay must preserve the exact ordering of valid records around it.

## Gleam thesis

Typed streaming stages and explicit result values make partial consumption and
failure propagation difficult to ignore. It loses if an ordinary Node stream
pipeline is equally legible, safe, and easy to embed.

## Non-goals

No Kafka replacement, object store, fixture service, distributed replay system,
schema registry, or web UI.
