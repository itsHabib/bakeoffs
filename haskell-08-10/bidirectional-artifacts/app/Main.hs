module Main (main) where

import Artifact
  ( BiView (..),
    Intent,
    renameProject,
    workflowView,
  )
import Example
  ( illegalDependencyEdit,
    initialIntent,
    legalEdit,
  )

main :: IO ()
main = do
  let generated = get workflowView initialIntent
      edited = legalEdit generated
  putStrLn "GENERATED"
  print generated
  reconciled <- requireRight (put workflowView initialIntent edited)
  let renamed = renameProject "ship-next" reconciled
      regenerated = get workflowView renamed
  putStrLn "\nLEGAL OVERRIDE, THEN UPSTREAM RENAME"
  print regenerated
  putStrLn "\nATTEMPT TO EDIT DERIVED DEPENDENCY"
  case put workflowView initialIntent (illegalDependencyEdit generated) of
    Left conflicts -> mapM_ print conflicts
    Right _ -> fail "derived dependency edit unexpectedly succeeded"
  putStrLn "\nRound-trip laws: see stack test (all pass)."

requireRight :: Either problem Intent -> IO Intent
requireRight result =
  case result of
    Left _ -> fail "legal artifact edit was refused"
    Right value -> pure value
