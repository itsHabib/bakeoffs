module Example
  ( mergeRule,
    subject,
    validEvidence,
    staleEvidence,
    humanEvidence,
  )
where

import Assurance
  ( Evidence (..),
    Observation (..),
    Requirement (..),
    Revision (..),
    Rule,
    Subject (..),
    checkPassed,
    exactHead,
    gateReceipt,
    humanJudgment,
    ready,
    (.&&.),
    (.||.),
  )

mergeRule :: Rule
mergeRule =
  ready "merge" $
    exactHead
      .&&. checkPassed "unit"
      .&&. checkPassed "lint"
      .&&. (gateReceipt .||. humanJudgment)

subject :: Subject
subject = Subject "PR-42" (Revision "sha-b")

validEvidence :: Evidence
validEvidence =
  Evidence
    (Revision "sha-b")
    [ Observation (CheckPassed "unit") (Revision "sha-b") True "check/unit/91",
      Observation (CheckPassed "lint") (Revision "sha-b") True "check/lint/92",
      Observation GateReceipt (Revision "sha-b") True "gate/run-17"
    ]

staleEvidence :: Evidence
staleEvidence =
  Evidence
    (Revision "sha-b")
    [ Observation (CheckPassed "unit") (Revision "sha-a") True "check/unit/80",
      Observation (CheckPassed "lint") (Revision "sha-a") True "check/lint/81",
      Observation GateReceipt (Revision "sha-a") True "gate/run-12"
    ]

humanEvidence :: Evidence
humanEvidence =
  validEvidence
    { observations =
        filter ((/= GateReceipt) . observedRequirement) (observations validEvidence)
          ++ [Observation HumanJudgment (Revision "sha-b") True "judgment/4"]
    }
