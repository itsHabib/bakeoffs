"""Comprehension questions with known answers.

Each question has a single, distinctive answer so grading is deterministic
exact-match (see comprehension.grade). Numeric answers are matched on a word
boundary so "6" never matches "16"; string answers are substring-matched after
normalization, so a chatty model that says "Theo Vance (VP Engineering)" still
scores the "Theo Vance" answer. The model is the SUBJECT under test — never the
grader. Answers are derived by hand from the corpus values.

Shape: {payload_name: [(question, expected_answer, numeric_bool), ...]}
"""

QUESTIONS = {
    "deploy_record": [
        ("How many replicas does the service run?", "6", True),
        ("What region is the service in?", "eu-west-2", False),
        ("What is the commit hash?", "d41ab9c", False),
    ],
    "flight_table": [
        ("What is the departure time of flight EK 409?", "12:10", False),
        ("Which flight departs from MEL?", "QF 442", False),
        ("What is the destination of flight NZ 118?", "SYD", False),
        ("How many flights are listed?", "5", True),
    ],
    "metric_table": [
        ("What is the p99_ms of the recommend service?", "430", True),
        ("What is the rps of the catalog service?", "5120", True),
        ("Which service has an error_pct of 12?", "recommend", False),
    ],
    "verify_report": [
        ("What is the status of the unit-tests check?", "fail", False),
        ("Which check has a status of warn?", "boundary-law", False),
        ("What is the detail of the unit-tests check?", "2 of 314 failed", False),
    ],
    "org_chart": [
        ("What is the headcount of Lead Platform?", "9", True),
        ("Who does Suki Ito report to directly?", "Theo Vance", False),
        ("What is the name of the VP Revenue?", "Dario Fuchs", False),
        ("How many direct reports does Theo Vance have?", "3", True),
    ],
    "network_graph": [
        ("How many outgoing edges does the api node have?", "3", True),
        ("What single node does the queue node point to?", "db", False),
        ("Does the web node point to auth? Answer yes or no.", "yes", False),
    ],
}
