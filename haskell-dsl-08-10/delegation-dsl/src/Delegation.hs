module Delegation
  ( Role (..),
    Artifact (..),
    RoleSpec,
    agent,
    requires,
    produces,
    andProduces,
    contract,
    Contract,
    Problem (..),
    Brief (..),
    Compiled (..),
    validate,
    compile,
    renderBrief,
  )
where

import Data.List (intercalate, nub, sort)
import qualified Data.Map.Strict as Map
import Data.Map.Strict (Map)
import qualified Data.Set as Set

newtype Role = Role String
  deriving (Eq, Ord, Show)

newtype Artifact = Artifact String
  deriving (Eq, Ord, Show)

data RoleSpec = RoleSpec
  { specRole :: Role,
    specRequires :: [Artifact],
    specProduces :: [Artifact]
  }
  deriving (Eq, Show)

newtype Contract = Contract [RoleSpec]
  deriving (Eq, Show)

agent :: String -> RoleSpec
agent name = RoleSpec (Role name) [] []

requires :: RoleSpec -> Artifact -> RoleSpec
requires spec artifact = spec {specRequires = specRequires spec ++ [artifact]}

produces :: RoleSpec -> Artifact -> RoleSpec
produces spec artifact = spec {specProduces = specProduces spec ++ [artifact]}

andProduces :: RoleSpec -> Artifact -> RoleSpec
andProduces = produces

contract :: [RoleSpec] -> Contract
contract = Contract

data Problem
  = DuplicateRole Role
  | MissingProducer Role Artifact
  | AmbiguousProducer Artifact [Role]
  | SelfRequirement Role Artifact
  | DependencyCycle [Role]
  deriving (Eq, Show)

data Brief = Brief
  { briefRole :: Role,
    receives :: [(Artifact, Role)],
    delivers :: [(Artifact, [Role])]
  }
  deriving (Eq, Show)

data Compiled = Compiled
  { roleBriefs :: [Brief],
    executionBatches :: [[Role]]
  }
  deriving (Eq, Show)

validate :: Contract -> [Problem]
validate contractValue@(Contract specs) =
  duplicateProblems ++ artifactProblems ++ cycleProblems
  where
    duplicateProblems = map DuplicateRole (duplicates (map specRole specs))
    producerIndex = producersByArtifact specs
    artifactProblems = concatMap (validateRequirements producerIndex) specs
    cycleProblems
      | null duplicateProblems && null artifactProblems =
          case batchesFor contractValue of
            Left cycleRoles -> [DependencyCycle cycleRoles]
            Right _ -> []
      | otherwise = []

compile :: Contract -> Either [Problem] Compiled
compile contractValue@(Contract specs) =
  case validate contractValue of
    [] ->
      case batchesFor contractValue of
        Left roles -> Left [DependencyCycle roles]
        Right batches -> Right (Compiled (map (makeBrief specs) specs) batches)
    problems -> Left problems

producersByArtifact :: [RoleSpec] -> Map Artifact [Role]
producersByArtifact specs =
  Map.fromListWith (++)
    [ (artifact, [specRole spec])
      | spec <- specs,
        artifact <- specProduces spec
    ]

validateRequirements :: Map Artifact [Role] -> RoleSpec -> [Problem]
validateRequirements producerIndex spec = concatMap validateOne (specRequires spec)
  where
    validateOne artifact =
      case sort (Map.findWithDefault [] artifact producerIndex) of
        [] -> [MissingProducer (specRole spec) artifact]
        [producer]
          | producer == specRole spec -> [SelfRequirement producer artifact]
          | otherwise -> []
        roles -> [AmbiguousProducer artifact roles]

makeBrief :: [RoleSpec] -> RoleSpec -> Brief
makeBrief specs spec =
  Brief
    { briefRole = specRole spec,
      receives =
        [ (artifact, soleProducer artifact)
          | artifact <- specRequires spec
        ],
      delivers =
        [ (artifact, consumers artifact)
          | artifact <- specProduces spec
        ]
    }
  where
    producerIndex = producersByArtifact specs
    soleProducer artifact =
      case Map.findWithDefault [] artifact producerIndex of
        [role] -> role
        _ -> error "makeBrief called before successful validation"
    consumers artifact =
      sort
        [ specRole candidate
          | candidate <- specs,
            artifact `elem` specRequires candidate
        ]

batchesFor :: Contract -> Either [Role] [[Role]]
batchesFor (Contract specs) = go Set.empty allRoles []
  where
    allRoles = Set.fromList (map specRole specs)
    producerIndex = producersByArtifact specs
    dependencies =
      Map.fromList
        [ (specRole spec, Set.fromList (concatMap producersOf (specRequires spec)))
          | spec <- specs
        ]
    producersOf artifact = Map.findWithDefault [] artifact producerIndex
    go completed remaining batches
      | Set.null remaining = Right (reverse batches)
      | null readyRoles = Left (Set.toAscList remaining)
      | otherwise =
          let readySet = Set.fromList readyRoles
           in go
                (Set.union completed readySet)
                (Set.difference remaining readySet)
                (readyRoles : batches)
      where
        readyRoles =
          [ role
            | role <- Set.toAscList remaining,
              Map.findWithDefault Set.empty role dependencies `Set.isSubsetOf` completed
          ]

duplicates :: (Ord a) => [a] -> [a]
duplicates values =
  [ value
    | value <- nub values,
      length (filter (== value) values) > 1
  ]

renderBrief :: Brief -> String
renderBrief brief =
  unlines
    [ "ROLE " ++ renderRole (briefRole brief),
      "RECEIVES " ++ renderPairs (receives brief),
      "DELIVERS " ++ renderDeliveries (delivers brief)
    ]
  where
    renderRole (Role value) = value
    renderArtifact (Artifact value) = value
    renderPairs [] = "nothing"
    renderPairs values =
      intercalate ", "
        [ renderArtifact artifact ++ " from " ++ renderRole role
          | (artifact, role) <- values
        ]
    renderDeliveries [] = "nothing"
    renderDeliveries values =
      intercalate ", "
        [ renderArtifact artifact
            ++ " to ["
            ++ intercalate ", " (map renderRole roles)
            ++ "]"
          | (artifact, roles) <- values
        ]
