import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative, sep } from "node:path";

const root = new URL("..", import.meta.url).pathname;
const srcDir = join(root, "src");
const generated = new Set(["src/routeTree.gen.ts", "src/lib/server/client.gen.ts"]);

const rules = [
  {
    name: "arbitrary Tailwind value",
    pattern:
      /(?<![\w[-])[a-z][a-z0-9]*(?:-[a-z0-9.]+)*-(?:\[[^\]\s"'`]+\]|\(--[^)\s"'`]+\))(?!\/?[\w-]*:)/g,
    allowedIn: ["src/components/ui/", "src/components/layout/"],
  },
  {
    name: "raw <button> or <input> element",
    pattern: /<(?:button|input)(?=[\s>/])/g,
    allowedIn: ["src/components/ui/"],
  },
  {
    name: "page-title text size",
    pattern: /\btext-[2-9]xl\b/g,
    allowedIn: ["src/components/layout/"],
  },
  {
    name: "alert()",
    pattern: /\balert\(/g,
    allowedIn: [],
  },
];

function walk(dir) {
  return readdirSync(dir).flatMap((entry) => {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) return walk(full);
    return /\.(tsx?|mjs|js)$/.test(entry) ? [full] : [];
  });
}

function lineOf(text, index) {
  return text.slice(0, index).split("\n").length;
}

const violations = [];
for (const file of walk(srcDir)) {
  const rel = relative(root, file).split(sep).join("/");
  if (generated.has(rel)) continue;
  const text = readFileSync(file, "utf8");
  for (const rule of rules) {
    if (rule.allowedIn.some((prefix) => rel.startsWith(prefix))) continue;
    for (const match of text.matchAll(rule.pattern)) {
      violations.push(`${rel}:${lineOf(text, match.index)}  ${rule.name}: ${match[0]}`);
    }
  }
}

if (violations.length > 0) {
  console.error(`UI consistency check failed with ${violations.length} problem(s):`);
  for (const violation of violations) console.error(`  ${violation}`);
  process.exit(1);
}
console.log("UI consistency check passed.");
