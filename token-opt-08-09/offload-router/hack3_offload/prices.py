"""The ONE editable $/MTok price table.

House rule: prices are never hardcoded from model memory. Fill these in from
https://docs.claude.com/en/docs/about-claude/pricing before trusting any dollar
figure. Until then they are ``None`` and ``report`` prints tokens only.

Four columns because the four token kinds are priced differently:
  - fresh_input        : uncached input tokens
  - cache_write        : writing a prompt-cache entry (usually > fresh_input)
  - cache_read         : reading a cached prefix (usually << fresh_input)
  - output             : generated tokens

For the offload router the relevant comparison is a *fresh* frontier call
(fresh_input + output) that a local model displaced. The other columns are here
so this table stays the single source of price truth for the repo.
"""

from __future__ import annotations

# TODO(operator): fill $/MTok for the frontier model you are displacing.
# Values are dollars per 1,000,000 tokens. Leave None to keep report tokens-only.
FRONTIER_MODEL = "claude-<TODO>"  # e.g. the tier your agents actually run on

PRICES_PER_MTOK = {
    "fresh_input": None,   # TODO
    "cache_write": None,   # TODO
    "cache_read": None,    # TODO
    "output": None,        # TODO
}


def can_price() -> bool:
    """True once the operator has filled the columns the router needs."""
    return (
        PRICES_PER_MTOK["fresh_input"] is not None
        and PRICES_PER_MTOK["output"] is not None
    )


def dollars(input_tokens: int, output_tokens: int) -> float | None:
    """Dollar cost of a fresh frontier call, or None if prices unfilled."""
    if not can_price():
        return None
    fi = PRICES_PER_MTOK["fresh_input"]
    out = PRICES_PER_MTOK["output"]
    return (input_tokens / 1_000_000) * fi + (output_tokens / 1_000_000) * out
