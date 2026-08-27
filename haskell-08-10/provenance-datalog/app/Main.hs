module Main (main) where

import Example
  ( afterHeadMove,
    baseFacts,
    impactContract,
    readyNewHead,
    readyOldHead,
    rules,
  )
import Provenance
  ( evaluate,
    query,
    renderFact,
    renderProvenance,
  )

main :: IO ()
main = do
  putStrLn "INITIAL KNOWLEDGE"
  initial <- requireKnowledge baseFacts
  printQuery initial readyOldHead
  printQuery initial impactContract
  putStrLn "\nHEAD MOVES sha-a -> sha-b; exact-head evidence is replaced"
  moved <- requireKnowledge afterHeadMove
  printQuery moved readyOldHead
  printQuery moved readyNewHead
  printQuery moved impactContract
  where
    requireKnowledge bases =
      case evaluate rules bases of
        Left problem -> fail problem
        Right knowledge -> pure knowledge
    printQuery knowledge fact = do
      putStrLn (renderFact fact)
      case query fact knowledge of
        Nothing -> putStrLn "  NOT DERIVABLE"
        Just provenance -> putStrLn (renderProvenance provenance)
