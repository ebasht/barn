import type { ServerNode, ServersSettings } from "@/lib/types";

export type BarnScope =
  | { type: "global" }
  | { type: "node"; nodeId: string };

export type BarnTargetKind = "local" | "barn" | "agent";

export type BarnTarget = {
  id: string;
  name: string;
  kind: BarnTargetKind;
  status: ServerNode["status"];
  isMaster: boolean;
  baseUrl?: string;
  capabilities: string[];
};

export const GLOBAL_BARN_SCOPE: BarnScope = { type: "global" };
export const BARN_SCOPE_PARAM = "barn";
export const GLOBAL_BARN_SCOPE_VALUE = "all";

export function parseBarnScope(value: string | null): BarnScope {
  const nodeId = value?.trim();
  return nodeId && nodeId !== GLOBAL_BARN_SCOPE_VALUE
    ? { type: "node", nodeId }
    : GLOBAL_BARN_SCOPE;
}

export function serializeBarnScope(scope: BarnScope): string {
  return scope.type === "node" ? scope.nodeId : GLOBAL_BARN_SCOPE_VALUE;
}

export function withBarnScope(href: string, scope: BarnScope): string {
  const [path, query = ""] = href.split("?", 2);
  const params = new URLSearchParams(query);
  const value = serializeBarnScope(scope);
  params.set(BARN_SCOPE_PARAM, value);
  const suffix = params.toString();
  return suffix ? `${path}?${suffix}` : path;
}

export function normalizeBarnTargets(
  settings: ServersSettings,
  nodes: ServerNode[],
): BarnTarget[] {
  return nodes
    .filter((node) => node.connection_type !== "agent")
    .map((node) => ({
      id: node.id,
      name:
        node.connection_type === "local"
          ? settings.node_name || node.name || "Master"
          : node.name,
      kind: node.connection_type === "local" ? "local" : "barn",
      status: node.status,
      isMaster: node.connection_type === "local" || node.role === "master",
      baseUrl: node.base_url || undefined,
      capabilities: node.capabilities ?? [],
    }));
}
