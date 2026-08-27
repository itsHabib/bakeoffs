module Main (main) where

import Capability (analyze, execute, renderPlan)
import Example (fixture, mergePlan, readOnlyGrant, reviewGrant, reviewPlan)

main :: IO ()
main = do
  putStrLn "PLAN"
  mapM_ putStrLn (renderPlan reviewPlan)
  putStrLn "ENVELOPE"
  print (analyze reviewPlan)
  putStrLn "READ-ONLY REFUSAL"
  print (execute readOnlyGrant fixture reviewPlan)
  putStrLn "EXACT EXECUTION"
  print (execute reviewGrant fixture reviewPlan)
  putStrLn "MERGE NEEDS MORE"
  print (execute reviewGrant fixture mergePlan)
