package drill

import (
	"errors"
	"math"
	"strings"
	"unicode"
)

const PassThreshold = 0.82

type TrapRule struct {
	Name           string   `json:"name"`
	Detail         string   `json:"detail"`
	WrongFragments []string `json:"wrong_fragments,omitempty"`
	RequiredWords  []string `json:"required_words,omitempty"`
}

type Phrase struct {
	ID       string     `json:"id"`
	Target   string     `json:"target"`
	English  string     `json:"english"`
	Variants []string   `json:"variants"`
	Traps    []TrapRule `json:"traps"`
}

type WordDiff struct {
	Expected string `json:"expected,omitempty"`
	Heard    string `json:"heard,omitempty"`
	Status   string `json:"status"`
}

type Result struct {
	PhraseID     string     `json:"phrase_id"`
	Target       string     `json:"target"`
	Attempt      string     `json:"attempt"`
	Matched      string     `json:"matched_variant"`
	Score        float64    `json:"score"`
	Passed       bool       `json:"passed"`
	Diff         []WordDiff `json:"diff"`
	DetectedTrap []TrapRule `json:"detected_traps"`
}

type Fixture struct {
	Label    string `json:"label"`
	PhraseID string `json:"phrase_id"`
	Attempt  string `json:"attempt"`
}

func wrong(name, detail string, fragments ...string) TrapRule {
	return TrapRule{Name: name, Detail: detail, WrongFragments: fragments}
}

func missing(name, detail string, words ...string) TrapRule {
	return TrapRule{Name: name, Detail: detail, RequiredWords: words}
}

func phrase(id, target, english string, variants []string, traps ...TrapRule) Phrase {
	return Phrase{ID: id, Target: target, English: english, Variants: variants, Traps: traps}
}

