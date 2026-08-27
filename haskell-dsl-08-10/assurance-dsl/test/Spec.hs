module Main (main) where

import Assurance
  ( Evidence (..),
    Finding (..),
    Observation (..),
    Requirement (..),
    Status (..),
    Subject (..),
    Trace (..),
    evaluate,
    requirements,
    traceStatus,
  )
import Control.Monad (unless)
import qualified Data.Set as Set
import Example (humanEvidence, mergeRule, staleEvidence, subject, validEvidence)

main :: IO ()
main = do
  assert "gate evidence satisfies policy" $
    traceStatus (evaluate subject validEvidence mergeRule) == Passed
  assert "human alternative satisfies policy" $
    traceStatus (evaluate subject humanEvidence mergeRule) == Passed
  let contradictory =
        validEvidence
          { observations =
              Observation GateReceipt (subjectRevision subject) False "older failure"
                : observations validEvidence
          }
  assert "evidence order does not hide a valid proof" $
    traceStatus (evaluate subject contradictory mergeRule) == Passed
  let staleTrace = evaluate subject staleEvidence mergeRule
  assert "stale evidence cannot satisfy policy" (traceStatus staleTrace == Failed)
  assert "stale evidence is diagnosed distinctly" (containsStale staleTrace)
  assert "requirement interpreter sees both authorization paths" $
    requirements mergeRule
      == Set.fromList
        [ ExactHead,
          CheckPassed "unit",
          CheckPassed "lint",
          GateReceipt,
          HumanJudgment
        ]
  putStrLn "assurance-dsl: all tests passed"

containsStale :: Trace -> Bool
containsStale (Leaf _ _ (Stale _ _)) = True
containsStale (Leaf _ _ _) = False
containsStale (Both _ left right) = containsStale left || containsStale right
containsStale (EitherPath _ left right) = containsStale left || containsStale right

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
