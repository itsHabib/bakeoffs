module Main (main) where

import Control.Monad (unless)
import Example
  ( deliveryProgram,
    deliveryProgramWithDispatchId,
    emptyJournal,
    initialFixture,
    initialTask,
  )
import Workflow
  ( CrashPoint (..),
    Entry (..),
    Fault (..),
    Fixture (..),
    Journal (..),
    JournalError (..),
    Outcome (..),
    Program (..),
    Task (..),
    WorkflowError (..),
    runProgram,
  )

main :: IO ()
main = do
  checkCrashAfterCommit
  checkAtLeastOnceWindow
  checkChangedStepRefused
  checkVersionDriftRefused
  checkReorderedAndDuplicateRefused
  checkDuplicateWorkflowStepRefused
  checkInputConflict
  putStrLn "durable-workflows: all replay tests passed"

checkCrashAfterCommit :: IO ()
checkCrashAfterCommit =
  case crashAfterCommit of
    Crashed (AfterCommit "dispatch") journal afterCrash -> do
      assert "dispatch executed before crash" (dispatchCount afterCrash == 1)
      assert "dispatch result committed" (length (entries journal) == 1)
      case runProgram NoFault afterCrash journal initialTask deliveryProgram of
        Completed _ completedJournal final -> do
          assert "recorded dispatch is not repeated" (dispatchCount final == 1)
          assert "remaining activities run once" $
            verifyCount final == 1 && recordCount final == 1
          assert "journal completes all steps" (length (entries completedJournal) == 3)
        _ -> fail "resume after committed dispatch did not complete"
    _ -> fail "crash-after-commit fixture did not crash at dispatch"

checkAtLeastOnceWindow :: IO ()
checkAtLeastOnceWindow =
  case runProgram (CrashAfterEffectBeforeCommit "dispatch") initialFixture emptyJournal initialTask deliveryProgram of
    Crashed (AfterEffectBeforeCommit "dispatch") journal afterCrash -> do
      assert "effect happened" (dispatchCount afterCrash == 1)
      assert "result was not journaled" (entries journal == [])
      case runProgram NoFault afterCrash journal initialTask deliveryProgram of
        Completed _ _ final ->
          assert "uncommitted effect honestly retries" (dispatchCount final == 2)
        _ -> fail "resume from uncommitted effect did not complete"
    _ -> fail "effect-before-commit fixture did not crash"

checkChangedStepRefused :: IO ()
checkChangedStepRefused =
  case crashAfterCommit of
    Crashed _ journal afterCrash ->
      case runProgram NoFault afterCrash journal initialTask (deliveryProgramWithDispatchId "dispatch-v2") of
        Refused (InvalidJournal (ShapeMismatch 0 _ _)) _ -> pure ()
        _ -> fail "changed stable step ID did not refuse replay"
    _ -> fail "could not prepare replay journal"

checkVersionDriftRefused :: IO ()
checkVersionDriftRefused =
  case crashAfterCommit of
    Crashed _ journal afterCrash ->
      let changedVersion = deliveryProgram {programVersion = "delivery-v2"}
       in case runProgram NoFault afterCrash journal initialTask changedVersion of
            Refused (InvalidJournal (VersionMismatch "delivery-v2" "delivery-v1")) _ -> pure ()
            _ -> fail "version drift did not refuse replay"
    _ -> fail "could not prepare replay journal"

checkReorderedAndDuplicateRefused :: IO ()
checkReorderedAndDuplicateRefused =
  case runProgram NoFault initialFixture emptyJournal initialTask deliveryProgram of
    Completed _ (Journal [dispatchEntry, verifyEntry, recordEntry]) final -> do
      let reordered = Journal [verifyEntry, dispatchEntry]
          duplicate = Journal [dispatchEntry, dispatchEntry]
      case runProgram NoFault final reordered initialTask deliveryProgram of
        Refused (InvalidJournal (ShapeMismatch 0 _ _)) _ -> pure ()
        _ -> fail "reordered journal was not refused"
      case runProgram NoFault final duplicate initialTask deliveryProgram of
        Refused (InvalidJournal (DuplicateStepId "dispatch")) _ -> pure ()
        _ -> fail "duplicate journal entry was not refused"
      assert "completed fixture was used" (recordCount final == 1)
      assert "record entry was produced" (entryStepId recordEntry == "record")
    _ -> fail "could not prepare completed journal"

checkDuplicateWorkflowStepRefused :: IO ()
checkDuplicateWorkflowStepRefused =
  case runProgram NoFault initialFixture emptyJournal initialTask (deliveryProgramWithDispatchId "verify") of
    Refused (InvalidJournal (DuplicateWorkflowStepId "verify")) _ -> pure ()
    _ -> fail "workflow with duplicate stable step IDs was allowed to execute"

checkInputConflict :: IO ()
checkInputConflict =
  case crashAfterCommit of
    Crashed _ journal afterCrash ->
      case runProgram NoFault afterCrash journal (Task "ship/pr-99") deliveryProgram of
        Refused (InputConflict "dispatch" _ _) _ -> pure ()
        _ -> fail "changed input reused a recorded result"
    _ -> fail "could not prepare replay journal"

crashAfterCommit :: Outcome value
crashAfterCommit =
  case runProgram (CrashAfterCommit "dispatch") initialFixture emptyJournal initialTask deliveryProgram of
    Crashed point journal fixture -> Crashed point journal fixture
    _ -> error "expected crash"

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
