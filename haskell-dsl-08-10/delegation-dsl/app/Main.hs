module Main (main) where

import Delegation (Compiled (..), compile, renderBrief, validate)
import Example (ambiguousContract, missingContract, validContract)

main :: IO ()
main = do
  putStrLn "COMPILED CONTRACT"
  case compile validContract of
    Left problems -> fail (show problems)
    Right compiled -> do
      print (executionBatches compiled)
      mapM_ (putStrLn . renderBrief) (roleBriefs compiled)
  putStrLn "MISSING PRODUCER"
  mapM_ print (validate missingContract)
  putStrLn "AMBIGUOUS PRODUCER"
  mapM_ print (validate ambiguousContract)
