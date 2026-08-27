module Provenance
  ( Term (..),
    Pattern (..),
    Fact (..),
    Rule (..),
    Provenance,
    Knowledge,
    zero,
    one,
    source,
    plus,
    times,
    evaluate,
    query,
    alternatives,
    renderFact,
    renderProvenance,
  )
where

import Data.List (intercalate)
import qualified Data.Map.Strict as Map
import Data.Map.Strict (Map)
import qualified Data.Set as Set
import Data.Set (Set)

data Term
  = Variable String
  | Literal String
  deriving (Eq, Ord, Show)

data Pattern = Pattern String [Term]
  deriving (Eq, Ord, Show)

data Fact = Fact String [String]
  deriving (Eq, Ord, Show)

data Rule = Rule
  { ruleName :: String,
    premises :: [Pattern],
    conclusion :: Pattern
  }
  deriving (Eq, Show)

newtype Monomial = Monomial (Set String)
  deriving (Eq, Ord, Show)

newtype Provenance = Provenance (Set Monomial)
  deriving (Eq, Ord, Show)

type Knowledge = Map Fact Provenance

type Environment = Map String String

zero :: Provenance
zero = Provenance Set.empty

one :: Provenance
one = Provenance (Set.singleton (Monomial Set.empty))

source :: String -> Provenance
source identifier = token ("fact:" ++ identifier)

token :: String -> Provenance
token value = Provenance (Set.singleton (Monomial (Set.singleton value)))

plus :: Provenance -> Provenance -> Provenance
plus (Provenance left) (Provenance right) = Provenance (Set.union left right)

times :: Provenance -> Provenance -> Provenance
times (Provenance left) (Provenance right) =
  Provenance $
    Set.fromList
      [ Monomial (Set.union leftTokens rightTokens)
        | Monomial leftTokens <- Set.toList left,
          Monomial rightTokens <- Set.toList right
      ]

evaluate :: [Rule] -> [(Fact, String)] -> Either String Knowledge
evaluate rules bases = do
  mapM_ validateRule rules
  let initial =
        Map.fromListWith plus [(fact, source identifier) | (fact, identifier) <- bases]
  Right (fixedPoint (deriveOnce rules) initial)

query :: Fact -> Knowledge -> Maybe Provenance
query = Map.lookup

alternatives :: Provenance -> Int
alternatives (Provenance values) = Set.size values

fixedPoint :: (Eq a) => (a -> a) -> a -> a
fixedPoint step current =
  let next = step current
   in if next == current then current else fixedPoint step next

deriveOnce :: [Rule] -> Knowledge -> Knowledge
deriveOnce rules knowledge = foldl applyRule knowledge rules
  where
    applyRule accumulated rule =
      foldl addDerivation accumulated (matches knowledge rule)
    addDerivation accumulated (fact, provenance) =
      Map.insertWith plus fact provenance accumulated

matches :: Knowledge -> Rule -> [(Fact, Provenance)]
matches knowledge rule =
  [ (fact, token ("rule:" ++ ruleName rule) `times` premiseProvenance)
    | (environment, premiseProvenance) <-
        matchPremises knowledge Map.empty one (premises rule),
      Just fact <- [instantiate environment (conclusion rule)]
  ]

matchPremises :: Knowledge -> Environment -> Provenance -> [Pattern] -> [(Environment, Provenance)]
matchPremises _ environment provenance [] = [(environment, provenance)]
matchPremises knowledge environment provenance (patternValue : rest) = do
  (fact, factProvenance) <- Map.toList knowledge
  nextEnvironment <- maybeToList (matchPattern environment patternValue fact)
  matchPremises knowledge nextEnvironment (provenance `times` factProvenance) rest

matchPattern :: Environment -> Pattern -> Fact -> Maybe Environment
matchPattern environment (Pattern expectedPredicate terms) (Fact actualPredicate values)
  | expectedPredicate /= actualPredicate = Nothing
  | length terms /= length values = Nothing
  | otherwise = foldPairs environment (zip terms values)
  where
    foldPairs current [] = Just current
    foldPairs current ((term, value) : rest) =
      case matchTerm current term value of
        Nothing -> Nothing
        Just next -> foldPairs next rest

matchTerm :: Environment -> Term -> String -> Maybe Environment
matchTerm environment (Literal expected) actual
  | expected == actual = Just environment
  | otherwise = Nothing
matchTerm environment (Variable name) actual =
  case Map.lookup name environment of
    Nothing -> Just (Map.insert name actual environment)
    Just expected
      | expected == actual -> Just environment
      | otherwise -> Nothing

instantiate :: Environment -> Pattern -> Maybe Fact
instantiate environment (Pattern predicate terms) =
  Fact predicate <$> traverse resolve terms
  where
    resolve (Literal value) = Just value
    resolve (Variable name) = Map.lookup name environment

validateRule :: Rule -> Either String ()
validateRule rule =
  if conclusionVariables `Set.isSubsetOf` premiseVariables
    then Right ()
    else
      Left
        ( "unsafe rule "
            ++ ruleName rule
            ++ ": conclusion contains unbound variables "
            ++ show (Set.toList (Set.difference conclusionVariables premiseVariables))
        )
  where
    conclusionVariables = variablesInPattern (conclusion rule)
    premiseVariables = Set.unions (map variablesInPattern (premises rule))

variablesInPattern :: Pattern -> Set String
variablesInPattern (Pattern _ terms) =
  Set.fromList [name | Variable name <- terms]

renderFact :: Fact -> String
renderFact (Fact predicate values) = predicate ++ "(" ++ intercalate ", " values ++ ")"

renderProvenance :: Provenance -> String
renderProvenance (Provenance values)
  | Set.null values = "no derivation"
  | otherwise =
      intercalate
        "\nOR\n"
        [ "  " ++ intercalate " AND " (Set.toAscList tokens)
          | Monomial tokens <- Set.toAscList values
        ]

maybeToList :: Maybe a -> [a]
maybeToList Nothing = []
maybeToList (Just value) = [value]
