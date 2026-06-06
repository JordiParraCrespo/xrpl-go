// extract.mjs parses the xrpl.js request model interfaces and emits a pinned
// JSON spec describing the required/optional fields of every *Request interface.
//
// The Go conformance harness consumes the generated jsspec.json to assert that
// each Go query-request Validate() enforces exactly the fields that xrpl.js
// marks as required (un-"?"-marked) on the matching request interface.
//
// Usage:
//   node extract.mjs <path-to-xrpl.js/packages/xrpl/src/models/methods> <out.json>
//
// The spec is keyed by the TypeScript interface name (e.g. "AccountInfoRequest")
// so that Go requests which split a single JS request into several structs can
// each point at the right interface.

import { readFileSync, readdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

const methodsDir = process.argv[2]
const outPath = process.argv[3]
if (!methodsDir || !outPath) {
  console.error('usage: node extract.mjs <methodsDir> <out.json>')
  process.exit(1)
}

// Structural fields that live on BaseRequest and are not validated as
// request parameters by clients.
const STRUCTURAL = new Set(['command', 'api_version', 'id'])

// Strip line and block comments so they do not confuse the field scanner.
function stripComments(src) {
  return src
    .replace(/\/\*[\s\S]*?\*\//g, '')
    .replace(/\/\/[^\n]*/g, '')
}

// Collapse nested object/array literals so only the top-level fields of the
// interface remain. Inner fields (e.g. the items of `oracles: { account:
// string }[]`) are properties of a nested shape, not request parameters, and
// must not be mistaken for top-level required fields.
function removeNested(body) {
  let prev
  do {
    prev = body
    body = body.replace(/\{[^{}]*\}/g, ' ')
  } while (body !== prev)
  return body
}

// Given the source of one file, find every `export interface XxxRequest ... { body }`
// and return [{ name, requiredFields, optionalFields }].
function parseRequests(src) {
  const out = []
  const re = /export\s+interface\s+([A-Za-z0-9_]*Request)\b([^{]*)\{/g
  let m
  while ((m = re.exec(src)) !== null) {
    const name = m[1]
    // Walk braces from the opening { to find the matching close.
    let depth = 1
    let i = re.lastIndex
    const start = i
    while (i < src.length && depth > 0) {
      const ch = src[i]
      if (ch === '{') depth++
      else if (ch === '}') depth--
      i++
    }
    const body = removeNested(src.slice(start, i - 1))
    const required = []
    const optional = []
    // Match top-level field declarations: `name: type` or `name?: type`.
    const fieldRe = /(^|\n)\s*([a-z_][A-Za-z0-9_]*)(\?)?\s*:/g
    let f
    while ((f = fieldRe.exec(body)) !== null) {
      const field = f[2]
      const isOptional = f[3] === '?'
      if (STRUCTURAL.has(field)) continue
      if (isOptional) optional.push(field)
      else required.push(field)
    }
    out.push({ name, required, optional })
  }
  return out
}

const spec = {}
const files = readdirSync(methodsDir).filter((f) => f.endsWith('.ts'))
for (const file of files) {
  const src = stripComments(readFileSync(join(methodsDir, file), 'utf8'))
  for (const iface of parseRequests(src)) {
    spec[iface.name] = {
      file,
      required: iface.required,
      optional: iface.optional,
    }
  }
}

const sorted = {}
for (const key of Object.keys(spec).sort()) sorted[key] = spec[key]
writeFileSync(outPath, JSON.stringify(sorted, null, 2) + '\n')
console.error(`wrote ${Object.keys(sorted).length} request interfaces to ${outPath}`)
