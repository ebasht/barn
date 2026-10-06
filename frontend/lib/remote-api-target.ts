/** When set, panel API calls (sites, DBs, …) are proxied through Master to this node. */
let remoteNodeId: string | null = null;

export function setRemoteApiTarget(nodeId: string | null): void {
  remoteNodeId = nodeId && nodeId.trim() ? nodeId.trim() : null;
}

export function getRemoteApiTarget(): string | null {
  return remoteNodeId;
}

/** Paths that must stay on Master even when a remote Barn is selected. */
export function isMasterLocalApiPath(path: string): boolean {
  const p = path.startsWith("/") ? path : `/${path}`;
  if (p.startsWith("/api/servers/nodes/") && p.includes("/proxy/")) {
    return true;
  }
  if (p.startsWith("/api/servers")) return true;
  if (p.startsWith("/api/auth")) return true;
  return false;
}

/** Rewrite `/api/foo` → `/api/servers/nodes/{id}/proxy/foo` when targeting a remote Barn. */
export function resolveRemoteApiPath(path: string): string {
  if (!remoteNodeId || isMasterLocalApiPath(path)) return path;
  if (!path.startsWith("/api/")) return path;
  const rest = path.slice("/api/".length);
  const qIndex = rest.indexOf("?");
  if (qIndex === -1) {
    return `/api/servers/nodes/${remoteNodeId}/proxy/${rest}`;
  }
  const pathname = rest.slice(0, qIndex);
  const query = rest.slice(qIndex);
  return `/api/servers/nodes/${remoteNodeId}/proxy/${pathname}${query}`;
}
