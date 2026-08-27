module Example
  ( rules,
    baseFacts,
    afterHeadMove,
    readyOldHead,
    readyNewHead,
    impactContract,
  )
where

import Provenance (Fact (..), Pattern (..), Rule (..), Term (..))

rules :: [Rule]
rules =
  [ Rule
      "ready-via-gate"
      [pattern2 "gate_pass" "pr" "head", pattern2 "checks_pass" "pr" "head"]
      (pattern2 "ready" "pr" "head"),
    Rule
      "ready-via-human"
      [pattern2 "human_override" "pr" "head", pattern2 "checks_pass" "pr" "head"]
      (pattern2 "ready" "pr" "head"),
    Rule
      "dependency-transitive"
      [pattern2 "depends" "consumer" "middle", pattern2 "depends" "middle" "provider"]
      (pattern2 "depends" "consumer" "provider"),
    Rule
      "dependency-impact"
      [pattern2 "depends" "consumer" "provider"]
      (Pattern "impact" [Variable "provider", Variable "consumer"])
  ]

baseFacts :: [(Fact, String)]
baseFacts =
  [ (Fact "gate_pass" ["pr-42", "sha-a"], "gate-run-17"),
    (Fact "checks_pass" ["pr-42", "sha-a"], "checks-91"),
    (Fact "human_override" ["pr-42", "sha-a"], "judgment-4"),
    (Fact "depends" ["ship", "workbench/contracts"], "ship-import"),
    (Fact "depends" ["agent-console", "ship"], "console-import")
  ]

afterHeadMove :: [(Fact, String)]
afterHeadMove =
  [ (Fact "gate_pass" ["pr-42", "sha-b"], "gate-run-18"),
    (Fact "depends" ["ship", "workbench/contracts"], "ship-import"),
    (Fact "depends" ["agent-console", "ship"], "console-import")
  ]

readyOldHead :: Fact
readyOldHead = Fact "ready" ["pr-42", "sha-a"]

readyNewHead :: Fact
readyNewHead = Fact "ready" ["pr-42", "sha-b"]

impactContract :: Fact
impactContract = Fact "impact" ["workbench/contracts", "agent-console"]

pattern2 :: String -> String -> String -> Pattern
pattern2 predicate left right =
  Pattern predicate [Variable left, Variable right]
