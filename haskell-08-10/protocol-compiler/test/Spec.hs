module Main (main) where

import Control.Monad (unless)
import Example (invalidProtocol, roles, validProtocol)
import Protocol
  ( CompileError (..),
    Local (..),
    Protocol (..),
    compile,
    project,
  )

main :: IO ()
main = do
  assert "valid protocol projects every role" $
    case compile roles validProtocol of
      Right contracts -> length contracts == length roles
      Left _ -> False
  assert "collector owns the outer choice" $
    case project "collector" validProtocol of
      Right (Receive "producer" "evidence.receipt" (Select _ "accepted" _ "malformed" _)) -> True
      _ -> False
  assert "unnotified role is rejected with its identity" $
    case compile roles invalidProtocol of
      Left (UnobservableChoice [] "gate" _ _) -> True
      _ -> False
  assert "self-messages are rejected" $
    project "agent" (Message "agent" "agent" "loop" End)
      == Left (SelfMessage [] "agent" "loop")
  assert "uninvolved identical branches need no notification" $
    project
      "observer"
      (Choice "chooser" [] "left" End "right" End)
      == Right Done
  putStrLn "protocol-compiler: all tests passed"

assert :: String -> Bool -> IO ()
assert label condition = unless condition (fail ("FAIL: " ++ label))
