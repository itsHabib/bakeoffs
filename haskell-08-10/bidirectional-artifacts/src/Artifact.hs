module Artifact
  ( Model (..),
    Policy (..),
    Stream (..),
    Intent (..),
    Task (..),
    Artifact (..),
    Conflict (..),
    BiView (..),
    workflowView,
    effectivePolicy,
    renameProject,
  )
where

data Model
  = Frontier
  | Balanced
  | Local
  deriving (Eq, Show)

data Policy = Policy
  { model :: Model,
    concurrency :: Int,
    retries :: Int
  }
  deriving (Eq, Show)

data Stream = Stream
  { streamKey :: String,
    dependencyKeys :: [String]
  }
  deriving (Eq, Show)

data Intent = Intent
  { project :: String,
    streams :: [Stream],
    defaultPolicy :: Policy,
    policyOverride :: Maybe Policy
  }
  deriving (Eq, Show)

data Task = Task
  { taskId :: String,
    dependsOn :: [String]
  }
  deriving (Eq, Show)

data Artifact = Artifact
  { artifactProject :: String,
    tasks :: [Task],
    executionPolicy :: Policy
  }
  deriving (Eq, Show)

data Conflict
  = DerivedProjectChanged
      { expectedProject :: String,
        actualProject :: String
      }
  | DerivedTasksChanged
      { expectedTasks :: [Task],
        actualTasks :: [Task]
      }
  | InvalidConcurrency Int
  | InvalidRetries Int
  deriving (Eq, Show)

data BiView source view = BiView
  { get :: source -> view,
    put :: source -> view -> Either [Conflict] source
  }

workflowView :: BiView Intent Artifact
workflowView = BiView render reconcile

effectivePolicy :: Intent -> Policy
effectivePolicy intent =
  case policyOverride intent of
    Nothing -> defaultPolicy intent
    Just overridden -> overridden

renameProject :: String -> Intent -> Intent
renameProject name intent = intent {project = name}

render :: Intent -> Artifact
render intent =
  Artifact
    { artifactProject = project intent,
      tasks = map (renderTask intent) (streams intent),
      executionPolicy = effectivePolicy intent
    }

renderTask :: Intent -> Stream -> Task
renderTask intent stream =
  Task
    { taskId = identifier intent (streamKey stream),
      dependsOn = map (identifier intent) (dependencyKeys stream)
    }

identifier :: Intent -> String -> String
identifier intent key = project intent ++ "/" ++ key

reconcile :: Intent -> Artifact -> Either [Conflict] Intent
reconcile intent edited =
  case conflicts of
    [] -> Right (applyPolicy intent (executionPolicy edited))
    values -> Left values
  where
    generated = render intent
    conflicts =
      projectConflicts generated edited
        ++ taskConflicts generated edited
        ++ policyConflicts (executionPolicy edited)

projectConflicts :: Artifact -> Artifact -> [Conflict]
projectConflicts generated edited
  | artifactProject generated == artifactProject edited = []
  | otherwise =
      [ DerivedProjectChanged
          (artifactProject generated)
          (artifactProject edited)
      ]

taskConflicts :: Artifact -> Artifact -> [Conflict]
taskConflicts generated edited
  | tasks generated == tasks edited = []
  | otherwise = [DerivedTasksChanged (tasks generated) (tasks edited)]

policyConflicts :: Policy -> [Conflict]
policyConflicts policy =
  invalidConcurrency ++ invalidRetries
  where
    invalidConcurrency
      | concurrency policy >= 1 && concurrency policy <= 16 = []
      | otherwise = [InvalidConcurrency (concurrency policy)]
    invalidRetries
      | retries policy >= 0 && retries policy <= 5 = []
      | otherwise = [InvalidRetries (retries policy)]

applyPolicy :: Intent -> Policy -> Intent
applyPolicy intent chosen =
  intent
    { policyOverride =
        if chosen == defaultPolicy intent
          then Nothing
          else Just chosen
    }
