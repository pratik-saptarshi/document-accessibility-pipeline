"use strict";

const fs = require("node:fs");
const path = require("node:path");

for (const lane of ["generation", "remediation"]) {
  const schemaPath = path.join(
    __dirname,
    "..",
    "docs",
    "workflows",
    lane,
    "profiles",
    "profile.schema.json",
  );
  const schema = JSON.parse(fs.readFileSync(schemaPath, "utf8"));
  const patterns = [];
  const visit = (value) => {
    if (Array.isArray(value)) {
      value.forEach(visit);
    } else if (value && typeof value === "object") {
      for (const [key, child] of Object.entries(value)) {
        if (key === "pattern") patterns.push(child);
        visit(child);
      }
    }
  };
  visit(schema);
  if (patterns.length === 0) {
    throw new Error(`${lane} schema has no regular-expression patterns`);
  }
  for (const pattern of patterns) new RegExp(pattern);
}
