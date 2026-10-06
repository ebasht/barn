"use client";

import { useBarnScope } from "@/lib/barn-scope-context";
import { useI18n } from "@/lib/i18n/context";

export function ScopeHeading({ page }: { page: string }) {
  const { activeBarn, isGlobalScope, isMasterMode } = useBarnScope();
  const { t } = useI18n();
  if (!isMasterMode) return null;
  return (
    <div className="scope-heading">
      <span>{isGlobalScope ? t("master.allBarns") : activeBarn?.name || t("master.unknownBarn")}</span>
      <span aria-hidden>›</span>
      <span>{page}</span>
    </div>
  );
}
