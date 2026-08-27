module Main (main) where

import Control.Monad (unless)
import Example (overlapProject, reviewKernel, ungatedProject)
import WorkDriver
  ( CompileError (..),
    DriverPlan (..),
    ParallelDecision (..),
    Project,
    TaskView (..),
    compile,
  )

main :: IO ()
main = do
  valid <- requirePlan reviewKernel
  assert "independent implementation and fixtures share a batch" $
    map (map taskName) (batches valid)
      == [["spec"], ["implement", "fixtures"], ["local-green"], ["merge"]]
  assert "critical path reaches landing" $
    criticalPath valid
      `elem` [ ["spec", "implement", "local-green", "merge"],
               ["spec", "fixtures", "local-green", "merge"]
             ]
  assert "parallel request is accepted" $
    parallelDecisions valid == [ParallelAccepted "implement" "fixtures"]
  overlap <- requirePlan overlapProject
  assert "overlapping work is serialized" $
    map (map taskName) (batches overlap)
      == [["spec"], ["edit-gate"], ["edit-gate-tests"], ["green"], ["merge"]]
  assert "serialization explains the common scope" $
    parallelDecisions overlap
      == [ParallelSerialized "edit-gate" "edit-gate-tests" ["src/**"]]
  assert "landing requires a validation ancestor" $
    compile ungatedProject == Left [UngatedLanding "merge"]
  assert "compilation is deterministic" (compile reviewKernel == compile reviewKernel)
  putStrLn "work-driver-dsl: all tests passed"

requirePlan :: Project -> IO DriverPlan
requirePlan projectValue =
  case compile projectValue of
    Left problems -> fail (show problems)
    Right plan -> pure plan

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
