#!/bin/sh
set -eu
cd "$(dirname "$0")"
if [ ! -d node_modules/puppeteer-core ]; then npm ci --no-audit --no-fund; fi
npm run check
npm test
node src/demo.mjs