var Pack = []Phrase{
	phrase("greeting", "buenas tardes", "good afternoon", []string{"buenas"},
		wrong("time-of-day mix-up", "Use tardes after midday, not días.", "buenos dias"), missing("plural agreement", "The greeting is buenas, not buena.", "buenas")),
	phrase("table-two", "una mesa para dos, por favor", "a table for two, please", []string{"mesa para dos, por favor"},
		wrong("article gender", "Mesa is feminine: una mesa.", "un mesa"), missing("politeness", "Keep por favor in this request.", "por favor")),
	phrase("reservation", "tengo una reserva a nombre de Ana", "I have a reservation under Ana", []string{"tenemos una reserva a nombre de Ana"},
		wrong("article gender", "Reserva takes una.", "un reserva"), missing("booking name", "Say a nombre de before the name.", "nombre")),
	phrase("outside", "¿podemos sentarnos afuera?", "can we sit outside?", []string{"¿nos podemos sentar afuera?"},
		wrong("person mismatch", "Use podemos for we, not puedo.", "puedo sentarnos"), missing("location", "Include afuera so the request is specific.", "afuera")),
	phrase("menu", "¿nos trae el menú, por favor?", "can you bring us the menu, please?", []string{"¿puede traernos el menú, por favor?"},
		wrong("article gender", "Menú takes el.", "la menu"), missing("politeness", "Keep por favor in this request.", "por favor")),
	phrase("recommend", "¿qué nos recomienda?", "what do you recommend?", []string{"¿qué recomienda?"},
		wrong("verb person", "Address the waiter with recomienda, not recomiendo.", "que recomiendo"), missing("question word", "The question needs qué.", "que")),
	phrase("special", "¿cuál es el plato del día?", "what is today's special?", []string{"¿qué plato tienen hoy?"},
		wrong("article gender", "Plato takes el.", "la plato"), missing("daily special", "Include día or hoy.", "dia")),
	phrase("vegetarian", "¿tienen opciones vegetarianas?", "do you have vegetarian options?", []string{"¿hay opciones vegetarianas?"},
		wrong("agreement", "Opciones is feminine plural: vegetarianas.", "opciones vegetarianos"), missing("diet requirement", "Include vegetarianas.", "vegetarianas")),
	phrase("allergy", "soy alérgico a los cacahuetes", "I am allergic to peanuts", []string{"soy alérgica a los cacahuetes", "tengo alergia a los cacahuetes"},
		wrong("false construction", "Use alérgico a, not alérgico de.", "alergico de"), missing("allergen", "Name cacahuetes explicitly.", "cacahuetes")),
	phrase("gluten", "¿este plato lleva gluten?", "does this dish contain gluten?", []string{"¿esto lleva gluten?"},
		wrong("English carry-over", "Spanish uses llevar for containing an ingredient.", "contiene el gluten"), missing("ingredient", "Include gluten.", "gluten")),
	phrase("water", "una botella de agua, por favor", "a bottle of water, please", []string{"agua, por favor"},
		wrong("article gender", "Botella takes una.", "un botella"), missing("politeness", "Keep por favor in this request.", "por favor")),
	phrase("sparkling", "agua con gas, por favor", "sparkling water, please", []string{"una botella de agua con gas, por favor"},
		wrong("wrong water", "Sin gas means still water.", "agua sin gas"), missing("sparkling", "Include con gas.", "gas")),
	phrase("still", "agua sin gas, por favor", "still water, please", []string{"una botella de agua sin gas, por favor"},
		wrong("wrong water", "Con gas means sparkling water.", "agua con gas"), missing("still", "Include sin gas.", "sin")),
	phrase("coffee", "un café solo, por favor", "an espresso, please", []string{"un café, por favor"},
		wrong("article gender", "Café takes un.", "una cafe"), missing("politeness", "Keep por favor in this request.", "por favor")),
	phrase("wine", "una copa de vino tinto, por favor", "a glass of red wine, please", []string{"vino tinto, por favor"},
		wrong("article gender", "Copa takes una.", "un copa"), missing("wine type", "Include tinto for red wine.", "tinto")),
	phrase("ready", "estamos listos para pedir", "we are ready to order", []string{"ya podemos pedir"},
		wrong("person agreement", "For we, use estamos.", "estoy listos"), missing("purpose", "Include pedir.", "pedir")),
	phrase("starter", "para empezar, queremos las croquetas", "to start, we want the croquettes", []string{"de primero, las croquetas"},
		wrong("article agreement", "Croquetas takes las.", "los croquetas"), missing("starter", "Name croquetas.", "croquetas")),
	phrase("order-paella", "para mí, la paella", "the paella for me", []string{"quiero la paella", "me pone la paella"},
		wrong("article gender", "Paella takes la.", "el paella"), missing("dish", "Name paella.", "paella")),
	phrase("order-chicken", "quiero el pollo a la plancha", "I want the grilled chicken", []string{"para mí, el pollo a la plancha"},
		wrong("article gender", "Pollo takes el.", "la pollo"), missing("preparation", "Include a la plancha.", "plancha")),
	phrase("order-fish", "¿qué pescado tienen hoy?", "what fish do you have today?", []string{"¿cuál es el pescado del día?"},
		wrong("article gender", "Pescado takes el.", "la pescado"), missing("today", "Include hoy or del día.", "hoy")),
	phrase("without-onion", "sin cebolla, por favor", "without onion, please", []string{"lo quiero sin cebolla"},
		wrong("opposite request", "Con cebolla asks for onion.", "con cebolla"), missing("ingredient", "Include cebolla.", "cebolla")),
	phrase("sauce-side", "la salsa aparte, por favor", "the sauce on the side, please", []string{"¿puede poner la salsa aparte?"},
		wrong("article gender", "Salsa takes la.", "el salsa"), missing("separation", "Include aparte.", "aparte")),
	phrase("well-done", "la carne bien hecha, por favor", "the meat well done, please", []string{"quiero la carne bien hecha"},
		wrong("agreement", "Carne is feminine: bien hecha.", "carne bien hecho"), missing("doneness", "Include bien hecha.", "hecha")),
	phrase("medium", "la carne al punto, por favor", "the meat medium, please", []string{"al punto, por favor"},
		wrong("wrong doneness", "Poco hecha means rare, not medium.", "poco hecha"), missing("doneness", "Include al punto.", "punto")),
	phrase("rare", "la carne poco hecha, por favor", "the meat rare, please", []string{"poco hecha, por favor"},
		wrong("wrong doneness", "Bien hecha means well done.", "bien hecha"), missing("doneness", "Include poco hecha.", "poco")),
	phrase("another", "¿nos trae otra ronda, por favor?", "can you bring another round, please?", []string{"otra ronda, por favor"},
		wrong("agreement", "Ronda is feminine: otra.", "otro ronda"), missing("quantity", "Include otra.", "otra")),
	phrase("missing", "todavía falta mi plato", "my dish is still missing", []string{"aún no ha llegado mi plato"},
		wrong("possessive agreement", "Plato takes mi, not mis.", "mis plato"), missing("missing item", "Include plato.", "plato")),
	phrase("wrong-order", "esto no es lo que pedí", "this isn't what I ordered", []string{"yo no pedí esto"},
		wrong("tense", "Use pedí for what you ordered.", "no es lo que pido"), missing("negation", "Include no.", "no")),
	phrase("cold", "la comida está fría", "the food is cold", []string{"este plato está frío"},
		wrong("agreement", "Comida is feminine: fría.", "comida esta frio"), missing("problem", "Include fría.", "fria")),
	phrase("hot", "¿puede calentarlo, por favor?", "can you heat it, please?", []string{"¿lo puede calentar, por favor?"},
		wrong("wrong request", "Enfriar means cool down.", "puede enfriarlo"), missing("action", "Include calentar.", "calentar")),
	phrase("delicious", "está delicioso", "it is delicious", []string{"está muy rico"},
		wrong("agreement", "For plato, use delicioso.", "esta deliciosa"), missing("evaluation", "Include delicioso or rico.", "delicioso")),
	phrase("dessert-menu", "¿nos trae la carta de postres?", "can you bring the dessert menu?", []string{"¿podemos ver la carta de postres?"},
		wrong("article gender", "Carta takes la.", "el carta"), missing("dessert", "Include postres.", "postres")),
	phrase("dessert", "quiero el flan, por favor", "I want the flan, please", []string{"para mí, el flan"},
		wrong("article gender", "Flan takes el.", "la flan"), missing("dessert", "Name flan.", "flan")),
	phrase("check", "la cuenta, por favor", "the check, please", []string{"¿nos trae la cuenta, por favor?", "¿me trae la cuenta, por favor?"},
		wrong("gender article", "Cuenta is feminine: la cuenta.", "el cuenta"), wrong("English carry-over", "Use cuenta, not check.", "check"), missing("politeness", "Keep por favor in this request.", "por favor")),
	phrase("together", "todo junto, por favor", "all together, please", []string{"una sola cuenta, por favor"},
		wrong("opposite request", "Por separado asks for separate checks.", "por separado"), missing("together", "Include junto or una sola cuenta.", "junto")),
	phrase("separate", "por separado, por favor", "separately, please", []string{"cuentas separadas, por favor"},
		wrong("opposite request", "Todo junto asks for one check.", "todo junto"), missing("separate", "Include separado.", "separado")),
	phrase("card", "¿puedo pagar con tarjeta?", "can I pay by card?", []string{"¿aceptan tarjeta?"},
		wrong("preposition", "Use con tarjeta.", "pagar por tarjeta"), missing("payment method", "Include tarjeta.", "tarjeta")),
	phrase("cash", "voy a pagar en efectivo", "I am going to pay in cash", []string{"pago en efectivo"},
		wrong("preposition", "Use en efectivo.", "con efectivo"), missing("payment method", "Include efectivo.", "efectivo")),
	phrase("tip", "¿está incluida la propina?", "is the tip included?", []string{"¿la propina está incluida?"},
		wrong("agreement", "Propina is feminine: incluida.", "propina incluido"), missing("charge", "Include propina.", "propina")),
	phrase("goodbye", "muchas gracias, hasta luego", "thank you very much, see you later", []string{"gracias, hasta luego"},
		wrong("agreement", "Gracias pairs with muchas.", "muchos gracias"), missing("farewell", "Include hasta luego.", "luego")),
}

