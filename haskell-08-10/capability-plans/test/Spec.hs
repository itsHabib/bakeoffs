module Main (main) where

import qualified Data.Set as Set
import Control.Monad (unless)
import Example
  ( exactMergeGrant,
    exactReviewGrant,
    fixture,
    mergePlan,
    readOnlyGrant,
    reviewPlan,
  )
import Plan
  ( AuthorizationError (..),
    Capability (..),
    Envelope (..),
    Execution (..),
    Fixture (..),
    Mutation (..),
    analyze,
    executeAuthorized,
    operationOrder,
  )

main :: IO ()
main = do
  let reviewEnvelope = analyze reviewPlan
  assert "analysis sees all independent operations" (cost reviewEnvelope == 3)
  assert "analysis sees external mutation" (mutation reviewEnvelope == ExternalWrite)
  assert "analysis sees comment capability" $
    CommentOnPullRequest `Set.member` capabilities reviewEnvelope
  checkPreflightIsAtomic
  checkExactGrant
  checkMergeBoundary
  putStrLn "capability-plans: all tests passed"

checkPreflightIsAtomic :: IO ()
checkPreflightIsAtomic =
  case executeAuthorized readOnlyGrant fixture reviewPlan of
    Left problem -> do
      assert "missing comment capability is explicit" $
        missingCapabilities problem == Set.singleton CommentOnPullRequest
      assert "mutation ceiling is explicit" $
        requiredMutation problem == ExternalWrite
          && grantedMutation problem == ReadOnly
      assert "original fixture has no effects" (fixtureLog fixture == [])
    Right _ -> fail "read-only grant unexpectedly executed review plan"

checkExactGrant :: IO ()
checkExactGrant =
  case executeAuthorized exactReviewGrant fixture reviewPlan of
    Left problem -> fail ("exact grant refused: " ++ show problem)
    Right execution -> do
      let final = finalFixture execution
      assert "dry-run order equals execution order" $
        operationOrder reviewPlan == fixtureLog final
      assert "one comment is posted" $
        fixtureComments final == [(42, "preflight passed")]

checkMergeBoundary :: IO ()
checkMergeBoundary = do
  let envelope = analyze mergePlan
  assert "merge uses strongest mutation class" (mutation envelope == RepositoryMerge)
  assert "merge capability is present" $
    MergePullRequest `Set.member` capabilities envelope
  case executeAuthorized exactReviewGrant fixture mergePlan of
    Left problem ->
      assert "review grant lacks precisely merge" $
        missingCapabilities problem == Set.singleton MergePullRequest
    Right _ -> fail "review grant unexpectedly authorized merge"
  case executeAuthorized exactMergeGrant fixture mergePlan of
    Left problem -> fail ("merge grant refused: " ++ show problem)
    Right execution ->
      assert "merge executes once" (fixtureMerges (finalFixture execution) == [42])

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
