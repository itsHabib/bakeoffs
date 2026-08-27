module Main (main) where

import Control.Monad (unless)
import Example (changedDelivery, delivery, emptyJournal, fixture, notification, task, verified)
import Recovery
  ( Error (..),
    Event (..),
    Fault (..),
    Fixture (..),
    Outcome (..),
    run,
  )

main :: IO ()
main = do
  checkDeduplicated
  checkAtLeastOnce
  checkJournalReuse
  checkShapeDrift
  putStrLn "recovery-dsl: all tests passed"

checkDeduplicated :: IO ()
checkDeduplicated =
  case run (CrashAfterEffect "dispatch") fixture emptyJournal task delivery of
    Crashed _ journal afterCrash ->
      case run NoFault afterCrash journal task delivery of
        Complete _ _ final -> do
          assert "external dedupe prevents a second dispatch" (dispatches final == 1)
          assert "dedupe reuse is visible" (ReusedExternal "dispatch" `elem` events final)
        _ -> fail "deduplicated resume failed"
    _ -> fail "dedupe fault did not fire"

checkAtLeastOnce :: IO ()
checkAtLeastOnce =
  case run (CrashAfterEffect "notify") fixture emptyJournal verified notification of
    Crashed _ journal afterCrash ->
      case run NoFault afterCrash journal verified notification of
        Complete _ _ final -> do
          assert "at-least-once effect duplicates" (records final == 2)
          assert "duplication risk is explicit" (RetriedAtLeastOnce "notify" `elem` events final)
        _ -> fail "at-least-once resume failed"
    _ -> fail "at-least-once fault did not fire"

checkJournalReuse :: IO ()
checkJournalReuse =
  case run (CrashAfterCommit "dispatch") fixture emptyJournal task delivery of
    Crashed _ journal afterCrash ->
      case run NoFault afterCrash journal task delivery of
        Complete _ _ final -> do
          assert "committed dispatch is not repeated" (dispatches final == 1)
          assert "journal reuse is visible" (ReusedJournal "dispatch" `elem` events final)
        _ -> fail "journal resume failed"
    _ -> fail "commit fault did not fire"

checkShapeDrift :: IO ()
checkShapeDrift =
  case run (CrashAfterCommit "dispatch") fixture emptyJournal task delivery of
    Crashed _ journal afterCrash ->
      case run NoFault afterCrash journal task changedDelivery of
        Refused (ShapeDrift 0 "dispatch-v2" "dispatch") _ -> pure ()
        _ -> fail "changed step identity was accepted"
    _ -> fail "could not prepare journal"

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
