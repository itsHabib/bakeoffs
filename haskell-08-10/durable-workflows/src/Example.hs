module Example
  ( deliveryProgram,
    deliveryProgramWithDispatchId,
    initialFixture,
    initialTask,
    emptyJournal,
  )
where

import Workflow
  ( Activity (..),
    Fixture (..),
    Journal (..),
    Program (..),
    Receipt,
    Step (..),
    Task (..),
    Workflow (..),
  )

deliveryProgram :: Program Task Receipt
deliveryProgram = deliveryProgramWithDispatchId "dispatch"

deliveryProgramWithDispatchId :: String -> Program Task Receipt
deliveryProgramWithDispatchId dispatchIdentifier =
  Program
    { programVersion = "delivery-v1",
      programWorkflow =
        Then
          (Step dispatchIdentifier Dispatch)
          ( Then
              (Step "verify" Verify)
              (Then (Step "record" Record) Halt)
          )
    }

initialFixture :: Fixture
initialFixture =
  Fixture
    { nextRunNumber = 1,
      dispatchCount = 0,
      verifyCount = 0,
      recordCount = 0,
      verificationPasses = True
    }

initialTask :: Task
initialTask = Task "ship/pr-42"

emptyJournal :: Journal
emptyJournal = Journal []
