module Main (main) where

import Control.Monad (unless)
import Example
  ( afterHeadMove,
    baseFacts,
    impactContract,
    readyNewHead,
    readyOldHead,
    rules,
  )
import Provenance
  ( Fact (..),
    Pattern (..),
    Rule (..),
    Term (..),
    alternatives,
    evaluate,
    one,
    plus,
    query,
    source,
    times,
    zero,
  )

main :: IO ()
main = do
  let a = source "a"
      b = source "b"
      c = source "c"
  assert "plus is associative" ((a `plus` b) `plus` c == a `plus` (b `plus` c))
  assert "times is associative" ((a `times` b) `times` c == a `times` (b `times` c))
  assert "times distributes over plus" $
    a `times` (b `plus` c) == (a `times` b) `plus` (a `times` c)
  assert "zero is additive identity" (a `plus` zero == a)
  assert "one is multiplicative identity" (a `times` one == a)

  initial <- requireKnowledge baseFacts
  assert "readiness retains two alternative derivations" $
    maybe False ((== 2) . alternatives) (query readyOldHead initial)
  assert "recursive dependency reaches transitive impact" $
    query impactContract initial /= Nothing

  moved <- requireKnowledge afterHeadMove
  assert "stale readiness retracts after exact-head evidence moves" $
    query readyOldHead moved == Nothing
  assert "new head lacks checks and is not ready" $
    query readyNewHead moved == Nothing
  assert "unrelated dependency provenance survives recomputation" $
    query impactContract moved /= Nothing

  let unsafe =
        Rule
          "unsafe"
          [Pattern "input" [Variable "x"]]
          (Pattern "output" [Variable "missing"])
  assert "unsafe rules are refused" $
    case evaluate [unsafe] [(Fact "input" ["value"], "source")] of
      Left _ -> True
      Right _ -> False
  putStrLn "provenance-datalog: all tests passed"
  where
    requireKnowledge bases =
      case evaluate rules bases of
        Left problem -> fail problem
        Right knowledge -> pure knowledge

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
