module Example
  ( roles,
    validProtocol,
    invalidProtocol,
  )
where

import Protocol (Protocol (..), Role)

roles :: [Role]
roles = ["producer", "collector", "kernel", "human", "gate"]

validProtocol :: Protocol
validProtocol =
  Message "producer" "collector" "evidence.receipt" $
    Choice
      "collector"
      ["producer", "kernel", "human", "gate"]
      "accepted"
      ( Message "collector" "kernel" "validate" $
          Choice
            "kernel"
            ["collector", "human", "gate"]
            "pass"
            (Message "kernel" "gate" "assurance.pass" End)
            "gap"
            ( Message "kernel" "human" "assurance.gap" $
                Message "human" "gate" "judgment" End
            )
      )
      "malformed"
      (Message "collector" "producer" "receipt.rejected" End)

-- Gate has different obligations in the accepted and malformed branches but
-- collector does not notify it. Projection must refuse this protocol.
invalidProtocol :: Protocol
invalidProtocol =
  Message "producer" "collector" "evidence.receipt" $
    Choice
      "collector"
      ["producer", "kernel", "human"]
      "accepted"
      (Message "kernel" "gate" "assurance.pass" End)
      "malformed"
      End
