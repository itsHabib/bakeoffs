# Demo

Run `stack run`. A four-stage project compiles into two genuinely parallel work
items followed by validation and landing. A second project requests parallel
edits under overlapping `src` scopes and is serialized with the common scope
named. A third attempts to land without a validation ancestor and is refused.
