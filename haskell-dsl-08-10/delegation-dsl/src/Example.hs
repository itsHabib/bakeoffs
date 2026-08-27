module Example
  ( findings,
    patch,
    verdict,
    validContract,
    missingContract,
    ambiguousContract,
  )
where

import Delegation
  ( Artifact (..),
    Contract,
    agent,
    andProduces,
    contract,
    produces,
    requires,
  )

findings :: Artifact
findings = Artifact "Findings"

patch :: Artifact
patch = Artifact "Patch"

verdict :: Artifact
verdict = Artifact "Verdict"

validContract :: Contract
validContract =
  contract
    [ agent "researcher" `produces` findings,
      agent "implementer" `requires` findings `andProduces` patch,
      agent "reviewer" `requires` patch `andProduces` verdict
    ]

missingContract :: Contract
missingContract =
  contract
    [ agent "implementer" `requires` findings `andProduces` patch,
      agent "reviewer" `requires` patch
    ]

ambiguousContract :: Contract
ambiguousContract =
  contract
    [ agent "researcher-a" `produces` findings,
      agent "researcher-b" `produces` findings,
      agent "implementer" `requires` findings
    ]
