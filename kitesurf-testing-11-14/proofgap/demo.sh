#!/bin/sh
set -eu
cd "$(dirname "$0")"
npm test
npm run check
node src/cli.mjs workloads/checkout/workload.json --demo-check
PROOFGAP_OUT=artifacts/shipping node src/cli.mjs workloads/shipping/workload.json --demo-check
node scripts/verify-controls.mjs
printf '\nOpen artifacts/report.html for the checkout audit.\nOpen artifacts/shipping/report.html for the second workload.\n'
