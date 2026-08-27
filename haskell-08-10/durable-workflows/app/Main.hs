module Main (main) where

import Example
  ( deliveryProgram,
    deliveryProgramWithDispatchId,
    emptyJournal,
    initialFixture,
    initialTask,
  )
import Workflow
  ( Fault (..),
    Outcome (..),
    runProgram,
  )

main :: IO ()
main = do
  putStrLn "CRASH AFTER DISPATCH COMMIT"
  case runProgram (CrashAfterCommit "dispatch") initialFixture emptyJournal initialTask deliveryProgram of
    Crashed point journal afterCrash -> do
      print point
      print journal
      print afterCrash
      putStrLn "\nRESUME"
      case runProgram NoFault afterCrash journal initialTask deliveryProgram of
        Completed receipt finalJournal finalFixture -> do
          print receipt
          print finalJournal
          print finalFixture
          putStrLn "\nCHANGE STABLE STEP ID (REFUSED)"
          printRefusal $
            runProgram
              NoFault
              finalFixture
              journal
              initialTask
              (deliveryProgramWithDispatchId "dispatch-v2")
        _ -> fail "resume did not complete"
    _ -> fail "fault was not injected"

printRefusal :: Outcome value -> IO ()
printRefusal outcome =
  case outcome of
    Refused problem _ -> print problem
    _ -> fail "changed workflow unexpectedly accepted old journal"
