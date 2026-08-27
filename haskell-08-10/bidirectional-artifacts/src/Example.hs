module Example
  ( initialIntent,
    overriddenPolicy,
    legalEdit,
    illegalDependencyEdit,
  )
where

import Artifact
  ( Artifact (..),
    Intent (..),
    Model (..),
    Policy (..),
    Stream (..),
    Task (..),
  )

initialIntent :: Intent
initialIntent =
  Intent
    { project = "ship",
      streams =
        [ Stream "implement" [],
          Stream "verify" ["implement"],
          Stream "record" ["verify"]
        ],
      defaultPolicy = Policy Balanced 2 1,
      policyOverride = Nothing
    }

overriddenPolicy :: Policy
overriddenPolicy = Policy Frontier 3 2

legalEdit :: Artifact -> Artifact
legalEdit artifact = artifact {executionPolicy = overriddenPolicy}

illegalDependencyEdit :: Artifact -> Artifact
illegalDependencyEdit artifact =
  artifact
    { tasks = map editRecord (tasks artifact)
    }
  where
    editRecord task
      | taskId task == "ship/record" = task {dependsOn = ["ship/implement"]}
      | otherwise = task
