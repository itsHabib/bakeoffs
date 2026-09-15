# Bounded-phraseology voice products: adversarial market report

Research current through 2026-08-07. This is a research report, not a product specification.

Confidence labels: **H** = directly supported by a current primary source or arithmetic from one; **M** = supported inference, vendor claim, or secondary source; **L** = directional estimate or anecdotal evidence that needs field validation.

## 1. Verdict on the ATC pick

### Verdict: no-go on the stated product; go on a narrower falsification demo

**NO-GO: “a subscription AI controller for student pilots” is no longer an open product thesis. [H]** [ATC One](https://atcone.app/) already sells the standalone, no-flight-simulator version almost verbatim: realtime AI radio conversations, VFR and IFR scenarios, readback drills, 1,000+ CFI-created exercises, instant feedback, school seats, and a $79.99 annual plan. It says it serves 10,000+ pilots and instructors and names five school/training partners; those are vendor claims, not audited adoption figures. Its site also says Global Language Training acquired it in August 2026, again without disclosed terms. [Readback](https://readback.app/) markets an even sharper direct hit—ground/tower/approach AI conversations, instant feedback, $0.025 per call or $49/year—although its update-list call to action leaves current general availability unclear. The broad idea has been built.

**GO: test a “readback debugger” that makes fewer false accusations than those products. [M]** The surviving wedge is one VFR tower-pattern flow whose differentiator is not natural conversation or an expansive curriculum. It is an inspectable semantic grader: show what clearance elements were issued, what the recognizer heard, which required safety elements matched, and when the audio is too uncertain to grade. The product must say *“I could not hear that reliably”* instead of *“wrong”* when ASR confidence is insufficient. That is a falsifiable claim and a small demo. It is not yet evidence of a durable subscription business.

### The strongest attack, and what survived

The strongest attack is three attacks reinforcing one another:

1. **The LLM-era entrant already shipped. [H/M]** ATC One is the exact standalone product and now has a school portal, multilingual content, a large exercise library, named partners, and an announced acquisition. Readback advertises a $49/year price point, though I did not independently verify that it is generally available. Mature PlaneEnglish ARSim already owns structured curriculum and institutional distribution. There is no unoccupied “old products were scripted; now voice AI makes this possible” gap.
2. **“Published grammar” does not mean one correct sentence. [H]** The FAA AIM says that understanding is the most important objective, that concise phraseology may be inadequate, and that pilots should use whatever words are necessary to get the message across ([AIM 4-2-1](https://www.faa.gov/air_traffic/publications/atpubs/aim_html/chap4_section_2.html)). A literal grammar checker would reproduce the exact complaint visible in incumbent reviews: valid, concise real-world readbacks are marked wrong. The deterministic object is a state-dependent set of required semantic slots and forbidden contradictions—not a canonical transcript.
3. **The sensor can destroy trust before the grader gets a chance. [H/M]** Current [ATC One App Store reviews](https://apps.apple.com/us/app/atc-one-ai-radio-practice/id6754663148) mention accents, numbers, and incorrect-call-sign failures; [ARSim reviews](https://apps.apple.com/us/app/planeenglish-arsim/id1461037185?see-all=reviews) describe missed words, robotic voices, and correct real-world variants being penalized. These are individual reviews, not a representative survey, but they converge on the core technical risk. A deterministic grader applied to a bad transcript is deterministically wrong.

**What survived:** radio practice is a real and named pain, people already pay, and the narrow tower-pattern workflow is accessible to a solo operator. AOPA calls the anxiety “mic fright” and recommends study and practice ([AOPA guide](https://www.aopa.org/training-and-safety/students/presolo/special/new-pilots-guide-to-atc-communication)). What did *not* survive is novelty, a generic subscription pitch, or the idea that an exact-string grammar is the moat.

### Competitor map

| Product | Current price | What it actually does | Complaint / structural gap | Confidence |
|---|---:|---|---|---|
| [ATC One](https://atcone.app/pricing) | Free tier; $79.99/year; school plans custom | Standalone realtime AI ATC, VFR/IFR, readbacks, phrase builder, ATIS, holding, grading, progress, school portal | Directly occupies the hypothesis. App Store complaints include recognition of accents/numbers and wrong callsigns; annual subscription friction. Its long prose “AI feedback” also suggests the model remains somewhere in the assessment presentation path. | **H** on product/price/reviews; **M** on adoption and grading inference |
| [Readback](https://readback.app/) | Advertised at $0.025/call or $49/year for 10,000 calls | Markets a standalone AI controller for ground, tower, approach with instant explanatory feedback | Second exact offer and price pressure. Its update-list CTA leaves launch/availability unclear; it says “AI explains,” not that a disclosed deterministic engine computes correctness. | **H** on advertised offer/price; **L** on availability; **M** on implementation |
| [PlaneEnglish ARSim](https://planeenglishsim.com/collections/homepage-grid-collection) | VFR+IFR starts at $8/month billed annually; enterprise custom | Mobile/web structured VFR-to-IFR curriculum, objective feedback, FAA WINGS credit, dashboards/seats | 4.1/5 from 323 US ratings when checked. Reviews complain about missed words, robotic audio, and rigid “correct” phrasing that differs from live ATC. | **H** |
| [PilotEdge](https://www.pilotedge.net/pages/pricing) | 5-hour trial; $19.95/month or $179/year one region; $34.90/month or $329/year combined | Live human controllers connected to a flight simulator | High realism but requires simulator setup, supported geography/hours, and substantially more money. It is practice, not deterministic diagnosis. | **H** on price/shape; **M** on user friction |
| [VATSIM](https://vatsim.net/docs/) | Free | Global volunteer human ATC network for flight simulators | Excellent exposure but simulator setup, variable volunteer coverage, social pressure, and no uniform deterministic grading. | **H** on free/network; **M** on friction |
| [SayIntentions.AI](https://www.sayintentions.ai/premium) | $18.95 monthly; about $190/year annually | Unscripted AI ATC and a broad flight-sim ecosystem at 88,000+ airports | Windows and a flight simulator required; optimized for immersion, not a tightly controlled teaching/evidence loop. | **H** |
| [BeyondATC](https://www.beyondatc.net/pricing) | $29.99 once; optional premium voice credits | Flight-sim ATC with unlimited basic voices and paid premium voices | Demonstrates how cheaply immersion can be sold when local/basic voices are acceptable. Not a standalone curriculum or transparent readback grader. | **H** |
| [Welly's ATC](https://github.com/rwellinger/xp_wellys_vfr_atc/blob/main/docs/README.md) | Open source, GPL-3.0 | X-Plane VFR voice ATC with local whisper.cpp/Piper options and a rule/state-machine-first intent engine | Proves that the architecture is buildable and partly local. It is a simulator plugin, not a zero-setup US student product; callsign recognition remains a documented limitation. | **H** |

The competitor conclusion is not “too crowded to enter.” It is narrower: **feature parity is worthless.** A new entrant needs measured evidence that it accepts valid variants and refuses critical errors more reliably than the existing products. [H/M]

### Market size, budget, distribution, and churn

- **Do not use 370,286 “active student pilots” as a live TAM. [H]** The FAA says the count reached 370,286 at the end of 2025 but suspended forecasting because student certificates issued after the 2016 rule change do not expire; the number is cumulative and no longer tracks active training ([FAA 2026 forecast](https://www.faa.gov/data_research/aviation/aerospace_forecasts/2026_FAA_Aerospace_Forecasts_FY2026-2046-2.pdf)).
- **The best available annual-entry proxy is 58,761 new student certificates issued in 2025. [M]** The number comes from the [FAA Civil Airmen Statistics workbook](https://www.faa.gov/data_research/aviation_data_statistics/civil_airmen_statistics); a readable extraction shows 61,353 in 2024 and 58,761 in 2025 ([Flight Brief digest](https://www.theflightbrief.com/articles/flight-school-enrollment-spring-2026-trends)). Issuance is not enrollment and happens at varying points in training.
- **The obvious US new-cohort revenue is small. [M]** At a $79.99 annual price, converting 10% of 58,761 annual certificate issuances produces about **$470,000** gross annual cohort revenue; 25% produces about **$1.18 million**. Those are scenarios, not forecasts, and exclude renewals and international demand. They describe a plausible bootstrapped niche, not venture-scale US B2C.
- **The school channel is real but fragmented. [H]** The FAA counted 509 certificated Part 141 schools in 2025, and says roughly 77% of private-pilot training occurs outside Part 141 ([FAA Part 141 modernization background](https://www.faa.gov/about/office_org/headquarters_offices/avs/offices/afx/afs/afs800/afs810/modernization_of_part-141_initiative/Part_141_Pilot_School_Modernization_Introductory.pdf)). There is no authoritative count of every independent Part 61 instructor/school, so a fabricated “total school TAM” would be misleading.
- **Students already pay comparable or larger amounts for training software. [H]** ForeFlight is $130/$260/$390 per year ([pricing](https://foreflight.com/pricing)); Sporty's Learn to Fly and King Schools private-pilot ground school are each $299 list ([Sporty's](https://www.sportys.com/learn-to-fly-course-private-pilot-test-prep-online-app-and-tv.html), [King](https://kingschools.com/private-pilot-ground-school-test-prep)). Willingness to pay is established; differentiation and retention are the problems.
- **Best distribution order: CFI assignment → flight-school seats → app stores/direct web; Reddit is discovery and support, not the primary engine. [M]** Both ATC One and PlaneEnglish expose institutional plans, and ATC One names school partners. App reviews also say instructors recommended or used these tools. A CFI saying “do these three drills before next lesson” is a stronger trigger than generic consumer acquisition.
- **Churn is structurally episodic. [M/L]** Radio anxiety is highest early, practice demand drops as a student solos and becomes comfortable, then returns when the pilot is rusty or begins instrument training. An ARSim review explicitly objects to paying continuously for a temporary need and wanting access again later. There is no public retention cohort data, so treat this as a hypothesis. It argues for a course, credits, school bundle, or lifetime unlock—not an assumption of evergreen monthly retention.

### Cost of one 30-minute session

[OpenAI's current pricing](https://developers.openai.com/api/docs/pricing) lists `gpt-realtime-2.1-mini` audio at $10/M input tokens, $0.30/M cached input, and $20/M output; full `gpt-realtime-2.1` is $32/$0.40/$64. [Realtime cost documentation](https://developers.openai.com/api/docs/guides/realtime-costs) says user audio is one token per 100 ms, assistant audio one token per 50 ms, and the conversation is resent on every response; later turns therefore cost more, though prior context is likely cacheable. [H]

Assume a 30-minute wall-clock drill has **6 minutes of student speech, 6 minutes of controller speech, 12 exchanges, and 18 minutes of listening/thinking/silence**. VAD excludes silence. [M]

| Path | Arithmetic | Estimated variable cost | Confidence |
|---|---|---:|---|
| Realtime mini, unique audio only | 3,600 input tokens × $10/M + 7,200 output × $20/M | **$0.18 lower bound** | **H** arithmetic; **M** usage assumption |
| Realtime mini, 12 turns with prior audio cached | $0.036 unique input + 59,400 cached input tokens × $0.30/M + $0.144 output | **$0.198 audio; budget $0.22–$0.30** with text/overhead | **M**; must verify from `response.done` usage |
| Realtime full, same cached assumption | $0.115 unique input + $0.024 cached history + $0.461 output | **$0.60 audio; budget $0.60–$0.80** | **M** |
| Realtime mini if caching misses | 63,000 total billed input tokens × $10/M + $0.144 output | **$0.77 before text** | **M**, adverse case |
| STT + deterministic engine + local/pre-rendered controller voice | 6 spoken input minutes × $0.003/min with `gpt-4o-mini-transcribe`; no model grader | **$0.018 STT plus near-zero marginal local/pre-rendered output** | **H** on STT; **M** on chosen voice path |

At competitor prices of roughly $4–$8/month on annual plans, full Realtime is a poor default for heavy users. Realtime mini can work with usage limits and good caching, but it is not a moat. The bounded-domain architecture should instead stream transcription, parse a finite set of slots, advance a deterministic scenario engine, and synthesize only the controller's known response. [M]

### Technical, content, asset, and regulatory feasibility

**Recognition:** local is credible enough to benchmark, not safe enough to assume. [H/M] [Vosk](https://alphacephei.com/vosk/) is offline, streaming, lightweight, and supports vocabulary reconfiguration; [whisper.cpp](https://github.com/ggml-org/whisper.cpp) supports Apple platforms and WebAssembly. Grammar post-matching improves interpretation but cannot retroactively recover a number the recognizer heard incorrectly. Preserve alternatives/confidence and make *uncertain* a first-class non-numeric outcome. Never let one transcript prove the user wrong.

**Latency:** browser WebRTC is the recommended client path for consistent Realtime performance ([OpenAI WebRTC guide](https://developers.openai.com/api/docs/guides/realtime-webrtc)). [H] There is no published end-to-end p95 that proves ATC realism on this exact workload. With push-to-talk, grading can begin while audio streams and controller output can start after release. A local/streaming ASR → slot parser → preselected response path should be competitive, but p50/p95 must be measured with accents, cabin noise, and ordinary laptops. [M]

**Content:** FAA-authored AIM, Pilot/Controller Glossary, and JO 7110.65 material is generally a US government work and not US-copyrightable under [17 USC §105](https://www.law.cornell.edu/uscode/text/17/105). [H] That does not automatically clear third-party charts, maps, recordings, logos, contractor-created material, or foreign rights. Use the text as authority, version every rule against its effective source, and create original scenarios/audio. The active controller order is [JO 7110.65BB](https://www.faa.gov/regulations_policies/orders_notices/index.cfm/go/document.current/documentnumber/7110.65). [H]

Encoding tower-pattern work is more content QA than code. A useful VFR slice is roughly 8–12 states—ATIS, ground call, taxi/hold short, tower/departure, pattern sequencing, landing/touch-and-go/go-around, exit, ground—with perhaps 50–100 controller templates once variants are included. Each issued instruction needs required slots, optional slots, accepted paraphrases, forbidden contradictions, and noisy/adversarial tests. Expect hundreds of positive and negative utterances before calling the grader trustworthy. [M]

**Open assets:** [Welly's ATC](https://github.com/rwellinger/xp_wellys_vfr_atc/blob/main/docs/README.md) supplies an existence proof and GPL-3.0 code, not a drop-in proprietary foundation. [ATCO2](https://atco2.org/data) has 5,281 hours of training audio under ELRA commercial/noncommercial terms, a four-hour licensed test set, and a free one-hour research-only subset; it is useful for evaluation but not “free production content.” VATSIM procedures can inform research but are not the FAA authority or a presumed reusable corpus. [H]

**Regulatory:** I found no FAA approval trigger for merely selling a supplemental communications-practice app that claims no training credit. [M] Approval matters when a device is represented or used for regulatory training credit: [AC 61-136B](https://www.faa.gov/regulations_policies/advisory_circulars/index.cfm/go/document.information/documentID/1034348) covers BATD/AATD approval and says a Part 141 school needs specific authorization to use an ATD in its approved course outline. [H] Therefore: make no “FAA approved,” loggable-hours, certification, or substitute-for-instruction claim; have current CFIs review scenarios; display the source/version; keep evidence of issued clearance, audio, transcript alternatives, and computed grade; and provide a correction channel. [M] This is product-risk analysis, not legal advice.

## 2. My candidates

These were generated from the bounded-language property, not from the brief's ICAO-English, medical, maritime, or military alternates.

| Rank | Candidate | Buyer and money already moving | Why correctness is bounded | Early killer | Confidence |
|---:|---|---|---|---|---|
| 1 | **Electric-grid three-part operating-instruction trainer** | Transmission operators, balancing authorities, reliability coordinators, utilities, and operator-training vendors. The NERC exam is $700; current prep is $1,695 at [ESC](https://electricsystemconsulting.com/curriculum-1), $2,999 at [Tonex](https://www.tonex.com/training-courses/nerc-system-operator-certification-training/), and $3,400 for HSI initial training ([price sheet](https://goto.hsi.com/hubfs/Industrial%20Skills/NERC%20Cert%20Exam%20Prep%20Program%20Pricing.pdf)). | [NERC COM-002-4](https://www.nerc.com/globalassets/standards/reliability-standards/com/com-002-4.pdf) requires the receiver to repeat an oral operating instruction, the issuer to confirm or reissue it, training, annual assessment, feedback, and evidence. Values and actions can be checked as typed fields. | Enterprise access and actual operating instructions are organization-specific and sensitive. No domain partner, no product. | **H** on protocol/money; **M** on product opening |
| 2 | **Railroad mandatory-directive readback trainer** | Freight/passenger railroads, contractors, and approved training providers already must fund formal initial/refresher programs under [49 CFR Part 243](https://railroads.dot.gov/railroad-safety/divisions/safety-partnerships/training-standards-rule). | [49 CFR 220.61](https://www.govinfo.gov/content/pkg/CFR-2021-title49-vol4/pdf/CFR-2021-title49-vol4-part220.pdf) requires identity/location/readiness, copying in prescribed format, immediate repetition in its entirety, dispatcher verification, time/name acknowledgment, and crew understanding. | Railroad-specific forms/rules and closed enterprise sales. Public law defines the loop, not every railroad's content. | **H** on protocol; **M** on demand wedge |
| 3 | **Fireground mayday/PAR radio drill** | Municipal departments and fire academies already run recurrent drills; public procedures teach structured LUNAR/UCAN-style mayday reports ([example state training](https://nj.gov/dca/hmfa/dca/divisions/dfs/publications/publication/reference_booklet_12.pdf)). | A scenario can check identity/unit, location, condition/problem, air/assignment, resources needed, and acknowledgment without judging prose quality. | **KILL broad product:** departments use different acronyms, radio plans, and command procedures. Only a department-configured drill survives. | **M** |
| 4 | **Crane signal-person voice drill** | Signal-person certification/training is $750 for a two-day CCO course in one current offering and $1,295 in another; both include voice-signal practice/exams ([Strategic Crane](https://strategiccrane.com/cco-signalperson), [CraneSafe](https://www.cranesafe.com/products/qualified-signal-person-training-registration)). | OSHA requires reliable dedicated radio transmission and governs voice-signal coordination ([1926.1420](https://www.osha.gov/laws-regs/regulations/standardnumber/1926/1926.1420)); the lift state constrains valid command/direction/distance/stop responses. | **KILL generic grammar:** OSHA requires the operator, signal person, and lift director to agree on voice signals before work; local vocabulary is the contract. Viable only as configurable courseware. | **H** on training/standard; **M** on grader |
| 5 | **Airline cabin emergency-command rehearsal** | Part 121 carriers and cabin-crew schools fund initial and recurrent emergency training; [14 CFR 121.417](https://www.ecfr.gov/current/title-14/chapter-I/subchapter-G/part-121/subpart-N/section-121.417) requires emergency training and drills. | A declared aircraft/airline procedure has a finite phase × hazard × exit-state command set; order and omissions are testable. | **KILL solo B2C:** exact commands vary by airline and aircraft and sit inside approved carrier programs. Requires an airline/training-center design partner. | **H** on recurrent training; **M** on shape |
| 6 | **Nuclear control-room three-way-communication trainer** | NRC licensees fund initial licensing, operating tests, and requalification. The NRC reports about 3,600 active power and 350 non-power reactor operators ([NRC](https://www.nrc.gov/reactors/operator-licensing)). | DOE human-performance guidance documents three-way/repeat-back communication; a facility procedure can check component, action, value, unit, and sender confirmation ([DOE handbook](https://www.energy.gov/ehss/articles/doe-hdbk-1028-2009)). | **KILL as an independent startup wedge:** tiny population, facility-specific procedures, nuclear-security constraints, and deeply embedded simulator/training vendors. | **H** on population/training; **M** on grammar; **H** on early kill |

### Top candidate 1: electric-grid three-part communications

This is the purest expression of the thesis. COM-002-4 does not merely publish preferred vocabulary; it specifies a state machine: issuer sends an operating instruction, receiver repeats it (not necessarily verbatim), issuer confirms or reissues, and the entity trains, assesses annually, gives feedback, and retains evidence. [H] That maps cleanly to deterministic slots. “Open breaker 7B at 230 kV” can be represented as action=`open`, object=`breaker 7B`, condition/value=`230 kV`; a changed device, action, number, or unit is a hard failure, while harmless phrasing changes are accepted.

The money is substantially better than consumer ATC: a $700 certification exam and four-figure prep programs demonstrate an employer-supported professional training budget. [H] The defensibility is also better because a utility's approved nomenclature, scenario library, assessment evidence, and instructor workflow become meaningful configuration rather than decorative content.

The attack is distribution and truth access. COM-002 proves the communication loop, not that a utility will upload operating instructions into an unknown solo vendor's browser app. Actual commands, topology, and recordings may be sensitive; procurement and security review can dwarf the code. [M] **This candidate is partner-gated.** One willing current operator or training lead is the admission ticket. Without one, stop after a synthetic demo and do not pretend the standard is customer discovery.

### Top candidate 2: railroad mandatory directives

Rail is almost as deterministic and even more explicit about readback. The receiver states identity, location, and readiness; copies the directive in the railroad's format; repeats it in its entirety; the dispatcher verifies it; both close with time and authorized name. Part 243 explicitly permits computer and simulator delivery and requires written, performance, verbal, or OJT assessments. [H] This is a ready-made evidence loop, not a vague “communication skills” rubric.

It also has a natural thin scenario: issue one track warrant/mandatory directive, let the trainee copy and repeat it, then identify every altered limit, track, time, or authority field. No conversational model is needed. [M] A small vendor could sell scenario authoring plus assessment records to training providers rather than simulate train operations.

The killer is the same, sharper: the legal minimum is public, but the prescribed forms, rule books, territorial names, and operating practices are railroad-specific. Buyers are employers, not hobbyists; access and integration are slow. [H/M] The idea merits a demo only if an instructor or retired dispatcher will supply sanitized examples and judge the acceptance lattice.

### A four-demo slate for later goal runs

This does not need to become a contest with one survivor. The strongest **demo portfolio** is: (1) ATC readback debugger, (2) grid three-part instruction, (3) railroad mandatory-directive readback, and (4) fireground mayday/PAR. [M] They can later be four separate `/goal`-style runs against one common evidence format—issued message, recognized alternatives, extracted slots, deterministic verdict, and uncertainty—without trying to share domain content or ship a platform first. Crane signals remain the first substitute if a signal-person trainer is easier to reach than a fire department.

## 3. Head-to-head: grid trainer vs ATC readback debugger

The comparison is against the narrowed ATC debugger, not the already-killed generic AI-controller subscription.

| Criterion | Grid three-part trainer | ATC readback debugger | Edge |
|---|---|---|---|
| Proof someone pays today | $700 NERC exam and $1,695–$3,400 prep/training; however, no proof found that buyers pay separately for this exact voice drill. **[H/M]** | Two exact AI radio products at $49–$79.99/year, a mature $8/month curriculum, school plans/partners, and live-human service at $179+/year. **[H]** | **ATC**—payment is proven for the exact job |
| One-person wedge | One synthetic instruction is tiny technically, but credible content, users, security, and procurement require a domain partner. **[M]** | One generic VFR tower pattern can use public FAA rules, original content, direct browser delivery, and reachable CFIs/students. **[M]** | **ATC** |
| Deterministic correctness path | COM-002's repeat/confirm/reissue loop plus typed device/action/value fields is exceptionally crisp; local nomenclature must be configured. **[H/M]** | Critical slots are deterministic, but AIM permits flexible wording and ASR ambiguity sits ahead of grading. **[H]** | **Grid** |

**Winner for the first build decision: the ATC readback debugger. [M]** It wins two of the requested three criteria: exact willingness-to-pay proof and a one-person, public-content, reachable-user wedge. Grid wins the intellectual thesis and may become the more defensible business if a utility/operator partner appears. The result is therefore not “ATC product is a go.” It is **ATC is the fastest honest experiment; grid is the best partner-contingent follow-up.**

## 4. Verdicts on the brief's secondary markets

**ICAO English proficiency: NO-GO for deterministic grading; viable only as conventional human/AI-assisted test prep. [H]** ICAO requires at least Operational Level 4 and recommends reevaluation every three years for Level 4 and six years for Level 5 ([ICAO FAQ](https://www.icao.int/personnel-licensing-faq)); ILPT charges €179 for an approved test, €69 for prep, or €228 together ([ILPT](https://www.ilpt.net/)). Demand and recurring money are real. The assessment, however, spans pronunciation, structure, vocabulary, fluency, comprehension, and interaction and still uses trained human assessors. That is precisely the kind of holistic judgment the bounded-grammar thesis promised to avoid. Phraseology drills are deterministic but collapse back into ATC training. Kill the broad “deterministic ICAO English grader.”

**Medical closed-loop / SBAR: NO-GO on broad nursing simulation; maybe a narrow medication-order check-back component. [H/M]** AHRQ TeamSTEPPS publishes SBAR, closed-loop communication, and check-back tools ([AHRQ](https://www.ahrq.gov/teamstepps-program/resources/modules/index.html)), and buyers plainly pay: [SimPhone Pro](https://simphonepro.com/nursing-schools/maine) already offers AI voice ISBARR handoffs for nursing schools, while [SBAR Plus](https://sbarplus.com/) charges $0.90 per recorded scenario or $1.50 per interactive AI doctor call. The Recommendation and Assessment portions of SBAR are clinical reasoning, not bounded correctness. A dose/order repeat-back can be slot-checked, but a broad handoff score necessarily reintroduces rubrics and model/human judgment. Exact competition plus false determinism kills the proposed general product.

**Maritime VHF and military brevity: maritime survives as a later institutional adjunct; military is a no-go for a solo public product. [H/M]** IMO SMCP is a genuine standardized safety language, and STCW requires its use for officers in charge of a navigational watch on ships of 500 GT or more ([IMO](https://www.imo.org/en/ourwork/safety/pages/standardmarinecommunicationphrases.aspx)); current GMDSS GOC courses run about £1,550–£1,960 and include practical radio procedure/examination ([Warsash](https://maritime.solent.ac.uk/courses/stcw-safety-and-security/gmdss-general-operators), [UKSA](https://www.stcwdirect.com/stcw-course/1705/gmdss-general-operators-certificate-goc/)). That is strong market proof, but the certification also tests hardware/DSC/satellite operations and is delivered through approved national training centers; a voice app is an adjunct, not the credential. Military brevity has a current public multi-service vocabulary ([ALSSA Brevity](https://www.alssa.mil/mttps/brevity/)) and a cheap hobbyist reference app ([Brevity Codes](https://apps.apple.com/no/app/brevity-codes/id1295716939)), but real scenarios, procurement, and service/unit-specific doctrine are not captured by the public word list. The accessible consumer is mostly simulation enthusiasts, not the payer for operational readiness.

## 5. Thinnest demo to cut first

Build **one inspectable VFR tower-pattern readback debugger**, not “AI ATC.” [M]

### Boundary

- One fixed callsign, one generic Class D airport layout, one runway configuration, one Cessna-like aircraft.
- Seven exchanges: initial ground call; taxi clearance; hold-short crossing; ready/tower call; takeoff/pattern instruction; landing or go-around; runway exit/ground.
- Browser push-to-talk. Controller prompts are recorded, pre-rendered, or selected from finite templates. No open-ended controller model is needed.
- The only output is: issued elements, transcript/alternatives, extracted slots, `correct` / `incorrect` / `uncertain`, and the exact mismatched safety field. No AI-written coaching.
- No account, subscription, progress system, airport database, IFR, multi-aircraft traffic, school dashboard, or loggable-training claim.

### Correctness contract

Each controller instruction becomes typed data such as:

```text
callsign=N123AB
clearance=taxi
runway=27
route=[A,B]
hold_short=18
```

The recognizer may return multiple candidates with confidence. The grader accepts any CFI-approved wording that preserves the required slots, rejects a contradictory callsign/runway/action/hold-short/value, and emits `uncertain` when the acoustic evidence cannot distinguish them. Filler words, order changes that preserve meaning, and harmless omissions must not fail just because they differ from a model sentence. [H/M]

### Kill gate before product work

Create a frozen, CFI-reviewed evaluation set of at least 200 utterances: valid concise variants, student-like hesitations, several accents/noise levels, missing elements, and single-field critical mutations. [M] The demo earns another goal only if:

- it produces **zero false accepts on the curated safety-critical mutations** and **under 5% false rejects on CFI-approved valid utterances**;
- every acoustically ambiguous critical value becomes `uncertain`, never a fabricated numeric grade;
- measured response after push-to-talk release is **p50 under 1.2 seconds and p95 under 2 seconds** on the target browser/network—not because those numbers are known industry facts, but because they are a deliberately demanding usability gate;
- three current CFIs agree on the acceptance rules, at least two would assign the drill, and ten students can use it without coached setup; and
- at least three students pay a small one-time amount or a school agrees to a paid seat pilot. Praise is not payment evidence.

If recognition misses that bar, kill the ATC implementation rather than adding a model judge. If it passes, the next work is more scenario evidence—not accounts, subscriptions, or a platform. In parallel, the grid, rail, and fireground demos can later be run as independent goals using the same evidence shape, exactly as a build-all slate rather than a winner-take-all idea contest.
