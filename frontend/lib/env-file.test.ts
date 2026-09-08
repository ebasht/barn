import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { mergeEnvRows, parseEnvFile } from "./env-file.ts";

describe("env-file", () => {
  it("parses common dotenv syntax and keeps the last duplicate", () => {
    assert.deepEqual(
      parseEnvFile('\uFEFF# comment\nexport PORT=3000\nTOKEN="a\\nb"\nPORT=4000 # override\nEMPTY=\n'),
      [
        { key: "PORT", value: "4000" },
        { key: "TOKEN", value: "a\nb" },
        { key: "EMPTY", value: "" },
      ],
    );
  });

  it("replaces existing rows and appends new values", () => {
    assert.deepEqual(
      mergeEnvRows(
        [{ key: "PORT", value: "3000" }, { key: "DEBUG", value: "1" }],
        [{ key: "PORT", value: "8080" }, { key: "API_URL", value: "https://example.test" }],
      ),
      [
        { key: "PORT", value: "8080" },
        { key: "DEBUG", value: "1" },
        { key: "API_URL", value: "https://example.test" },
      ],
    );
  });

  it("rejects malformed declarations", () => {
    assert.throws(() => parseEnvFile("NOT A VARIABLE"));
  });
});
