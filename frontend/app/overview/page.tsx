"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { ScopeHeading } from "@/components/ScopeHeading";
import { ServerStatusBadge } from "@/components/ServerBadges";
import { api, ApiError } from "@/lib/api";
import { useBarnScope } from "@/lib/barn-scope-context";
import { useI18n } from "@/lib/i18n/context";
import type { ServersOverview } from "@/lib/types";

export default function MasterOverviewPage() {
  const { t } = useI18n();
  const { isMasterMode, isGlobalScope, nodes, loadingTargets, refreshTargets } = useBarnScope();
  const [overview, setOverview] = useState<ServersOverview | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!isMasterMode) return;
    try {
      const value = await api.getServersOverview();
      setOverview(value);
      setError(null);
    } catch (reason) {
      setError(reason instanceof ApiError ? reason.message : t("master.overviewLoadFailed"));
    }
  }, [isMasterMode, t]);

  useEffect(() => {
    void load();
    const timer = setInterval(() => {
      void load();
      void refreshTargets();
    }, 30_000);
    return () => clearInterval(timer);
  }, [load, refreshTargets]);

  const attention = useMemo(
    () => nodes.filter((node) => node.status !== "online" || node.open_incidents > 0 || (node.applications?.unhealthy ?? 0) > 0),
    [nodes],
  );
  const versions = useMemo(() => new Set(nodes.map((node) => node.version).filter(Boolean)), [nodes]);

  if (!isMasterMode) {
    return <div className="card"><p>{t("master.masterOnly")}</p><Link href="/sites">{t("nav.sites")}</Link></div>;
  }
  if (!isGlobalScope) return null;

  return (
    <div className="master-overview">
      <ScopeHeading page={t("nav.overview")} />
      <div className="page-header master-overview-header">
        <div>
          <div className="master-kicker">MASTER</div>
          <h1>{t("master.title")}</h1>
          <p className="muted">{t("master.subtitle")}</p>
        </div>
        <Link href="/servers/new" className="btn">{t("master.connectServer")}</Link>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {(loadingTargets || !overview) && !error ? <p className="muted">{t("common.loading")}</p> : overview && (
        <div className="master-summary-grid">
          <Link href="/servers" className="master-summary-card">
            <span>{t("master.summaryServers")}</span>
            <strong>{overview.servers_online} / {overview.servers_total}</strong>
            <small>{t("master.summaryOnline")}</small>
          </Link>
          <div className="master-summary-card">
            <span>{t("master.summaryAttention")}</span>
            <strong>{attention.length}</strong>
            <small>{overview.open_incidents > 0 ? t("servers.summaryIncidents", { count: overview.open_incidents }) : t("master.noOpenIncidents")}</small>
          </div>
          <div className="master-summary-card">
            <span>{t("master.summaryApps")}</span>
            <strong>{overview.apps_running} / {overview.apps_total}</strong>
            <small>{t("servers.summaryAppsHint", { unhealthy: overview.apps_unhealthy })}</small>
          </div>
          <div className="master-summary-card">
            <span>{t("master.summaryVersions")}</span>
            <strong>{versions.size || "—"}</strong>
            <small>{versions.size > 1 ? t("master.multipleVersions") : t("master.knownVersions")}</small>
          </div>
        </div>
      )}

      <section className="card master-attention">
        <div className="master-section-heading">
          <div><h2>{t("master.needsAttention")}</h2><p className="muted">{t("master.needsAttentionHint")}</p></div>
          <Link href="/servers">{t("master.viewAllServers")}</Link>
        </div>
        {attention.length === 0 ? (
          <div className="master-empty-state"><span aria-hidden>✓</span><p>{t("master.allHealthy")}</p></div>
        ) : (
          <div className="master-attention-list">
            {attention.map((node) => (
              <Link href={`/servers/${node.id}`} className="master-attention-row" key={node.id}>
                <ServerStatusBadge status={node.status} />
                <span><strong>{node.name}</strong><small>{node.status === "offline" ? t("master.serverOffline") : node.open_incidents > 0 ? t("servers.openIncidents", { count: node.open_incidents }) : t("master.appsUnhealthy", { count: node.applications?.unhealthy ?? 0 })}</small></span>
                <span aria-hidden>→</span>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
