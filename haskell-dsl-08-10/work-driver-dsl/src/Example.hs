module Example
  ( reviewKernel,
    overlapProject,
    ungatedProject,
  )
where

import WorkDriver
  ( Project,
    after,
    afterAll,
    land,
    parallel,
    project,
    task,
    touches,
    validate,
  )

reviewKernel :: Project
reviewKernel = project "review-kernel" $ do
  spec <- task "spec" `touches` ["docs/**"]
  (implementation, fixtures) <-
    parallel
      (task "implement" `after` spec `touches` ["src/**"])
      (task "fixtures" `after` spec `touches` ["test/**"])
  green <- validate "local-green" `afterAll` [implementation, fixtures]
  _ <- land "merge" `after` green
  pure ()

overlapProject :: Project
overlapProject = project "overlap" $ do
  spec <- task "spec"
  (left, right) <-
    parallel
      (task "edit-gate" `after` spec `touches` ["src/**"])
      (task "edit-gate-tests" `after` spec `touches` ["src/Gate.hs"])
  green <- validate "green" `afterAll` [left, right]
  _ <- land "merge" `after` green
  pure ()

ungatedProject :: Project
ungatedProject = project "unsafe" $ do
  implementation <- task "implement"
  _ <- land "merge" `after` implementation
  pure ()
