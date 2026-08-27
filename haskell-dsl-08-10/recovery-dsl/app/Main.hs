module Main (main) where

import Example (changedDelivery, delivery, emptyJournal, fixture, notification, task, verified)
import Recovery (Fault (..), Outcome (..), events, records, run)

main :: IO ()
main = do
  putStrLn "DEDUPLICATED DISPATCH: CRASH AFTER EFFECT"
  case run (CrashAfterEffect "dispatch") fixture emptyJournal task delivery of
    Crashed _ journal afterCrash ->
      case run NoFault afterCrash journal task delivery of
        Complete _ _ final -> print final
        _ -> fail "deduplicated resume failed"
    _ -> fail "fault did not fire"
  putStrLn "AT-LEAST-ONCE NOTIFICATION: SAME CRASH WINDOW"
  case run (CrashAfterEffect "notify") fixture emptyJournal verified notification of
    Crashed _ journal afterCrash ->
      case run NoFault afterCrash journal verified notification of
        Complete _ _ final -> do
          print (records final)
          print (events final)
        _ -> fail "notification resume failed"
    _ -> fail "fault did not fire"
  putStrLn "CHANGED STEP ID"
  case run (CrashAfterCommit "dispatch") fixture emptyJournal task delivery of
    Crashed _ journal afterCrash -> printRefusal (run NoFault afterCrash journal task changedDelivery)
    _ -> fail "fault did not fire"

printRefusal :: Outcome value -> IO ()
printRefusal outcome =
  case outcome of
    Refused problem _ -> print problem
    _ -> fail "changed workflow was accepted"
