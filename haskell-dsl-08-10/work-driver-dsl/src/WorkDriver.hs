module WorkDriver
  ( TaskRef,
    TaskKind (..),
    ProjectM,
    task,
    validate,
    land,
    after,
    afterAll,
    touches,
    parallel,
    Project,
    project,
    CompileError (..),
    ParallelDecision (..),
    TaskView (..),
    DriverPlan (..),
    compile,
    renderManifest,
  )
where

import Data.List (find, intercalate, maximumBy, sortOn)
import qualified Data.Set as Set
import Data.Set (Set)
import Data.Ord (comparing)

newtype TaskRef = TaskRef Int
  deriving (Eq, Ord, Show)

data TaskKind = Work | Validation | Landing
  deriving (Eq, Show)

data Node = Node
  { nodeRef :: TaskRef,
    nodeName :: String,
    nodeKind :: TaskKind,
    nodeDependencies :: Set TaskRef,
    nodeScopes :: [FilePath]
  }
  deriving (Eq, Show)

data BuildState = BuildState
  { nextRef :: Int,
    builtNodes :: [Node],
    parallelRequests :: [(TaskRef, TaskRef)]
  }

newtype ProjectM a = ProjectM
  { runProjectM :: BuildState -> (a, BuildState)
  }

instance Functor ProjectM where
  fmap function action = ProjectM $ \state ->
    let (result, nextState) = runProjectM action state
     in (function result, nextState)

instance Applicative ProjectM where
  pure result = ProjectM (\state -> (result, state))
  functionAction <*> argumentAction = ProjectM $ \state ->
    let (function, afterFunction) = runProjectM functionAction state
        (argument, afterArgument) = runProjectM argumentAction afterFunction
     in (function argument, afterArgument)

instance Monad ProjectM where
  action >>= next = ProjectM $ \state ->
    let (result, nextState) = runProjectM action state
     in runProjectM (next result) nextState

task :: String -> ProjectM TaskRef
task label = addNode label Work

validate :: String -> ProjectM TaskRef
validate label = addNode label Validation

land :: String -> ProjectM TaskRef
land label = addNode label Landing

addNode :: String -> TaskKind -> ProjectM TaskRef
addNode label kindValue = ProjectM $ \state ->
  let reference = TaskRef (nextRef state)
      node = Node reference label kindValue Set.empty []
   in ( reference,
        state
          { nextRef = nextRef state + 1,
            builtNodes = builtNodes state ++ [node]
          }
      )

infixl 5 `after`

after :: ProjectM TaskRef -> TaskRef -> ProjectM TaskRef
after action dependency = do
  reference <- action
  modifyNode reference $ \node ->
    node {nodeDependencies = Set.insert dependency (nodeDependencies node)}
  pure reference

infixl 5 `afterAll`

afterAll :: ProjectM TaskRef -> [TaskRef] -> ProjectM TaskRef
afterAll action prerequisites = do
  reference <- action
  modifyNode reference $ \node ->
    node {nodeDependencies = Set.union (Set.fromList prerequisites) (nodeDependencies node)}
  pure reference

infixl 4 `touches`

touches :: ProjectM TaskRef -> [FilePath] -> ProjectM TaskRef
touches action paths = do
  reference <- action
  modifyNode reference $ \node -> node {nodeScopes = nodeScopes node ++ paths}
  pure reference

parallel :: ProjectM TaskRef -> ProjectM TaskRef -> ProjectM (TaskRef, TaskRef)
parallel left right = do
  leftRef <- left
  rightRef <- right
  ProjectM $ \state ->
    ( (leftRef, rightRef),
      state {parallelRequests = parallelRequests state ++ [(leftRef, rightRef)]}
    )

modifyNode :: TaskRef -> (Node -> Node) -> ProjectM ()
modifyNode reference change = ProjectM $ \state ->
  ((), state {builtNodes = map update (builtNodes state)})
  where
    update node
      | nodeRef node == reference = change node
      | otherwise = node

data Project = Project
  { projectName :: String,
    projectNodes :: [Node],
    requestedParallel :: [(TaskRef, TaskRef)]
  }
  deriving (Eq, Show)

project :: String -> ProjectM a -> Project
project projectLabel body =
  let (_, final) = runProjectM body (BuildState 0 [] [])
   in Project projectLabel (builtNodes final) (parallelRequests final)

data CompileError
  = DuplicateTaskName String
  | UngatedLanding String
  | InternalCycle [String]
  deriving (Eq, Show)

data ParallelDecision
  = ParallelAccepted String String
  | ParallelSerialized String String [FilePath]
  | ParallelDeferred String String
  deriving (Eq, Show)

data TaskView = TaskView
  { taskName :: String,
    taskKind :: TaskKind,
    dependencies :: [String],
    scopes :: [FilePath]
  }
  deriving (Eq, Show)

data DriverPlan = DriverPlan
  { name :: String,
    batches :: [[TaskView]],
    criticalPath :: [String],
    parallelDecisions :: [ParallelDecision]
  }
  deriving (Eq, Show)

compile :: Project -> Either [CompileError] DriverPlan
compile projectValue =
  case errors of
    [] ->
      case schedule nodes of
        Left remaining -> Left [InternalCycle (map nodeName remaining)]
        Right scheduled ->
          Right
            DriverPlan
              { name = projectName projectValue,
                batches = map (map toView) scheduled,
                criticalPath = map nodeName (longestPath nodes),
                parallelDecisions =
                  map (parallelDecision nodes scheduled) (requestedParallel projectValue)
              }
    values -> Left values
  where
    nodes = projectNodes projectValue
    errors = duplicateNameErrors nodes ++ ungatedErrors nodes
    toView node =
      TaskView
        (nodeName node)
        (nodeKind node)
        (map (nameFor nodes) (Set.toAscList (nodeDependencies node)))
        (nodeScopes node)

