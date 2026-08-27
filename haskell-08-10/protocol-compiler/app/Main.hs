module Main (main) where

import Example (invalidProtocol, roles, validProtocol)
import Protocol (compile, renderContract, renderError)

main :: IO ()
main = do
  putStrLn "VALID PROTOCOL — projected local contracts"
  case compile roles validProtocol of
    Left compileError -> putStrLn ("unexpected refusal: " ++ renderError compileError)
    Right contracts -> mapM_ (putStrLn . uncurry renderContract) contracts
  putStrLn "ADVERSARIAL PROTOCOL — missing branch notification"
  case compile roles invalidProtocol of
    Left compileError -> putStrLn ("REFUSED\n" ++ renderError compileError)
    Right _ -> putStrLn "unexpectedly accepted"
