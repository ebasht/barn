"use client";

import { Fragment } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ScopeHeading } from "@/components/ScopeHeading";
import { useBarnScope } from "@/lib/barn-scope-context";
import { useI18n } from "@/lib/i18n/context";

const LOCAL_ONLY_PREFIXES = ["/databases", "/backups", "/payments", "/notifications"];

export function BarnScopeBoundary({ children }: { children: React.ReactNode }) {
  const pathname = usePathname() || "";
  const { activeBarn, isGlobalScope, isMasterMode } = useBarnScope();
  const { t } = useI18n();
  const isLocalOnly =
    LOCAL_ONLY_PREFIXES.some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`)) ||
    pathname === "/sites/new" ||
    pathname.startsWith("/sites/");
  const canManagePanel = activeBarn?.kind === "local" || activeBarn?.kind === "barn";
  // Remount page content when the active Barn changes so lists/forms reload
  // against the new remote API target (same route, only ?barn= changes).
  const contentKey = !isMasterMode
    ? "standalone"
    : isGlobalScope
      ? "all"
      : activeBarn?.id ?? "pending";

  if (!isMasterMode || !isLocalOnly || canManagePanel) {
    return <Fragment key={contentKey}>{children}</Fragment>;
  }

  return (
    <div className="scope-limited-page" key={contentKey}>
      <ScopeHeading page={t("master.localAction")} />
      <div className="card scope-limitation-card">
        <div className="scope-limitation-icon" aria-hidden>!</div>
        <div>
          <h1>{t("master.scopeUnavailable")}</h1>
          <p className="muted">
            {isGlobalScope
              ? t("master.selectLocalBarnHint")
              : t("master.agentScopeHint", { name: activeBarn?.name || t("master.unknownBarn") })}
          </p>
          <Link className="btn btn-secondary" href="/overview?barn=all">{t("master.backToOverview")}</Link>
        </div>
      </div>
    </div>
  );
}