duplicateNameErrors :: [Node] -> [CompileError]
duplicateNameErrors nodes =
  [ DuplicateTaskName duplicate
    | duplicate <- Set.toAscList duplicates
  ]
  where
    names = map nodeName nodes
    duplicates =
      Set.fromList
        [ candidate
          | candidate <- names,
            length (filter (== candidate) names) > 1
        ]

ungatedErrors :: [Node] -> [CompileError]
ungatedErrors nodes =
  [ UngatedLanding (nodeName node)
    | node <- nodes,
      nodeKind node == Landing,
      not (any ((== Validation) . nodeKind) (ancestors nodes node))
  ]

ancestors :: [Node] -> Node -> [Node]
ancestors nodes node = go Set.empty (Set.toList (nodeDependencies node))
  where
    go _ [] = []
    go seen (reference : rest)
      | reference `Set.member` seen = go seen rest
      | otherwise =
          case find ((== reference) . nodeRef) nodes of
            Nothing -> go (Set.insert reference seen) rest
            Just dependency ->
              dependency
                : go
                  (Set.insert reference seen)
                  (Set.toList (nodeDependencies dependency) ++ rest)

schedule :: [Node] -> Either [Node] [[Node]]
schedule nodes = go Set.empty (sortOn nodeRef nodes) []
  where
    go _ [] completedBatches = Right (reverse completedBatches)
    go completed remaining completedBatches =
      let ready = filter ((`Set.isSubsetOf` completed) . nodeDependencies) remaining
          selected = greedilyCompatible ready
       in if null selected
            then Left remaining
            else
              let selectedRefs = Set.fromList (map nodeRef selected)
                  nextRemaining = filter ((`Set.notMember` selectedRefs) . nodeRef) remaining
               in go
                    (Set.union completed selectedRefs)
                    nextRemaining
                    (selected : completedBatches)

greedilyCompatible :: [Node] -> [Node]
greedilyCompatible = foldl addIfCompatible []
  where
    addIfCompatible selected candidate
      | all (null . overlappingScopes candidate) selected = selected ++ [candidate]
      | otherwise = selected

overlappingScopes :: Node -> Node -> [FilePath]
overlappingScopes left right =
  Set.toAscList $
    Set.fromList
      [ commonScope leftScope rightScope
        | leftScope <- nodeScopes left,
          rightScope <- nodeScopes right,
          scopesOverlap leftScope rightScope
      ]

scopesOverlap :: FilePath -> FilePath -> Bool
scopesOverlap left right =
  let leftPrefix = scopePrefix left
      rightPrefix = scopePrefix right
   in leftPrefix `isPrefixOf` rightPrefix || rightPrefix `isPrefixOf` leftPrefix

scopePrefix :: FilePath -> FilePath
scopePrefix = takeWhile (/= '*')

isPrefixOf :: String -> String -> Bool
isPrefixOf prefix value = take (length prefix) value == prefix

commonScope :: FilePath -> FilePath -> FilePath
commonScope left right
  | length (scopePrefix left) <= length (scopePrefix right) = left
  | otherwise = right

parallelDecision :: [Node] -> [[Node]] -> (TaskRef, TaskRef) -> ParallelDecision
parallelDecision nodes scheduled (leftRef, rightRef) =
  case (findNode leftRef, findNode rightRef) of
    (Just left, Just right)
      | batchIndex leftRef == batchIndex rightRef ->
          ParallelAccepted (nodeName left) (nodeName right)
      | not (null overlap) ->
          ParallelSerialized (nodeName left) (nodeName right) overlap
      | otherwise -> ParallelDeferred (nodeName left) (nodeName right)
      where
        overlap = overlappingScopes left right
    _ -> ParallelDeferred (show leftRef) (show rightRef)
  where
    findNode reference = find ((== reference) . nodeRef) nodes
    batchIndex :: TaskRef -> Maybe Int
    batchIndex reference =
      findIndex 0 scheduled
      where
        findIndex :: Int -> [[Node]] -> Maybe Int
        findIndex index (batch : rest)
          | any ((== reference) . nodeRef) batch = Just index
          | otherwise = findIndex (index + 1) rest
        findIndex _ [] = Nothing

longestPath :: [Node] -> [Node]
longestPath [] = []
longestPath nodes = maximumBy (comparing length) (map pathTo nodes)
  where
    pathTo node =
      case mapMaybeNode (Set.toList (nodeDependencies node)) of
        [] -> [node]
        dependenciesValue ->
          maximumBy (comparing length) (map pathTo dependenciesValue) ++ [node]
    mapMaybeNode references =
      [ node
        | reference <- references,
          Just node <- [find ((== reference) . nodeRef) nodes]
      ]

nameFor :: [Node] -> TaskRef -> String
nameFor nodes reference =
  case find ((== reference) . nodeRef) nodes of
    Nothing -> "<missing>"
    Just node -> nodeName node

renderManifest :: DriverPlan -> String
renderManifest plan =
  unlines $
    ["project: " ++ name plan]
      ++ concatMap renderBatch (zip [1 :: Int ..] (batches plan))
      ++ ["critical-path: " ++ intercalate " -> " (criticalPath plan)]
      ++ map (("parallel: " ++) . show) (parallelDecisions plan)
  where
    renderBatch (index, batch) =
      ("batch " ++ show index ++ ":")
        : [ "  - "
              ++ taskName taskValue
              ++ " after ["
              ++ intercalate ", " (dependencies taskValue)
              ++ "] scopes "
              ++ show (scopes taskValue)
            | taskValue <- batch
          ]
