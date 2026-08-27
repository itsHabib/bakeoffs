module Main (main) where

import Control.Monad (unless)
import Delegation
  ( Compiled (..),
    Problem (..),
    Role (..),
    compile,
    validate,
  )
import Example (ambiguousContract, findings, missingContract, validContract)

main :: IO ()
main = do
  case compile validContract of
    Left problems -> fail ("valid contract failed: " ++ show problems)
    Right compiled ->
      assert "dependency batches are deterministic" $
        executionBatches compiled
          == [ [Role "researcher"],
               [Role "implementer"],
               [Role "reviewer"]
             ]
  assert "missing producer is precise" $
    validate missingContract == [MissingProducer (Role "implementer") findings]
  assert "singleton artifact cannot have ambiguous producers" $
    validate ambiguousContract
      == [AmbiguousProducer findings [Role "researcher-a", Role "researcher-b"]]
  putStrLn "delegation-dsl: all tests passed"

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
