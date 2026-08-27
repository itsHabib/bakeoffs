"""hack3-offload-router: a deterministic quality gate for local-model offload.

Offloading mechanical sub-steps to a $0 local model saves frontier tokens, but
raw 7B output is not trustworthy. This package makes trust *mechanical*: every
local result is checked by a deterministic verifier before it is allowed out,
and a ledger proves how many frontier tokens were displaced.

The model is phrasing. The verifier is law.
"""
