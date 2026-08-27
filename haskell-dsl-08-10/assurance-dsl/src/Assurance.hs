module Assurance
  ( Revision (..),
    Subject (..),
    Requirement (..),
    Observation (..),
    Evidence (..),
    Policy,
    exactHead,
    checkPassed,
    gateReceipt,
    humanJudgment,
    (.&&.),
    (.||.),
    Rule,
    ready,
    requirements,
    Status (..),
    Finding (..),
    Trace (..),
    evaluate,
    traceStatus,
    renderTrace,
  )
where

import Data.List (find, intercalate)
import qualified Data.Set as Set
import Data.Set (Set)

newtype Revision = Revision String
  deriving (Eq, Ord, Show)

data Subject = Subject
  { subjectName :: String,
    subjectRevision :: Revision
  }
  deriving (Eq, Show)

data Requirement
  = ExactHead
  | CheckPassed String
  | GateReceipt
  | HumanJudgment
  deriving (Eq, Ord, Show)

data Observation = Observation
  { observedRequirement :: Requirement,
    observedRevision :: Revision,
    observationPassed :: Bool,
    observationSource :: String
  }
  deriving (Eq, Show)

data Evidence = Evidence
  { currentHead :: Revision,
    observations :: [Observation]
  }
  deriving (Eq, Show)

data Policy
  = Require Requirement
  | And Policy Policy
  | Or Policy Policy
  deriving (Eq, Show)

data Rule = Rule
  { actionName :: String,
    rulePolicy :: Policy
  }
  deriving (Eq, Show)

exactHead :: Policy
exactHead = Require ExactHead

checkPassed :: String -> Policy
checkPassed = Require . CheckPassed

gateReceipt :: Policy
gateReceipt = Require GateReceipt

humanJudgment :: Policy
humanJudgment = Require HumanJudgment

infixr 3 .&&.

(.&&.) :: Policy -> Policy -> Policy
(.&&.) = And

infixr 2 .||.

(.||.) :: Policy -> Policy -> Policy
(.||.) = Or

ready :: String -> Policy -> Rule
ready = Rule

requirements :: Rule -> Set Requirement
requirements = collect . rulePolicy
  where
    collect (Require requirement) = Set.singleton requirement
    collect (And left right) = Set.union (collect left) (collect right)
    collect (Or left right) = Set.union (collect left) (collect right)

data Status = Passed | Failed
  deriving (Eq, Show)

data Finding
  = Proven Requirement String
  | Missing Requirement
  | Rejected Requirement String
  | Stale Requirement [Revision]
  | HeadMismatch Revision Revision
  deriving (Eq, Show)

data Trace
  = Leaf Requirement Status Finding
  | Both Status Trace Trace
  | EitherPath Status Trace Trace
  deriving (Eq, Show)

evaluate :: Subject -> Evidence -> Rule -> Trace
evaluate subject evidence = go . rulePolicy
  where
    go (Require requirement) = evaluateRequirement subject evidence requirement
    go (And left right) =
      let leftTrace = go left
          rightTrace = go right
       in Both (andStatus (traceStatus leftTrace) (traceStatus rightTrace)) leftTrace rightTrace
    go (Or left right) =
      let leftTrace = go left
          rightTrace = go right
       in EitherPath (orStatus (traceStatus leftTrace) (traceStatus rightTrace)) leftTrace rightTrace

traceStatus :: Trace -> Status
traceStatus (Leaf _ status _) = status
traceStatus (Both status _ _) = status
traceStatus (EitherPath status _ _) = status

evaluateRequirement :: Subject -> Evidence -> Requirement -> Trace
evaluateRequirement subject evidence ExactHead
  | currentHead evidence == subjectRevision subject =
      Leaf ExactHead Passed (Proven ExactHead "repository head matches subject")
  | otherwise =
      Leaf ExactHead Failed (HeadMismatch (subjectRevision subject) (currentHead evidence))
evaluateRequirement subject evidence requirement =
  case find observationPassed matching of
    Just passed ->
      Leaf requirement Passed (Proven requirement (observationSource passed))
    Nothing ->
      case matching of
        failed : _ ->
          Leaf requirement Failed (Rejected requirement (observationSource failed))
        [] ->
          case staleRevisions of
            [] -> Leaf requirement Failed (Missing requirement)
            revisions -> Leaf requirement Failed (Stale requirement revisions)
  where
    relevant = filter ((== requirement) . observedRequirement) (observations evidence)
    matching = filter ((== subjectRevision subject) . observedRevision) relevant
    staleRevisions = Set.toAscList (Set.fromList (map observedRevision relevant))

andStatus :: Status -> Status -> Status
andStatus Passed Passed = Passed
andStatus _ _ = Failed

orStatus :: Status -> Status -> Status
orStatus Failed Failed = Failed
orStatus _ _ = Passed

renderTrace :: Trace -> String
renderTrace = unlines . render 0
  where
    render depth (Leaf requirement status finding) =
      [indent depth ++ show status ++ " " ++ show requirement ++ ": " ++ renderFinding finding]
    render depth (Both status left right) =
      (indent depth ++ show status ++ " ALL") : render (depth + 1) left ++ render (depth + 1) right
    render depth (EitherPath status left right) =
      (indent depth ++ show status ++ " ANY") : render (depth + 1) left ++ render (depth + 1) right
    indent depth = replicate (depth * 2) ' '
    renderFinding (Proven _ source) = "proven by " ++ source
    renderFinding (Missing _) = "missing"
    renderFinding (Rejected _ source) = "explicitly failed at " ++ source
    renderFinding (Stale _ revisions) = "stale at " ++ intercalate ", " (map show revisions)
    renderFinding (HeadMismatch expected actual) =
      "expected " ++ show expected ++ ", current " ++ show actual
