{-# LANGUAGE GADTs #-}

module Plan
  ( Operation (..),
    Plan,
    readFileP,
    runCheckP,
    postCommentP,
    mergeP,
    Capability (..),
    Resource (..),
    Mutation (..),
    Envelope (..),
    analyze,
    operationOrder,
    Grant (..),
    AuthorizationError (..),
    authorize,
    Fixture (..),
    Execution (..),
    executeAuthorized,
  )
where

import qualified Data.Set as Set
import Data.Set (Set)

data Operation a where
  ReadFile :: FilePath -> Operation String
  RunCheck :: String -> Operation Bool
  PostComment :: Int -> String -> Operation ()
  Merge :: Int -> Operation ()

data Plan a where
  Pure :: a -> Plan a
  Step :: Operation a -> Plan a
  Ap :: Plan (a -> b) -> Plan a -> Plan b

instance Functor Plan where
  fmap function plan = Pure function `Ap` plan

instance Applicative Plan where
  pure = Pure
  (<*>) = Ap

readFileP :: FilePath -> Plan String
readFileP = Step . ReadFile

runCheckP :: String -> Plan Bool
runCheckP = Step . RunCheck

postCommentP :: Int -> String -> Plan ()
postCommentP pullRequest = Step . PostComment pullRequest

mergeP :: Int -> Plan ()
mergeP = Step . Merge

data Capability
  = ReadRepository
  | RunChecks
  | CommentOnPullRequest
  | MergePullRequest
  deriving (Eq, Ord, Show)

data Resource
  = RepositoryPath FilePath
  | CheckName String
  | PullRequest Int
  deriving (Eq, Ord, Show)

data Mutation
  = ReadOnly
  | ExternalWrite
  | RepositoryMerge
  deriving (Eq, Ord, Show)

data Envelope = Envelope
  { capabilities :: Set Capability,
    resources :: Set Resource,
    mutation :: Mutation,
    cost :: Int
  }
  deriving (Eq, Show)

instance Semigroup Envelope where
  left <> right =
    Envelope
      { capabilities = Set.union (capabilities left) (capabilities right),
        resources = Set.union (resources left) (resources right),
        mutation = max (mutation left) (mutation right),
        cost = cost left + cost right
      }

instance Monoid Envelope where
  mempty = Envelope Set.empty Set.empty ReadOnly 0

analyze :: Plan a -> Envelope
analyze (Pure _) = mempty
analyze (Step operation) = operationEnvelope operation
analyze (Ap function argument) = analyze function <> analyze argument

operationOrder :: Plan a -> [String]
operationOrder (Pure _) = []
operationOrder (Step operation) = [operationLabel operation]
operationOrder (Ap function argument) = operationOrder function ++ operationOrder argument

operationEnvelope :: Operation a -> Envelope
operationEnvelope operation =
  case operation of
    ReadFile path -> singleton ReadRepository (RepositoryPath path) ReadOnly
    RunCheck name -> singleton RunChecks (CheckName name) ReadOnly
    PostComment pullRequest _ ->
      singleton CommentOnPullRequest (PullRequest pullRequest) ExternalWrite
    Merge pullRequest ->
      singleton MergePullRequest (PullRequest pullRequest) RepositoryMerge
  where
    singleton capability resource mutationValue =
      Envelope (Set.singleton capability) (Set.singleton resource) mutationValue 1

operationLabel :: Operation a -> String
operationLabel operation =
  case operation of
    ReadFile path -> "read:" ++ path
    RunCheck name -> "check:" ++ name
    PostComment pullRequest _ -> "comment:pr-" ++ show pullRequest
    Merge pullRequest -> "merge:pr-" ++ show pullRequest

data Grant = Grant
  { grantedCapabilities :: Set Capability,
    grantedResources :: Set Resource,
    maximumMutation :: Mutation
  }
  deriving (Eq, Show)

data AuthorizationError = AuthorizationError
  { missingCapabilities :: Set Capability,
    missingResources :: Set Resource,
    requiredMutation :: Mutation,
    grantedMutation :: Mutation
  }
  deriving (Eq, Show)

authorize :: Grant -> Envelope -> Either AuthorizationError ()
authorize grant envelope
  | Set.null absentCapabilities
      && Set.null absentResources
      && mutation envelope <= maximumMutation grant = Right ()
  | otherwise =
      Left
        AuthorizationError
          { missingCapabilities = absentCapabilities,
            missingResources = absentResources,
            requiredMutation = mutation envelope,
            grantedMutation = maximumMutation grant
          }
  where
    absentCapabilities = Set.difference (capabilities envelope) (grantedCapabilities grant)
    absentResources = Set.difference (resources envelope) (grantedResources grant)

data Fixture = Fixture
  { fixtureFiles :: [(FilePath, String)],
    fixtureChecks :: [(String, Bool)],
    fixtureComments :: [(Int, String)],
    fixtureMerges :: [Int],
    fixtureLog :: [String]
  }
  deriving (Eq, Show)

data Execution a = Execution
  { result :: a,
    finalFixture :: Fixture
  }
  deriving (Eq, Show)

executeAuthorized :: Grant -> Fixture -> Plan a -> Either AuthorizationError (Execution a)
executeAuthorized grant fixture plan = do
  authorize grant (analyze plan)
  Right (uncurry Execution (execute fixture plan))

execute :: Fixture -> Plan a -> (a, Fixture)
execute fixture (Pure value) = (value, fixture)
execute fixture (Step operation) = executeOperation fixture operation
execute fixture (Ap functionPlan argumentPlan) =
  let (function, afterFunction) = execute fixture functionPlan
      (argument, afterArgument) = execute afterFunction argumentPlan
   in (function argument, afterArgument)

executeOperation :: Fixture -> Operation a -> (a, Fixture)
executeOperation fixture operation =
  case operation of
    ReadFile path ->
      ( maybe "<missing>" id (lookup path (fixtureFiles fixture)),
        record operation fixture
      )
    RunCheck name ->
      ( maybe False id (lookup name (fixtureChecks fixture)),
        record operation fixture
      )
    PostComment pullRequest body ->
      ( (),
        (record operation fixture)
          { fixtureComments = fixtureComments fixture ++ [(pullRequest, body)]
          }
      )
    Merge pullRequest ->
      ( (),
        (record operation fixture)
          { fixtureMerges = fixtureMerges fixture ++ [pullRequest]
          }
      )

record :: Operation a -> Fixture -> Fixture
record operation fixture =
  fixture {fixtureLog = fixtureLog fixture ++ [operationLabel operation]}