var CannedFixtures = []Fixture{
	{Label: "clean pass", PhraseID: "check", Attempt: "la cuenta por favor"},
	{Label: "gender trap", PhraseID: "check", Attempt: "el cuenta por favor"},
	{Label: "dropped politeness", PhraseID: "check", Attempt: "la cuenta"},
	{Label: "English carry-over", PhraseID: "check", Attempt: "el check por favor"},
	{Label: "acceptable variant", PhraseID: "table-two", Attempt: "mesa para dos por favor"},
	{Label: "agreement trap", PhraseID: "vegetarian", Attempt: "tienen opciones vegetarianos"},
	{Label: "accent-tolerant pass", PhraseID: "coffee", Attempt: "un cafe solo por favor"},
	{Label: "total miss", PhraseID: "sparkling", Attempt: "queremos pagar ahora"},
}

func FindPhrase(id string) (Phrase, bool) {
	for _, candidate := range Pack {
		if candidate.ID == id {
			return candidate, true
		}
	}
	return Phrase{}, false
}

func Score(phraseID, attempt string) (Result, error) {
	p, ok := FindPhrase(phraseID)
	if !ok {
		return Result{}, errors.New("unknown phrase")
	}
	if strings.TrimSpace(attempt) == "" {
		return Result{}, errors.New("attempt is empty")
	}

	candidates := append([]string{p.Target}, p.Variants...)
	best := candidates[0]
	bestDistance := math.MaxInt
	for _, candidate := range candidates {
		distance := wordDistance(words(candidate), words(attempt))
		if distance < bestDistance {
			bestDistance = distance
			best = candidate
		}
	}
	maxWords := max(len(words(best)), len(words(attempt)))
	score := 1.0
	if maxWords > 0 {
		score = 1 - float64(bestDistance)/float64(maxWords)
	}
	if score < 0 {
		score = 0
	}

	detected := detectTraps(p, attempt, score)
	return Result{
		PhraseID:     p.ID,
		Target:       p.Target,
		Attempt:      attempt,
		Matched:      best,
		Score:        math.Round(score*100) / 100,
		Passed:       score >= PassThreshold && len(detected) == 0,
		Diff:         align(words(best), words(attempt)),
		DetectedTrap: detected,
	}, nil
}

