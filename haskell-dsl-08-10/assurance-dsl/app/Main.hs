module Main (main) where

import Assurance (evaluate, renderTrace, requirements)
import Example (humanEvidence, mergeRule, staleEvidence, subject, validEvidence)

main :: IO ()
main = do
  putStrLn "AUTHORED REQUIREMENTS"
  print (requirements mergeRule)
  putStrLn "\nGATE PATH"
  putStr (renderTrace (evaluate subject validEvidence mergeRule))
  putStrLn "\nHUMAN ALTERNATIVE"
  putStr (renderTrace (evaluate subject humanEvidence mergeRule))
  putStrLn "\nSTALE PREVIOUS-HEAD EVIDENCE"
  putStr (renderTrace (evaluate subject staleEvidence mergeRule))
