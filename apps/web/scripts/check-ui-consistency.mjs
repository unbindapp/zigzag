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
    name: "raw <button>, <input>, <select> or <textarea> element",
    pattern: /<(?:button|input|select|textarea)(?=[\s>/])/g,
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
  {
    name: "spacing outside the ui.md scale",
    pattern: /\b(?:gap|gap-x|gap-y|space-x|space-y)-(?:5|7)\b/g,
    allowedIn: ["src/components/ui/"],
  },
  {
    name: "radius larger than rounded-lg",
    pattern: /\brounded-(?:xl|2xl|3xl)\b/g,
    allowedIn: ["src/components/ui/"],
  },
  {
    name: "hardcoded color outside the palette",
    pattern: /["'`]#[0-9a-fA-F]{3,8}["'`]/g,
    allowedIn: [
      "src/lib/helpers/colors.ts",
      "src/components/providers/themes.ts",
      "src/components/shared/color-picker.tsx",
      "src/components/icons/google.tsx",
    ],
    skipTests: true,
  },
  {
    name: "hotkey registered outside the scoped registry",
    pattern: /\buseHotkeys?\(/g,
    allowedIn: ["src/lib/hotkeys/"],
  },
  {
    name: "editor.setEditable without emitUpdate=false (emits an update that autosaves)",
    pattern: /\.setEditable\([^,()]*\)/g,
    allowedIn: [],
  },
  {
    name: "confirm label without a noun (use Delete project, Remove member)",
    pattern: /confirmLabel="(?:Delete|Remove|Revoke|Archive|Trash)"/g,
    allowedIn: [],
  },
  {
    name: 'empty state title "No ..." without "yet"',
    pattern: /title(?:=|: )"No (?![^"]*\b(?:yet|matching|selected)\b)[^"]*"/g,
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
    if (rule.skipTests && rel.endsWith(".node-test.ts")) continue;
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
