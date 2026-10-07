import { execFileSync } from "node:child_process";

const files = execFileSync("go", ["-C", "apps/api", "list", "-f", "{{range .GoFiles}}{{$.Dir}}/{{.}} {{end}}{{range .TestGoFiles}}{{$.Dir}}/{{.}} {{end}}", "./..."], { encoding: "utf8" })
  .trim()
  .split(/\s+/)
  .filter(Boolean);

if (files.length === 0) {
  console.error("No Go source files were found.");
  process.exit(1);
}

const unformatted = execFileSync("gofmt", ["-l", ...files], { encoding: "utf8" }).trim();
if (unformatted) {
  console.error(`Go files need formatting:\n${unformatted}`);
  process.exit(1);
}

console.log(`gofmt: ${files.length} files formatted`);
