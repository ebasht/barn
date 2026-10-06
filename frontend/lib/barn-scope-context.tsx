"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { api } from "@/lib/api";
import {
  BARN_SCOPE_PARAM,
  GLOBAL_BARN_SCOPE,
  normalizeBarnTargets,
  parseBarnScope,
  withBarnScope,
  type BarnScope,
  type BarnTarget,
} from "@/lib/barn-scope";
import { setRemoteApiTarget } from "@/lib/remote-api-target";
import { useServersMode } from "@/lib/servers-mode";
import type { ServerNode } from "@/lib/types";

type BarnScopeContextValue = {
  currentScope: BarnScope;
  activeBarn: BarnTarget | null;
  targets: BarnTarget[];
  nodes: ServerNode[];
  isGlobalScope: boolean;
  isMasterMode: boolean;
  loadingTargets: boolean;
  scopeHref: (href: string) => string;
  selectGlobalScope: () => void;
  selectBarn: (target: BarnTarget) => void;
  refreshTargets: () => Promise<void>;
};

const BarnScopeContext = createContext<BarnScopeContextValue | null>(null);

const BARN_SCOPE_STORAGE = "barn.activeScope";

function browserScope(): BarnScope {
  if (typeof window === "undefined") return GLOBAL_BARN_SCOPE;
  const params = new URLSearchParams(window.location.search);
  if (params.has(BARN_SCOPE_PARAM)) return parseBarnScope(params.get(BARN_SCOPE_PARAM));
  const stored = window.sessionStorage.getItem(BARN_SCOPE_STORAGE);
  return stored ? parseBarnScope(stored) : GLOBAL_BARN_SCOPE;
}

export function BarnScopeProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const barnParam = searchParams.get(BARN_SCOPE_PARAM);
  const { settings, isMaster, loading: modeLoading } = useServersMode();
  const [currentScope, setCurrentScope] = useState<BarnScope>(GLOBAL_BARN_SCOPE);
  const [nodes, setNodes] = useState<ServerNode[]>([]);
  const [loadingTargets, setLoadingTargets] = useState(false);
  const [targetsLoaded, setTargetsLoaded] = useState(false);

  const refreshTargets = useCallback(async () => {
    if (!isMaster) {
      setNodes([]);
      setTargetsLoaded(true);
      return;
    }
    setLoadingTargets(true);
    try {
      setNodes(await api.listServerNodes());
    } catch {
      setNodes([]);
    } finally {
      setLoadingTargets(false);
      setTargetsLoaded(true);
    }
  }, [isMaster]);

  useEffect(() => {
    const scope = browserScope();
    setCurrentScope(scope);
    const params = new URLSearchParams(window.location.search);
    if (scope.type === "global" && params.has(BARN_SCOPE_PARAM)) {
      window.sessionStorage.removeItem(BARN_SCOPE_STORAGE);
    }
    if (scope.type === "node" && params.has(BARN_SCOPE_PARAM)) {
      window.sessionStorage.setItem(BARN_SCOPE_STORAGE, scope.nodeId);
    }
    if (scope.type === "node" && !params.has(BARN_SCOPE_PARAM)) {
      router.replace(withBarnScope(`${pathname || "/"}${window.location.search}`, scope));
    }
  }, [barnParam, pathname, router]);

  useEffect(() => {
    if (modeLoading) return;
    void refreshTargets();
  }, [modeLoading, refreshTargets]);

  const targets = useMemo(
    () => normalizeBarnTargets(settings, nodes),
    [settings, nodes],
  );
  const activeBarn = useMemo(
    () =>
      currentScope.type === "node"
        ? targets.find((target) => target.id === currentScope.nodeId) ?? null
        : null,
    [currentScope, targets],
  );

  useEffect(() => {
    if (isMaster && activeBarn?.kind === "barn") {
      setRemoteApiTarget(activeBarn.id);
    } else {
      setRemoteApiTarget(null);
    }
  }, [activeBarn, isMaster]);

  useEffect(() => {
    if (
      isMaster &&
      !loadingTargets &&
      currentScope.type === "node" &&
      targetsLoaded &&
      !activeBarn
    ) {
      setCurrentScope(GLOBAL_BARN_SCOPE);
      window.sessionStorage.removeItem(BARN_SCOPE_STORAGE);
      setRemoteApiTarget(null);
      router.replace(withBarnScope(pathname || "/overview", GLOBAL_BARN_SCOPE));
    }
  }, [activeBarn, currentScope, isMaster, loadingTargets, pathname, router, targetsLoaded]);

  const scopeHref = useCallback(
    (href: string) => isMaster ? withBarnScope(href, currentScope) : href,
    [currentScope, isMaster],
  );

  const selectGlobalScope = useCallback(() => {
    setCurrentScope(GLOBAL_BARN_SCOPE);
    window.sessionStorage.removeItem(BARN_SCOPE_STORAGE);
    setRemoteApiTarget(null);
    router.push(withBarnScope("/overview", GLOBAL_BARN_SCOPE));
  }, [router]);

  const selectBarn = useCallback((target: BarnTarget) => {
    const scope: BarnScope = { type: "node", nodeId: target.id };
    setCurrentScope(scope);
    window.sessionStorage.setItem(BARN_SCOPE_STORAGE, target.id);
    if (target.kind === "barn") {
      setRemoteApiTarget(target.id);
    } else {
      setRemoteApiTarget(null);
    }
    const home =
      target.kind === "local" || target.kind === "barn"
        ? "/sites"
        : `/servers/${target.id}`;
    router.push(withBarnScope(home, scope));
  }, [router]);

  const value = useMemo<BarnScopeContextValue>(
    () => ({
      currentScope: isMaster ? currentScope : GLOBAL_BARN_SCOPE,
      activeBarn: isMaster ? activeBarn : null,
      targets,
      nodes,
      isGlobalScope: !isMaster || currentScope.type === "global",
      isMasterMode: isMaster,
      loadingTargets,
      scopeHref,
      selectGlobalScope,
      selectBarn,
      refreshTargets,
    }),
    [activeBarn, currentScope, isMaster, loadingTargets, nodes, refreshTargets, scopeHref, selectBarn, selectGlobalScope, targets],
  );

  return <BarnScopeContext.Provider value={value}>{children}</BarnScopeContext.Provider>;
}

export function useBarnScope(): BarnScopeContextValue {
  const value = useContext(BarnScopeContext);
  if (!value) throw new Error("useBarnScope must be used within BarnScopeProvider");
  return value;
}
