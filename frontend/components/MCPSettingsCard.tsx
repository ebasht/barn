"use client";

import { useCallback, useEffect, useState } from "react";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { api, getApiBase } from "@/lib/api";
import { useI18n } from "@/lib/i18n/context";
import type { MCPSettings } from "@/lib/types";

export function MCPSettingsCard() {
  const { t, formatDateTime } = useI18n();
  const [settings, setSettings] = useState<MCPSettings | null>(null);
  const [name, setName] = useState("");
  const [writes, setWrites] = useState(false);
  const [key, setKey] = useState("");
  const [endpoint, setEndpoint] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [confirmation, setConfirmation] = useState<"rotate" | "revoke" | null>(null);

  const load = useCallback(async () => {
    const value = await api.getMCPSettings();
    setSettings(value);
    setName(value.instance_name);
    setWrites(value.allow_writes);
  }, []);

  useEffect(() => {
    setEndpoint(`${getApiBase()}/mcp`);
    load().catch((e: unknown) => setError(e instanceof Error ? e.message : t("mcp.failed")));
  }, [load, t]);

  const run = async (action: () => Promise<void>) => {
    setBusy(true); setError(""); setMessage("");
    try { await action(); } catch (e) { setError(e instanceof Error ? e.message : t("mcp.failed")); }
    finally { setBusy(false); }
  };
  const generate = () => run(async () => {
    const value = await api.generateMCPKey();
    setKey(value.token);
    setConfirmation(null);
    await load();
  });
  const revoke = () => run(async () => {
    await api.revokeMCPKey();
    setKey(""); setConfirmation(null);
    await load();
    setMessage(t("mcp.revoked"));
  });
  const copy = (value: string) => run(async () => {
    if (!navigator.clipboard) throw new Error(t("mcp.copyFailed"));
    await navigator.clipboard.writeText(value);
    setMessage(t("mcp.copied"));
  });
  const config = `[mcp_servers.barn]\nurl = ${JSON.stringify(endpoint)}\nbearer_token_env_var = "BARN_MCP_TOKEN"`;

  return (
    <section className="card" style={{ marginBottom: "1rem" }}>
      <h2 className="section-title">{t("mcp.title")}</h2>
      <p className="muted">{t("mcp.hint")}</p>
      {error && <div className="alert alert-error" role="alert">{error}</div>}
      {message && <div className="alert alert-success" role="status">{message}</div>}
      {!settings && !error && <p className="muted">{t("common.loading")}</p>}
      {!settings && error && <button className="btn btn-secondary" disabled={busy} onClick={() => run(load)}>{t("common.retry")}</button>}
      {settings && <>
        <form onSubmit={(e) => { e.preventDefault(); void run(async () => {
          const value = await api.updateMCPSettings({ instance_name: name.trim(), allow_writes: writes });
          setSettings(value); setName(value.instance_name); setWrites(value.allow_writes);
          setMessage(t("mcp.saved"));
        }); }}>
          <div className="field">
            <label className="label" htmlFor="mcp-panel-name">{t("mcp.panelName")}</label>
            <input id="mcp-panel-name" className="input" value={name} onChange={(e) => setName(e.target.value)} required disabled={busy} />
            <p className="muted">{t("mcp.panelHint")}</p>
          </div>
          <label style={{ display: "flex", gap: "0.5rem", marginBottom: "0.5rem" }}>
            <input type="checkbox" checked={writes} disabled={busy} onChange={(e) => setWrites(e.target.checked)} />
            {t("mcp.allowWrites")}
          </label>
          <p className="muted">{t("mcp.scopeHint")}</p>
          <button className="btn" type="submit" disabled={busy || !name.trim()}>{busy ? t("common.saving") : t("common.save")}</button>
        </form>
        <p>{t("mcp.keyStatus")}: <strong>{t(settings.key_configured ? "mcp.active" : "mcp.inactive")}</strong></p>
        {settings.key_created_at && <p className="muted">{t("mcp.created")}: {formatDateTime(settings.key_created_at)}</p>}
        <div className="form-actions">
          <button className="btn btn-secondary" disabled={busy} onClick={() => settings.key_configured ? setConfirmation("rotate") : void generate()}>
            {t(settings.key_configured ? "mcp.rotate" : "mcp.generate")}
          </button>
          {settings.key_configured && <button className="btn btn-secondary" disabled={busy} onClick={() => setConfirmation("revoke")}>{t("mcp.revoke")}</button>}
        </div>
      </>}
      {key && <div style={{ marginTop: "1rem" }}>
        <p>{t("mcp.once")}</p>
        <label className="label" htmlFor="mcp-new-key">{t("mcp.newKey")}</label>
        <input id="mcp-new-key" className="input" readOnly value={key} autoComplete="off" spellCheck={false} onFocus={(e) => e.target.select()} />
        <div className="form-actions" style={{ marginTop: "0.5rem" }}>
          <button className="btn btn-secondary" disabled={busy} onClick={() => copy(key)}>{t("mcp.copyKey")}</button>
          <button className="btn btn-secondary" disabled={busy} onClick={() => setKey("")}>{t("mcp.hideKey")}</button>
        </div>
      </div>}
      {endpoint && <div style={{ marginTop: "1rem" }}>
        <p>{t("mcp.endpoint")}: <code style={{ overflowWrap: "anywhere" }}>{endpoint}</code></p>
        <p className="muted">{t("mcp.connectHint")}</p>
        <pre style={{ overflowX: "auto", whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}>{config}</pre>
        <button className="btn btn-secondary" disabled={busy} onClick={() => copy(config)}>{t("mcp.copyConfig")}</button>
      </div>}
      <ConfirmDialog open={confirmation !== null} title={t(confirmation === "revoke" ? "mcp.revoke" : "mcp.rotate")}
        message={t(confirmation === "revoke" ? "mcp.revokeConfirm" : "mcp.rotateConfirm")}
        confirmLabel={t(confirmation === "revoke" ? "mcp.revoke" : "mcp.rotate")} danger busy={busy}
        onConfirm={() => { if (confirmation === "revoke") void revoke(); else void generate(); }}
        onCancel={() => { if (!busy) setConfirmation(null); }} />
    </section>
  );
}
