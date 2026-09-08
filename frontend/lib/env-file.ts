import type { EnvVar } from "./types";

const ENV_KEY = /^[A-Za-z_][A-Za-z0-9_.-]*$/;

function unquote(value: string): string {
  if (value.length < 2) return value;
  const quote = value[0];
  if ((quote !== '"' && quote !== "'") || value.at(-1) !== quote) return value;
  const inner = value.slice(1, -1);
  if (quote === "'") return inner;
  return inner.replace(/\\n/g, "\n").replace(/\\r/g, "\r").replace(/\\t/g, "\t").replace(/\\"/g, '"').replace(/\\\\/g, "\\");
}

/** Parse a dotenv-style file. Later occurrences of the same key win. */
export function parseEnvFile(contents: string): EnvVar[] {
  const values = new Map<string, string>();

  contents.replace(/^\uFEFF/, "").split(/\r?\n/).forEach((line, index) => {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) return;

    const declaration = trimmed.replace(/^export\s+/, "");
    const equals = declaration.indexOf("=");
    if (equals < 1) throw new Error(`line ${index + 1}`);

    const key = declaration.slice(0, equals).trim();
    if (!ENV_KEY.test(key)) throw new Error(`line ${index + 1}`);

    let value = declaration.slice(equals + 1).trim();
    if (!value.startsWith('"') && !value.startsWith("'")) {
      value = value.replace(/\s+#.*$/, "").trimEnd();
    }
    values.set(key, unquote(value));
  });

  return Array.from(values, ([key, value]) => ({ key, value }));
}

/** Merge imported values into editor rows, replacing matching keys in place. */
export function mergeEnvRows(rows: EnvVar[], imported: EnvVar[]): EnvVar[] {
  const next = rows.filter((row) => row.key.trim() || row.value);
  for (const item of imported) {
    const index = next.findIndex((row) => row.key.trim() === item.key);
    if (index === -1) next.push(item);
    else next[index] = item;
  }
  return next;
}
