module Main (main) where

import Example (overlapProject, reviewKernel, ungatedProject)
import WorkDriver (compile, renderManifest)

main :: IO ()
main = do
  putStrLn "VALID PROJECT"
  render reviewKernel
  putStrLn "OVERLAPPING NOMINAL PARALLELISM"
  render overlapProject
  putStrLn "UNGATED LANDING"
  print (compile ungatedProject)
  where
    render projectValue =
      case compile projectValue of
        Left problems -> fail (show problems)
        Right plan -> putStrLn (renderManifest plan)
