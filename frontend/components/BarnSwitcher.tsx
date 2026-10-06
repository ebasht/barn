"use client";

import Link from "next/link";
import { useState } from "react";
import { useBarnScope } from "@/lib/barn-scope-context";
import { useI18n } from "@/lib/i18n/context";

export function BarnSwitcher({ onSelect }: { onSelect?: () => void }) {
  const { t } = useI18n();
  const { activeBarn, isGlobalScope, loadingTargets, targets, selectBarn, selectGlobalScope } = useBarnScope();
  const [open, setOpen] = useState(false);
  const choose = (action: () => void) => { action(); setOpen(false); onSelect?.(); };

  return (
    <div className="barn-switcher">
      <button type="button" className="barn-switcher-trigger" aria-expanded={open} onClick={() => setOpen((value) => !value)}>
        <span><small>{t("master.identity")}</small><strong>{isGlobalScope ? t("master.allBarns") : activeBarn?.name || t("master.unknownBarn")}</strong></span>
        <span aria-hidden>⌄</span>
      </button>
      {open && (
        <div className="barn-switcher-menu">
          <button type="button" className={`barn-switcher-option${isGlobalScope ? " active" : ""}`} onClick={() => choose(selectGlobalScope)}>
            <span className="barn-target-icon">⌂</span><span><strong>{t("master.allBarns")}</strong><small>{t("master.globalScope")}</small></span>
          </button>
          {loadingTargets && <span className="barn-switcher-loading">{t("common.loading")}</span>}
          {targets.map((target) => (
            <button type="button" key={target.id} className={`barn-switcher-option${activeBarn?.id === target.id ? " active" : ""}`} onClick={() => choose(() => selectBarn(target))}>
              <span className={`barn-target-status barn-target-status-${target.status}`} aria-hidden />
              <span><strong>{target.name}</strong><small>{target.isMaster ? t("master.thisMaster") : t(`master.targetKind.${target.kind}`)}</small></span>
              {target.isMaster && <b>MASTER</b>}
            </button>
          ))}
          <Link href="/servers/new?barn=all" className="barn-switcher-add" onClick={() => { setOpen(false); onSelect?.(); }}>+ {t("master.connectServer")}</Link>
        </div>
      )}
    </div>
  );
}