func detectTraps(p Phrase, attempt string, score float64) []TrapRule {
	normalizedAttempt := normalize(attempt)
	var detected []TrapRule
	for _, trap := range p.Traps {
		found := false
		for _, fragment := range trap.WrongFragments {
			if strings.Contains(normalizedAttempt, normalize(fragment)) {
				found = true
				break
			}
		}
		if !found && score >= 0.45 && len(trap.RequiredWords) > 0 {
			for _, required := range trap.RequiredWords {
				if !strings.Contains(normalizedAttempt, normalize(required)) {
					found = true
					break
				}
			}
		}
		if found {
			detected = append(detected, trap)
		}
	}
	return detected
}

func words(value string) []string {
	return strings.Fields(normalize(value))
}

func normalize(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u",
	)
	value = replacer.Replace(value)
	return strings.Join(strings.FieldsFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != 'ñ'
	}), " ")
}

func wordDistance(a, b []string) int {
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
		dp[i][0] = i
	}
	for j := range dp[0] {
		dp[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			dp[i][j] = min(dp[i-1][j]+1, dp[i][j-1]+1, dp[i-1][j-1]+cost)
		}
	}
	return dp[len(a)][len(b)]
}

func align(expected, heard []string) []WordDiff {
	dp := make([][]int, len(expected)+1)
	for i := range dp {
		dp[i] = make([]int, len(heard)+1)
		dp[i][0] = i
	}
	for j := range dp[0] {
		dp[0][j] = j
	}
	for i := 1; i <= len(expected); i++ {
		for j := 1; j <= len(heard); j++ {
			cost := 1
			if expected[i-1] == heard[j-1] {
				cost = 0
			}
			dp[i][j] = min(dp[i-1][j]+1, dp[i][j-1]+1, dp[i-1][j-1]+cost)
		}
	}

	i, j := len(expected), len(heard)
	reversed := make([]WordDiff, 0, max(i, j))
	for i > 0 || j > 0 {
		if i > 0 && j > 0 {
			cost := 1
			status := "changed"
			if expected[i-1] == heard[j-1] {
				cost, status = 0, "match"
			}
			if dp[i][j] == dp[i-1][j-1]+cost {
				reversed = append(reversed, WordDiff{Expected: expected[i-1], Heard: heard[j-1], Status: status})
				i--
				j--
				continue
			}
		}
		if i > 0 && dp[i][j] == dp[i-1][j]+1 {
			reversed = append(reversed, WordDiff{Expected: expected[i-1], Status: "missing"})
			i--
			continue
		}
		reversed = append(reversed, WordDiff{Heard: heard[j-1], Status: "extra"})
		j--
	}
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	return reversed
}
