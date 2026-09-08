"use client";

import type { EnvVar } from "@/lib/types";
import { useRef, useState } from "react";
import { useI18n } from "@/lib/i18n/context";
import { mergeEnvRows, parseEnvFile } from "@/lib/env-file";

type Props = {
  rows: EnvVar[];
  onChange: (rows: EnvVar[]) => void;
  valueInputType?: "text" | "password";
  keyPlaceholder?: string;
  valuePlaceholder?: string;
  addLabel?: string;
  showRemove?: boolean;
  existingKeys?: string[];
};

export function KeyValueEditor({
  rows,
  onChange,
  valueInputType = "text",
  keyPlaceholder = "NAME",
  valuePlaceholder = "value",
  addLabel,
  showRemove = true,
  existingKeys = [],
}: Props) {
  const { t } = useI18n();
  const keyBeforeEdit = useRef<Record<number, string>>({});
  const [importOpen, setImportOpen] = useState(false);
  const [importText, setImportText] = useState("");
  const [importError, setImportError] = useState(false);
  const resolvedAddLabel = addLabel ?? t("kvEditor.addRow");

  const setAt = (index: number, patch: Partial<EnvVar>) => {
    const next = [...rows];
    next[index] = { ...next[index], ...patch };
    onChange(next);
  };

  const removeAt = (index: number) => {
    onChange(rows.filter((_, i) => i !== index));
  };

  const finishKeyEdit = (index: number) => {
    const key = rows[index]?.key.trim();
    const oldKey = keyBeforeEdit.current[index] ?? "";
    if (!key || key === oldKey) return;
    const duplicateIndex = rows.findIndex((row, i) => i !== index && row.key.trim() === key);
    const alreadySaved = existingKeys.includes(key);
    if (duplicateIndex < 0 && !alreadySaved) return;

    if (!confirm(t("kvEditor.replaceConfirm", { key }))) {
      setAt(index, { key: oldKey });
      return;
    }
    if (duplicateIndex >= 0) {
      onChange(rows.filter((_, i) => i !== duplicateIndex));
    }
  };

  const applyImport = () => {
    try {
      const imported = parseEnvFile(importText);
      onChange(mergeEnvRows(rows, imported));
      setImportError(false);
      setImportText("");
      setImportOpen(false);
    } catch {
      setImportError(true);
    }
  };

  return (
    <div className="kv-editor">
      <table className="kv-editor-table">
        <colgroup>
          <col className="kv-col-name" />
          <col className="kv-col-value" />
          {showRemove && <col className="kv-col-action" />}
        </colgroup>
        <thead>
          <tr>
            <th>{t("kvEditor.name")}</th>
            <th>{t("kvEditor.value")}</th>
            {showRemove && <th aria-label={t("common.actions")} />}
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td colSpan={showRemove ? 3 : 2} style={{ color: "var(--muted)" }}>
                {t("kvEditor.empty", { addLabel: resolvedAddLabel })}
              </td>
            </tr>
          ) : (
            rows.map((row, i) => (
              <tr key={`kv-${i}-${row.key || "new"}`}>
                <td>
                  <input
                    className="input"
                    placeholder={keyPlaceholder}
                    value={row.key}
                    onFocus={() => { keyBeforeEdit.current[i] = row.key; }}
                    onChange={(e) => setAt(i, { key: e.target.value })}
                    onBlur={() => finishKeyEdit(i)}
                    autoComplete="off"
                    aria-label={t("kvEditor.nameN", { n: i + 1 })}
                  />
                </td>
                <td>
                  <input
                    className="input"
                    type={valueInputType}
                    placeholder={valuePlaceholder}
                    value={row.value}
                    onChange={(e) => setAt(i, { value: e.target.value })}
                    autoComplete="off"
                    aria-label={t("kvEditor.valueN", { n: i + 1 })}
                  />
                </td>
                {showRemove && (
                  <td>
                    <button
                      type="button"
                      className="btn btn-secondary kv-remove"
                      onClick={() => removeAt(i)}
                      aria-label={t("kvEditor.removeRow", { n: i + 1 })}
                    >
                      ×
                    </button>
                  </td>
                )}
              </tr>
            ))
          )}
        </tbody>
      </table>
      <button
        type="button"
        className="btn btn-secondary"
        onClick={() => onChange([...rows, { key: "", value: "" }])}
      >
        {resolvedAddLabel}
      </button>
      <button
        type="button"
        className="btn btn-secondary"
        onClick={() => {
          setImportError(false);
          setImportOpen(true);
        }}
      >
        {t("kvEditor.pasteList")}
      </button>
      {importOpen && (
        <div
          className="modal-backdrop"
          role="presentation"
          onClick={() => setImportOpen(false)}
        >
          <div
            className="modal card"
            role="dialog"
            aria-modal="true"
            aria-labelledby="kv-import-title"
            onClick={(event) => event.stopPropagation()}
          >
            <h2 id="kv-import-title">{t("kvEditor.pasteTitle")}</h2>
            <p style={{ color: "var(--muted)", fontSize: "0.875rem" }}>
              {t("kvEditor.pasteHint")}
            </p>
            <textarea
              className="textarea"
              rows={12}
              value={importText}
              onChange={(event) => setImportText(event.target.value)}
              placeholder={"PORT=3000\nAPI_URL=https://example.com"}
              autoFocus
            />
            {importError && <div className="alert alert-error">{t("kvEditor.importFailed")}</div>}
            <div className="confirm-dialog-actions">
              <button type="button" className="btn btn-secondary" onClick={() => setImportOpen(false)}>
                {t("common.cancel")}
              </button>
              <button type="button" className="btn" onClick={applyImport} disabled={!importText.trim()}>
                {t("kvEditor.applyList")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
