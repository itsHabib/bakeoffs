module Main (main) where

import Capability
  ( Capability (..),
    Denial (..),
    Execution (..),
    Fixture (..),
    Mutation (..),
    analyze,
    effectLog,
    execute,
    maximumMutation,
    merges,
    operationCount,
    renderPlan,
  )
import Control.Monad (unless)
import qualified Data.Set as Set
import Example (fixture, mergeGrant, mergePlan, readOnlyGrant, reviewGrant, reviewPlan)

main :: IO ()
main = do
  assert "analysis retains every operation" (operationCount (analyze reviewPlan) == 3)
  case execute readOnlyGrant fixture reviewPlan of
    Left denial -> do
      assert "only comment capability is missing" $
        missingCapabilities denial == Set.singleton WriteComment
      assert "denied plan has zero effects" (effectLog fixture == [])
    Right _ -> fail "read-only grant executed a write"
  case execute reviewGrant fixture reviewPlan of
    Left denial -> fail (show denial)
    Right (Execution _ final) ->
      assert "dry-run order matches effects" (renderPlan reviewPlan == effectLog final)
  assert "merge is a distinct tier" (maximumMutation (analyze mergePlan) == LandCode)
  case execute reviewGrant fixture mergePlan of
    Left denial ->
      assert "merge capability is precisely missing" $
        missingCapabilities denial == Set.singleton MergePullRequest
    Right _ -> fail "review grant merged"
  case execute mergeGrant fixture mergePlan of
    Right (Execution _ final) -> assert "merge happens once" (merges final == [42])
    Left denial -> fail (show denial)
  putStrLn "capability-dsl: all tests passed"

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
