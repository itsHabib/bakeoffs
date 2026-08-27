module Example
  ( reviewPlan,
    mergePlan,
    fixture,
    readOnlyGrant,
    reviewGrant,
    mergeGrant,
  )
where

import Capability
  ( Capability (..),
    Fixture (..),
    Grant (..),
    Mutation (..),
    Plan,
    Resource (..),
    commentOn,
    merge,
    readRepo,
    runCheck,
  )
import qualified Data.Set as Set

reviewPlan :: Plan (String, Bool, ())
reviewPlan =
  (,,)
    <$> readRepo "src/Gate.hs"
    <*> runCheck "unit"
    <*> commentOn 42 "ready"

mergePlan :: Plan (String, Bool, (), String)
mergePlan =
  (,,,)
    <$> readRepo "src/Gate.hs"
    <*> runCheck "unit"
    <*> commentOn 42 "ready"
    <*> merge 42

fixture :: Fixture
fixture = Fixture [("src/Gate.hs", "module Gate where")] [("unit", True)] [] [] []

readOnlyGrant :: Grant
readOnlyGrant =
  Grant
    (Set.fromList [ReadRepository, ExecuteCheck])
    (Set.fromList [Path "src/Gate.hs", Check "unit"])
    Observe

reviewGrant :: Grant
reviewGrant =
  Grant
    (Set.fromList [ReadRepository, ExecuteCheck, WriteComment])
    (Set.fromList [Path "src/Gate.hs", Check "unit", PullRequest 42])
    Communicate

mergeGrant :: Grant
mergeGrant =
  reviewGrant
    { grantedCapabilities = Set.insert MergePullRequest (grantedCapabilities reviewGrant),
      mutationCeiling = LandCode
    }
