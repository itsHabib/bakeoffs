{-# LANGUAGE GADTs #-}

module Capability
  ( Plan,
    readRepo,
    runCheck,
    commentOn,
    merge,
    Capability (..),
    Resource (..),
    Mutation (..),
    Envelope (..),
    analyze,
    renderPlan,
    Grant (..),
    Denial (..),
    authorize,
    Fixture (..),
    Execution (..),
    execute,
  )
where

import qualified Data.Set as Set
import Data.Set (Set)

data Operation a where
  ReadRepo :: FilePath -> Operation String
  RunCheck :: String -> Operation Bool
  CommentOn :: Int -> String -> Operation ()
  Merge :: Int -> Operation String

data Plan a where
  Pure :: a -> Plan a
  Step :: Operation a -> Plan a
  Ap :: Plan (a -> b) -> Plan a -> Plan b

instance Functor Plan where
  fmap function value = Pure function `Ap` value

instance Applicative Plan where
  pure = Pure
  (<*>) = Ap

readRepo :: FilePath -> Plan String
readRepo = Step . ReadRepo

runCheck :: String -> Plan Bool
runCheck = Step . RunCheck

commentOn :: Int -> String -> Plan ()
commentOn pullRequest = Step . CommentOn pullRequest

merge :: Int -> Plan String
merge = Step . Merge

data Capability
  = ReadRepository
  | ExecuteCheck
  | WriteComment
  | MergePullRequest
  deriving (Eq, Ord, Show)

data Resource
  = Path FilePath
  | Check String
  | PullRequest Int
  deriving (Eq, Ord, Show)

data Mutation = Observe | Communicate | LandCode
  deriving (Eq, Ord, Show)

data Envelope = Envelope
  { requiredCapabilities :: Set Capability,
    requiredResources :: Set Resource,
    maximumMutation :: Mutation,
    operationCount :: Int
  }
  deriving (Eq, Show)

instance Semigroup Envelope where
  left <> right =
    Envelope
      (Set.union (requiredCapabilities left) (requiredCapabilities right))
      (Set.union (requiredResources left) (requiredResources right))
      (max (maximumMutation left) (maximumMutation right))
      (operationCount left + operationCount right)

instance Monoid Envelope where
  mempty = Envelope Set.empty Set.empty Observe 0

analyze :: Plan a -> Envelope
analyze (Pure _) = mempty
analyze (Step operation) = operationEnvelope operation
analyze (Ap function argument) = analyze function <> analyze argument

renderPlan :: Plan a -> [String]
renderPlan (Pure _) = []
renderPlan (Step operation) = [operationLabel operation]
renderPlan (Ap function argument) = renderPlan function ++ renderPlan argument

operationEnvelope :: Operation a -> Envelope
operationEnvelope operation =
  case operation of
    ReadRepo path -> singleton ReadRepository (Path path) Observe
    RunCheck name -> singleton ExecuteCheck (Check name) Observe
    CommentOn pullRequest _ -> singleton WriteComment (PullRequest pullRequest) Communicate
    Merge pullRequest -> singleton MergePullRequest (PullRequest pullRequest) LandCode
  where
    singleton capability resource mutation =
      Envelope (Set.singleton capability) (Set.singleton resource) mutation 1

operationLabel :: Operation a -> String
operationLabel operation =
  case operation of
    ReadRepo path -> "read " ++ path
    RunCheck name -> "check " ++ name
    CommentOn pullRequest _ -> "comment PR-" ++ show pullRequest
    Merge pullRequest -> "merge PR-" ++ show pullRequest

data Grant = Grant
  { grantedCapabilities :: Set Capability,
    grantedResources :: Set Resource,
    mutationCeiling :: Mutation
  }
  deriving (Eq, Show)

data Denial = Denial
  { missingCapabilities :: Set Capability,
    missingResources :: Set Resource,
    requestedMutation :: Mutation,
    allowedMutation :: Mutation
  }
  deriving (Eq, Show)

authorize :: Grant -> Plan a -> Either Denial ()
authorize grant plan
  | Set.null absentCapabilities
      && Set.null absentResources
      && maximumMutation envelope <= mutationCeiling grant = Right ()
  | otherwise =
      Left
        ( Denial
            absentCapabilities
            absentResources
            (maximumMutation envelope)
            (mutationCeiling grant)
        )
  where
    envelope = analyze plan
    absentCapabilities =
      Set.difference (requiredCapabilities envelope) (grantedCapabilities grant)
    absentResources = Set.difference (requiredResources envelope) (grantedResources grant)

data Fixture = Fixture
  { files :: [(FilePath, String)],
    checks :: [(String, Bool)],
    comments :: [(Int, String)],
    merges :: [Int],
    effectLog :: [String]
  }
  deriving (Eq, Show)

data Execution a = Execution a Fixture
  deriving (Eq, Show)

execute :: Grant -> Fixture -> Plan a -> Either Denial (Execution a)
execute grant fixture plan = do
  authorize grant plan
  let (value, final) = interpret fixture plan
  Right (Execution value final)

interpret :: Fixture -> Plan a -> (a, Fixture)
interpret fixture (Pure value) = (value, fixture)
interpret fixture (Step operation) = perform fixture operation
interpret fixture (Ap functionPlan argumentPlan) =
  let (function, afterFunction) = interpret fixture functionPlan
      (argument, afterArgument) = interpret afterFunction argumentPlan
   in (function argument, afterArgument)

perform :: Fixture -> Operation a -> (a, Fixture)
perform fixture operation =
  case operation of
    ReadRepo path ->
      (maybe "<missing>" id (lookup path (files fixture)), record operation fixture)
    RunCheck name ->
      (maybe False id (lookup name (checks fixture)), record operation fixture)
    CommentOn pullRequest body ->
      ((), (record operation fixture) {comments = comments fixture ++ [(pullRequest, body)]})
    Merge pullRequest ->
      ( "merged-" ++ show pullRequest,
        (record operation fixture) {merges = merges fixture ++ [pullRequest]}
      )

record :: Operation a -> Fixture -> Fixture
record operation fixture = fixture {effectLog = effectLog fixture ++ [operationLabel operation]}
