{-# LANGUAGE GADTs #-}

module Workflow
  ( Task (..),
    RunId (..),
    Verified (..),
    Receipt (..),
    Kind (..),
    Activity (..),
    Step (..),
    Workflow (..),
    Program (..),
    Entry (..),
    Journal (..),
    Fixture (..),
    Fault (..),
    CrashPoint (..),
    JournalError (..),
    WorkflowError (..),
    Outcome (..),
    workflowSignature,
    validateJournal,
    runProgram,
  )
where

import Data.List (find)

newtype Task = Task String
  deriving (Eq, Show)

newtype RunId = RunId String
  deriving (Eq, Show)

newtype Verified = Verified RunId
  deriving (Eq, Show)

newtype Receipt = Receipt String
  deriving (Eq, Show)

data Kind
  = DispatchKind
  | VerifyKind
  | RecordKind
  deriving (Eq, Show)

data Activity a b where
  Dispatch :: Activity Task RunId
  Verify :: Activity RunId Verified
  Record :: Activity Verified Receipt

data Step a b = Step
  { stepId :: String,
    activity :: Activity a b
  }

data Workflow a b where
  Halt :: Workflow a a
  Then :: Step a b -> Workflow b c -> Workflow a c

data Program a b = Program
  { programVersion :: String,
    programWorkflow :: Workflow a b
  }

data Entry = Entry
  { entryVersion :: String,
    entryStepId :: String,
    entryKind :: Kind,
    encodedInput :: String,
    encodedResult :: String
  }
  deriving (Eq, Show)

newtype Journal = Journal {entries :: [Entry]}
  deriving (Eq, Show)

data Fixture = Fixture
  { nextRunNumber :: Int,
    dispatchCount :: Int,
    verifyCount :: Int,
    recordCount :: Int,
    verificationPasses :: Bool
  }
  deriving (Eq, Show)

data Fault
  = NoFault
  | CrashAfterEffectBeforeCommit String
  | CrashAfterCommit String
  deriving (Eq, Show)

data CrashPoint
  = AfterEffectBeforeCommit String
  | AfterCommit String
  deriving (Eq, Show)

data JournalError
  = DuplicateStepId String
  | DuplicateWorkflowStepId String
  | VersionMismatch String String
  | JournalLongerThanWorkflow Int Int
  | ShapeMismatch Int (String, Kind) (String, Kind)
  deriving (Eq, Show)

data WorkflowError
  = InvalidJournal JournalError
  | InputConflict String String String
  | ResultDecodeFailure String String
  | ActivityFailure String String
  deriving (Eq, Show)

data Outcome b
  = Completed b Journal Fixture
  | Crashed CrashPoint Journal Fixture
  | Refused WorkflowError Fixture

workflowSignature :: Program a b -> [(String, Kind)]
workflowSignature program = signature (programWorkflow program)

signature :: Workflow a b -> [(String, Kind)]
signature Halt = []
signature (Then step rest) = stepSignature step : signature rest

stepSignature :: Step a b -> (String, Kind)
stepSignature step = (stepId step, activityKind (activity step))

validateJournal :: Program a b -> Journal -> Either JournalError ()
validateJournal program (Journal recorded) = do
  rejectDuplicateWorkflowSteps expected
  rejectDuplicates recorded
  mapM_ checkVersion recorded
  if length recorded > length expected
    then Left (JournalLongerThanWorkflow (length recorded) (length expected))
    else mapM_ checkShape (zip3 [0 ..] expected recorded)
  where
    expected = workflowSignature program
    checkVersion entry
      | entryVersion entry == programVersion program = Right ()
      | otherwise =
          Left (VersionMismatch (programVersion program) (entryVersion entry))
    checkShape (index, expectedStep, entry)
      | expectedStep == (entryStepId entry, entryKind entry) = Right ()
      | otherwise =
          Left
            ( ShapeMismatch
                index
                expectedStep
                (entryStepId entry, entryKind entry)
            )

rejectDuplicateWorkflowSteps :: [(String, Kind)] -> Either JournalError ()
rejectDuplicateWorkflowSteps expected =
  case find duplicate identifiers of
    Nothing -> Right ()
    Just identifier -> Left (DuplicateWorkflowStepId identifier)
  where
    identifiers = map fst expected
    duplicate identifier = length (filter (== identifier) identifiers) > 1

rejectDuplicates :: [Entry] -> Either JournalError ()
rejectDuplicates recorded =
  case find duplicate identifiers of
    Nothing -> Right ()
    Just identifier -> Left (DuplicateStepId identifier)
  where
    identifiers = map entryStepId recorded
    duplicate identifier = length (filter (== identifier) identifiers) > 1

runProgram :: Fault -> Fixture -> Journal -> a -> Program a b -> Outcome b
runProgram fault fixture journal initial program =
  case validateJournal program journal of
    Left problem -> Refused (InvalidJournal problem) fixture
    Right () ->
      runSteps
        fault
        (programVersion program)
        fixture
        (entries journal)
        (entries journal)
        initial
        (programWorkflow program)

runSteps :: Fault -> String -> Fixture -> [Entry] -> [Entry] -> a -> Workflow a b -> Outcome b
runSteps _ _ fixture [] complete value Halt = Completed value (Journal complete) fixture
runSteps _ _ fixture (_ : _) _ _ Halt =
  Refused (InvalidJournal (JournalLongerThanWorkflow 1 0)) fixture
runSteps fault version fixture remaining complete input (Then step rest) =
  case remaining of
    entry : later ->
      case replay step input entry of
        Left problem -> Refused problem fixture
        Right output ->
          runSteps fault version fixture later complete output rest
    [] ->
      case perform fixture (activity step) input of
        Left message -> Refused (ActivityFailure (stepId step) message) fixture
        Right (output, afterEffect) ->
          let entry = encodeEntry version step input output
              committed = complete ++ [entry]
           in case fault of
                CrashAfterEffectBeforeCommit target
                  | target == stepId step ->
                      Crashed
                        (AfterEffectBeforeCommit target)
                        (Journal complete)
                        afterEffect
                CrashAfterCommit target
                  | target == stepId step ->
                      Crashed (AfterCommit target) (Journal committed) afterEffect
                _ ->
                  runSteps fault version afterEffect [] committed output rest

replay :: Step a b -> a -> Entry -> Either WorkflowError b
replay step input entry
  | encodedInput entry /= encodeInput (activity step) input =
      Left
        ( InputConflict
            (stepId step)
            (encodeInput (activity step) input)
            (encodedInput entry)
        )
  | otherwise =
      case decodeResult (activity step) (encodedResult entry) of
        Nothing ->
          Left (ResultDecodeFailure (stepId step) (encodedResult entry))
        Just output -> Right output

encodeEntry :: String -> Step a b -> a -> b -> Entry
encodeEntry version step input output =
  Entry
    { entryVersion = version,
      entryStepId = stepId step,
      entryKind = activityKind (activity step),
      encodedInput = encodeInput (activity step) input,
      encodedResult = encodeResult (activity step) output
    }

activityKind :: Activity a b -> Kind
activityKind Dispatch = DispatchKind
activityKind Verify = VerifyKind
activityKind Record = RecordKind

encodeInput :: Activity a b -> a -> String
encodeInput Dispatch (Task name) = "task:" ++ name
encodeInput Verify (RunId identifier) = "run:" ++ identifier
encodeInput Record (Verified (RunId identifier)) = "verified:" ++ identifier

encodeResult :: Activity a b -> b -> String
encodeResult Dispatch (RunId identifier) = "run:" ++ identifier
encodeResult Verify (Verified (RunId identifier)) = "verified:" ++ identifier
encodeResult Record (Receipt identifier) = "receipt:" ++ identifier

decodeResult :: Activity a b -> String -> Maybe b
decodeResult Dispatch value = RunId <$> stripPrefix "run:" value
decodeResult Verify value = Verified . RunId <$> stripPrefix "verified:" value
decodeResult Record value = Receipt <$> stripPrefix "receipt:" value

stripPrefix :: String -> String -> Maybe String
stripPrefix prefix value
  | take (length prefix) value == prefix = Just (drop (length prefix) value)
  | otherwise = Nothing

perform :: Fixture -> Activity a b -> a -> Either String (b, Fixture)
perform fixture Dispatch _ =
  let identifier = "run-" ++ show (nextRunNumber fixture)
   in Right
        ( RunId identifier,
          fixture
            { nextRunNumber = nextRunNumber fixture + 1,
              dispatchCount = dispatchCount fixture + 1
            }
        )
perform fixture Verify runIdentifier
  | verificationPasses fixture =
      Right
        ( Verified runIdentifier,
          fixture {verifyCount = verifyCount fixture + 1}
        )
  | otherwise = Left "verification failed"
perform fixture Record (Verified (RunId identifier)) =
  Right
    ( Receipt ("receipt-for-" ++ identifier),
      fixture {recordCount = recordCount fixture + 1}
    )
