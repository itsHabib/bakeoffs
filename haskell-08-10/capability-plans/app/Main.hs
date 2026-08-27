module Main (main) where

import Example
  ( exactReviewGrant,
    fixture,
    mergePlan,
    readOnlyGrant,
    reviewPlan,
  )
import Plan
  ( Execution (..),
    Fixture (..),
    analyze,
    executeAuthorized,
    operationOrder,
  )

main :: IO ()
main = do
  putStrLn "REVIEW PLAN PREFLIGHT"
  print (analyze reviewPlan)
  print (operationOrder reviewPlan)
  putStrLn "\nREAD-ONLY GRANT (REFUSED BEFORE EFFECTS)"
  print (executeAuthorized readOnlyGrant fixture reviewPlan)
  putStrLn ("fixture remains: " ++ show fixture)
  putStrLn "\nEXACT GRANT (EXECUTED)"
  case executeAuthorized exactReviewGrant fixture reviewPlan of
    Left problem -> fail (show problem)
    Right execution -> do
      print (result execution)
      print (fixtureLog (finalFixture execution))
      print (fixtureComments (finalFixture execution))
  putStrLn "\nMERGE PLAN REQUIRES A DISTINCT CAPABILITY"
  print (analyze mergePlan)
  print (executeAuthorized exactReviewGrant fixture mergePlan)
