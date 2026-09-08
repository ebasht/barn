"use client";

import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { KeyValueEditor } from "@/components/KeyValueEditor";
import { SiteTabs } from "@/components/SiteTabs";
import { api, ApiError } from "@/lib/api";
import { useI18n } from "@/lib/i18n/context";
import type { EnvVar, SecretMeta, Site } from "@/lib/types";

export default function SiteSecretsPage() {
  const { id } = useParams<{ id: string }>();
  const { t, formatDateTime } = useI18n();
  const [site, setSite] = useState<Site | null>(null);
  const [secrets, setSecrets] = useState<SecretMeta[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [draft, setDraft] = useState<EnvVar[]>([{ key: "", value: "" }]);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    try {
      const [s, sec] = await Promise.all([
        api.getSite(id),
        api.listSecrets(id),
      ]);
      setSite(s);
      setSecrets(sec);
      setError(null);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : t("secrets.loadFailed"));
    }
  }, [id, t]);

  useEffect(() => {
    load();
  }, [load]);

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    const toSave = draft.filter((row) => row.key.trim() && row.value);
    if (toSave.length === 0) return;

    setSaving(true);
    setError(null);
    try {
      for (const row of toSave) {
        await api.upsertSecret(id, row.key.trim(), row.value);
      }
      setDraft([{ key: "", value: "" }]);
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("secrets.saveFailed"));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (key: string) => {
    if (!confirm(t("secrets.deleteConfirm", { key }))) return;
    try {
      await api.deleteSecret(id, key);
      await load();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("secrets.deleteFailed"));
    }
  };

  return (
    <div>
      <h1>{site?.name ?? t("secrets.title")}</h1>
      <SiteTabs siteId={id} active="secrets" />

      <p style={{ color: "var(--muted)", fontSize: "0.875rem" }}>
        {t("secrets.hint", {
          gitToken: "GIT_TOKEN",
          githubToken: "GITHUB_TOKEN",
          gitSshKey: "GIT_SSH_KEY",
          sshUrl: "git@github.com:…",
        })}
      </p>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="card" style={{ marginBottom: "1.5rem" }}>
        <h2>{t("secrets.storedSecrets")}</h2>
        {secrets.length === 0 ? (
          <p style={{ color: "var(--muted)" }}>{t("secrets.noSecrets")}</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>{t("common.name")}</th>
                <th>{t("common.updated")}</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {secrets.map((s, idx) => (
                <tr key={s.key || `secret-${idx}`}>
                  <td>
                    <code>{s.key || t("secrets.missingName")}</code>
                  </td>
                  <td>{formatDateTime(s.updated_at)}</td>
                  <td>
                    <button
                      type="button"
                      className="btn btn-danger"
                      style={{ padding: "0.25rem 0.5rem", fontSize: "0.75rem" }}
                      onClick={() => handleDelete(s.key)}
                      disabled={!s.key}
                    >
                      {t("common.delete")}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <form onSubmit={handleAdd} className="card">
        <h2>{t("secrets.addOrRotate")}</h2>
        <p style={{ color: "var(--muted)", fontSize: "0.875rem", marginBottom: "0.75rem" }}>
          {t("secrets.addHint")}
        </p>
        <KeyValueEditor
          rows={draft}
          onChange={setDraft}
          valueInputType="password"
          keyPlaceholder={t("secrets.keyPlaceholder")}
          valuePlaceholder={t("secrets.valuePlaceholder")}
          addLabel={t("secrets.addAnother")}
          existingKeys={secrets.map((secret) => secret.key)}
        />
        <button type="submit" className="btn" disabled={saving}>
          {saving ? t("common.saving") : t("secrets.saveSecrets")}
        </button>
      </form>
    </div>
  );
}
