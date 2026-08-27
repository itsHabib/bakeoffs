module Main (main) where

import Artifact
  ( Artifact (..),
    BiView (..),
    Conflict (..),
    Intent (..),
    Model (..),
    Policy (..),
    renameProject,
    workflowView,
  )
import Control.Monad (unless)
import Example
  ( illegalDependencyEdit,
    initialIntent,
    legalEdit,
    overriddenPolicy,
  )

main :: IO ()
main = do
  mapM_ checkGetPut intentCorpus
  mapM_ checkPutGet legalCorpus
  checkPutPut
  checkRenamePreservesOverride
  checkDerivedEditRefused
  checkInvalidPolicyRefused
  putStrLn "bidirectional-artifacts: all laws and hard cases passed"

intentCorpus :: [Intent]
intentCorpus =
  [ initialIntent,
    initialIntent {policyOverride = Just overriddenPolicy},
    renameProject "ship-next" initialIntent
  ]

legalCorpus :: [(Intent, Artifact)]
legalCorpus =
  [ (intent, edit artifact)
    | intent <- intentCorpus,
      let artifact = get workflowView intent,
      edit <-
        [ id,
          (\value -> value {executionPolicy = Policy Local 1 0}),
          (\value -> value {executionPolicy = Policy Frontier 16 5})
        ]
  ]

checkGetPut :: Intent -> IO ()
checkGetPut intent =
  assert "GetPut" (put workflowView intent (get workflowView intent) == Right intent)

checkPutGet :: (Intent, Artifact) -> IO ()
checkPutGet (intent, edited) =
  case put workflowView intent edited of
    Left conflicts -> fail ("PutGet setup refused legal edit: " ++ show conflicts)
    Right reconciled -> assert "PutGet" (get workflowView reconciled == edited)

checkPutPut :: IO ()
checkPutPut = do
  let first = legalEdit (get workflowView initialIntent)
      second = first {executionPolicy = Policy Local 4 0}
      sequential = do
        afterFirst <- put workflowView initialIntent first
        put workflowView afterFirst second
      direct = put workflowView initialIntent second
  assert "PutPut" (sequential == direct)

checkRenamePreservesOverride :: IO ()
checkRenamePreservesOverride = do
  reconciled <- requireRight (put workflowView initialIntent (legalEdit (get workflowView initialIntent)))
  let regenerated = get workflowView (renameProject "ship-next" reconciled)
  assert "rename updates derived project" (artifactProject regenerated == "ship-next")
  assert "rename preserves legal override" (executionPolicy regenerated == overriddenPolicy)

checkDerivedEditRefused :: IO ()
checkDerivedEditRefused =
  case put workflowView initialIntent (illegalDependencyEdit (get workflowView initialIntent)) of
    Left [DerivedTasksChanged _ _] -> pure ()
    other -> fail ("derived edit did not produce precise refusal: " ++ show other)

checkInvalidPolicyRefused :: IO ()
checkInvalidPolicyRefused = do
  let generated = get workflowView initialIntent
      invalid = generated {executionPolicy = Policy Frontier 0 9}
  case put workflowView initialIntent invalid of
    Left conflicts -> do
      assert "invalid concurrency is reported" (InvalidConcurrency 0 `elem` conflicts)
      assert "invalid retries are reported" (InvalidRetries 9 `elem` conflicts)
    Right _ -> fail "invalid policy unexpectedly reconciled"

requireRight :: Either problem value -> IO value
requireRight result =
  case result of
    Left _ -> fail "expected successful reconciliation"
    Right value -> pure value

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
