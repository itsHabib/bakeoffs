module Example
  ( reviewPlan,
    mergePlan,
    fixture,
    readOnlyGrant,
    exactReviewGrant,
    exactMergeGrant,
  )
where

import qualified Data.Set as Set
import Plan
  ( Capability (..),
    Fixture (..),
    Grant (..),
    Mutation (..),
    Plan,
    Resource (..),
    mergeP,
    postCommentP,
    readFileP,
    runCheckP,
  )

reviewPlan :: Plan (String, Bool, ())
reviewPlan =
  (,,)
    <$> readFileP "src/Gate.hs"
    <*> runCheckP "unit"
    <*> postCommentP 42 "preflight passed"

mergePlan :: Plan (String, Bool, (), ())
mergePlan =
  (,,,)
    <$> readFileP "src/Gate.hs"
    <*> runCheckP "unit"
    <*> postCommentP 42 "preflight passed"
    <*> mergeP 42

fixture :: Fixture
fixture =
  Fixture
    { fixtureFiles = [("src/Gate.hs", "module Gate where")],
      fixtureChecks = [("unit", True)],
      fixtureComments = [],
      fixtureMerges = [],
      fixtureLog = []
    }

readOnlyGrant :: Grant
readOnlyGrant =
  Grant
    { grantedCapabilities = Set.fromList [ReadRepository, RunChecks],
      grantedResources = Set.fromList [RepositoryPath "src/Gate.hs", CheckName "unit"],
      maximumMutation = ReadOnly
    }

exactReviewGrant :: Grant
exactReviewGrant =
  Grant
    { grantedCapabilities =
        Set.fromList [ReadRepository, RunChecks, CommentOnPullRequest],
      grantedResources =
        Set.fromList
          [ RepositoryPath "src/Gate.hs",
            CheckName "unit",
            PullRequest 42
          ],
      maximumMutation = ExternalWrite
    }

exactMergeGrant :: Grant
exactMergeGrant =
  exactReviewGrant
    { grantedCapabilities =
        Set.insert MergePullRequest (grantedCapabilities exactReviewGrant),
      maximumMutation = RepositoryMerge
    }
