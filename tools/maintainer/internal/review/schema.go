package review

import "encoding/json"

// lensSchema constrains every model lens to the same finding shape.
var lensSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "file": {"type": "string", "description": "Repo-relative path, forward slashes."},
          "line": {"type": "integer", "description": "1-based line on the new side of the diff; 0 if not tied to a line."},
          "severity": {"type": "string", "enum": ["blocker", "concern", "nit"]},
          "category": {"type": "string", "description": "Short kebab-case slug, e.g. correctness, structure, protocol, tests."},
          "title": {"type": "string", "description": "One-line claim, no trailing period."},
          "body": {"type": "string", "description": "What is wrong, the concrete failure path, and its effect. Markdown."},
          "suggestion": {"type": "string", "description": "Concrete fix; may be empty."},
          "source": {"type": "string", "description": "Rule or reference cited, e.g. TX-06, go-errors/error-wrapping, XLS-66."}
        },
        "required": ["file", "line", "severity", "title", "body"],
        "additionalProperties": false
      }
    }
  },
  "required": ["findings"],
  "additionalProperties": false
}`)

// verifySchema is the verifier's answer: one verdict per finding ID.
var verifySchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "results": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": {"type": "integer"},
          "verdict": {"type": "string", "enum": ["confirmed", "refuted", "uncertain"]},
          "reason": {"type": "string", "description": "The evidence: file:line you read and what it shows."},
          "severity": {"type": "string", "enum": ["blocker", "concern", "nit"], "description": "Corrected severity for confirmed findings."}
        },
        "required": ["id", "verdict", "reason"],
        "additionalProperties": false
      }
    }
  },
  "required": ["results"],
  "additionalProperties": false
}`)
