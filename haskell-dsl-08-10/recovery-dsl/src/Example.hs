module Example
  ( delivery,
    changedDelivery,
    notification,
    task,
    verified,
    fixture,
    emptyJournal,
  )
where

import Recovery
  ( Activity (..),
    Fixture (..),
    Journal (..),
    Program (..),
    Receipt,
    Task (..),
    Verified (..),
    RunId (..),
    Workflow (..),
    activity,
    atLeastOnce,
    deduplicated,
    idempotent,
    (>>>),
  )

delivery :: Program Task Receipt
delivery =
  Program "delivery-v1" $
    activity "dispatch" Dispatch deduplicated
      >>> activity "verify" Verify idempotent
      >>> activity "record" Record idempotent
      >>> Done

changedDelivery :: Program Task Receipt
changedDelivery =
  Program "delivery-v1" $
    activity "dispatch-v2" Dispatch deduplicated
      >>> activity "verify" Verify idempotent
      >>> activity "record" Record idempotent
      >>> Done

notification :: Program Verified Receipt
notification = Program "notify-v1" (activity "notify" Record atLeastOnce >>> Done)

task :: Task
task = Task "PR-42"

verified :: Verified
verified = Verified (RunId "run-existing")

fixture :: Fixture
fixture = Fixture 1 0 0 0 [] [] []

emptyJournal :: Journal
emptyJournal = Journal []
