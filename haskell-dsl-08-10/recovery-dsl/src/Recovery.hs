{-# LANGUAGE GADTs #-}

module Recovery
  ( Task (..),
    RunId (..),
    Verified (..),
    Receipt (..),
    Activity (..),
    Retry (..),
    Step,
    activity,
    deduplicated,
    idempotent,
    atLeastOnce,
    Workflow (..),
    (>>>),
    Program (..),
    Journal (..),
    Entry (..),
    Fixture (..),
    Event (..),
    Fault (..),
    Outcome (..),
    Error (..),
    run,
  )
where

import Data.List (find)

newtype Task = Task String deriving (Eq, Show)
newtype RunId = RunId String deriving (Eq, Show)
newtype Verified = Verified RunId deriving (Eq, Show)
newtype Receipt = Receipt String deriving (Eq, Show)

data Activity a b where
  Dispatch :: Activity Task RunId
  Verify :: Activity RunId Verified
  Record :: Activity Verified Receipt

data Retry = Deduplicated | Idempotent | AtLeastOnce
  deriving (Eq, Show)

data Step a b = Step
  { stepName :: String,
    stepActivity :: Activity a b,
    stepRetry :: Retry
  }

activity :: String -> Activity a b -> Retry -> Step a b
activity = Step

deduplicated :: Retry
deduplicated = Deduplicated

idempotent :: Retry
idempotent = Idempotent

atLeastOnce :: Retry
atLeastOnce = AtLeastOnce

data Workflow a b where
  Done :: Workflow a a
  Then :: Step a b -> Workflow b c -> Workflow a c

infixr 1 >>>

(>>>) :: Step a b -> Workflow b c -> Workflow a c
(>>>) = Then

data Program a b = Program
  { version :: String,
    workflow :: Workflow a b
  }

data Entry = Entry
  { entryVersion :: String,
    entryStep :: String,
    entryKind :: String,
    entryInput :: String,
    entryOutput :: String
  }
  deriving (Eq, Show)

newtype Journal = Journal {journalEntries :: [Entry]}
  deriving (Eq, Show)

data Event
  = Executed String
  | ReusedJournal String
  | ReusedExternal String
  | RetriedIdempotent String
  | RetriedAtLeastOnce String
  deriving (Eq, Show)

data Fixture = Fixture
  { nextRun :: Int,
    dispatches :: Int,
    verifications :: Int,
    records :: Int,
    externalDedupe :: [(String, String)],
    uncertainEffects :: [String],
    events :: [Event]
  }
  deriving (Eq, Show)

data Fault
  = NoFault
  | CrashAfterEffect String
  | CrashAfterCommit String
  deriving (Eq, Show)

data Error
  = DuplicateStep String
  | VersionDrift String String
  | ShapeDrift Int String String
  | InputDrift String String String
  | DecodeFailure String String
  deriving (Eq, Show)

data Outcome b
  = Complete b Journal Fixture
  | Crashed String Journal Fixture
  | Refused Error Fixture

run :: Fault -> Fixture -> Journal -> a -> Program a b -> Outcome b
run fault fixture journal input program =
  case validate program journal of
    Left problem -> Refused problem fixture
    Right () ->
      runSteps
        fault
        (version program)
        fixture
        (journalEntries journal)
        (journalEntries journal)
        input
        (workflow program)

validate :: Program a b -> Journal -> Either Error ()
validate program (Journal entries) = do
  rejectDuplicate (map fst expected)
  rejectDuplicate (map entryStep entries)
  mapM_ checkVersion entries
  mapM_ checkShape (zip3 [0 ..] expected entries)
  where
    expected = signature (workflow program)
    checkVersion entry
      | entryVersion entry == version program = Right ()
      | otherwise = Left (VersionDrift (version program) (entryVersion entry))
    checkShape (index, (name, expectedKind), entry)
      | name == entryStep entry && expectedKind == entryKind entry = Right ()
      | otherwise = Left (ShapeDrift index name (entryStep entry))

rejectDuplicate :: [String] -> Either Error ()
rejectDuplicate values =
  case find (\value -> length (filter (== value) values) > 1) values of
    Nothing -> Right ()
    Just value -> Left (DuplicateStep value)

signature :: Workflow a b -> [(String, String)]
signature Done = []
signature (Then step rest) = (stepName step, kind (stepActivity step)) : signature rest

runSteps :: Fault -> String -> Fixture -> [Entry] -> [Entry] -> a -> Workflow a b -> Outcome b
runSteps _ _ fixture [] committed value Done = Complete value (Journal committed) fixture
runSteps _ _ fixture (_ : _) _ _ Done = Refused (ShapeDrift 0 "end" "extra journal entry") fixture
runSteps fault programVersion fixture remaining committed input (Then step rest) =
  case remaining of
    entry : later ->
      case replay step input entry of
        Left problem -> Refused problem fixture
        Right output ->
          runSteps
            fault
            programVersion
            (logEvent (ReusedJournal (stepName step)) fixture)
            later
            committed
            output
            rest
    [] ->
      let key = effectKey step input
          prepared = markRetryEvent key step fixture
       in case perform prepared step input of
            (output, afterEffect) ->
              let entry = encodeEntry programVersion step input output
                  afterCommit = clearUncertain key afterEffect
                  nextJournal = committed ++ [entry]
               in case fault of
                    CrashAfterEffect target
                      | target == stepName step ->
                          Crashed target (Journal committed) (markUncertain key afterEffect)
                    CrashAfterCommit target
                      | target == stepName step ->
                          Crashed target (Journal nextJournal) afterCommit
                    _ ->
                      runSteps fault programVersion afterCommit [] nextJournal output rest

markRetryEvent :: String -> Step a b -> Fixture -> Fixture
markRetryEvent key step fixture
  | key `notElem` uncertainEffects fixture = fixture
  | otherwise =
      case stepRetry step of
        Deduplicated -> fixture
        Idempotent -> logEvent (RetriedIdempotent (stepName step)) fixture
        AtLeastOnce -> logEvent (RetriedAtLeastOnce (stepName step)) fixture

perform :: Fixture -> Step a b -> a -> (b, Fixture)
perform fixture step input =
  case stepRetry step of
    Deduplicated ->
      case lookup key (externalDedupe fixture) >>= decode (stepActivity step) of
        Just output ->
          (output, logEvent (ReusedExternal (stepName step)) fixture)
        Nothing ->
          let (output, changed) = executeActivity fixture (stepActivity step) input
              saved = (key, encodeOutput (stepActivity step) output) : externalDedupe changed
           in (output, (logEvent (Executed (stepName step)) changed) {externalDedupe = saved})
    _ ->
      let (output, changed) = executeActivity fixture (stepActivity step) input
       in (output, logEvent (Executed (stepName step)) changed)
  where
    key = effectKey step input

executeActivity :: Fixture -> Activity a b -> a -> (b, Fixture)
executeActivity fixture Dispatch _ =
  ( RunId ("run-" ++ show (nextRun fixture)),
    fixture {nextRun = nextRun fixture + 1, dispatches = dispatches fixture + 1}
  )
executeActivity fixture Verify runId =
  (Verified runId, fixture {verifications = verifications fixture + 1})
executeActivity fixture Record (Verified (RunId runId)) =
  (Receipt ("receipt-" ++ runId), fixture {records = records fixture + 1})

replay :: Step a b -> a -> Entry -> Either Error b
replay step input entry
  | entryInput entry /= encodeInput (stepActivity step) input =
      Left (InputDrift (stepName step) (encodeInput (stepActivity step) input) (entryInput entry))
  | otherwise =
      case decode (stepActivity step) (entryOutput entry) of
        Nothing -> Left (DecodeFailure (stepName step) (entryOutput entry))
        Just output -> Right output

encodeEntry :: String -> Step a b -> a -> b -> Entry
encodeEntry programVersion step input output =
  Entry
    programVersion
    (stepName step)
    (kind (stepActivity step))
    (encodeInput (stepActivity step) input)
    (encodeOutput (stepActivity step) output)

kind :: Activity a b -> String
kind Dispatch = "dispatch"
kind Verify = "verify"
kind Record = "record"

encodeInput :: Activity a b -> a -> String
encodeInput Dispatch (Task name) = "task:" ++ name
encodeInput Verify (RunId runId) = "run:" ++ runId
encodeInput Record (Verified (RunId runId)) = "verified:" ++ runId

encodeOutput :: Activity a b -> b -> String
encodeOutput Dispatch (RunId runId) = "run:" ++ runId
encodeOutput Verify (Verified (RunId runId)) = "verified:" ++ runId
encodeOutput Record (Receipt receipt) = "receipt:" ++ receipt

decode :: Activity a b -> String -> Maybe b
decode Dispatch value = RunId <$> strip "run:" value
decode Verify value = Verified . RunId <$> strip "verified:" value
decode Record value = Receipt <$> strip "receipt:" value

strip :: String -> String -> Maybe String
strip prefix value
  | take (length prefix) value == prefix = Just (drop (length prefix) value)
  | otherwise = Nothing

effectKey :: Step a b -> a -> String
effectKey step input = stepName step ++ ":" ++ encodeInput (stepActivity step) input

markUncertain :: String -> Fixture -> Fixture
markUncertain key fixture = fixture {uncertainEffects = key : uncertainEffects fixture}

clearUncertain :: String -> Fixture -> Fixture
clearUncertain key fixture = fixture {uncertainEffects = filter (/= key) (uncertainEffects fixture)}

logEvent :: Event -> Fixture -> Fixture
logEvent event fixture = fixture {events = events fixture ++ [event]}
