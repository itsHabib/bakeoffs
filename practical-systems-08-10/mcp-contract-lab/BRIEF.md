# MCP Contract Lab — frozen brief

## Bet

An MCP tool declaration plus a small set of recorded requests and responses can
become an executable compatibility contract: schema validation, JSON-RPC error
semantics, content-block shape, and deterministic redaction checked in one local
command.

## One job

Compile a readable contract and recorded exchanges into a portable fixture pack
that any MCP client or server test suite can replay.

## Hard case

A server returns HTTP success and JSON-RPC success, but its structured result no
longer satisfies the tool's declared output schema. The lab must report contract
drift at the exact response path rather than accepting transport success.

## Gleam thesis

Closed request/result/error types and exhaustive decoders make protocol failure
states explicit while retaining a small functional pipeline. It loses if a
TypeScript schema validator is equally clear and produces equally portable
failure explanations.

## Non-goals

No MCP proxy, hosted registry, server runtime, traffic collector, auth system, or
generic API-testing platform.
