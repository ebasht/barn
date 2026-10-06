import type { BarnTarget } from "@/lib/barn-scope";

export type BarnNavIcon =
  | "overview"
  | "projects"
  | "servers"
  | "database"
  | "backup"
  | "payments"
  | "notifications"
  | "settings"
  | "events";

export type BarnNavItem = {
  href: string;
  icon: BarnNavIcon;
  labelKey: string;
  match: string;
  external?: boolean;
};

const LOCAL_NAV: BarnNavItem[] = [
  { href: "/sites", icon: "projects", labelKey: "nav.sites", match: "sites" },
  { href: "/databases", icon: "database", labelKey: "nav.databases", match: "databases" },
  { href: "/backups", icon: "backup", labelKey: "nav.backups", match: "backups" },
  { href: "/payments", icon: "payments", labelKey: "nav.payments", match: "payments" },
  { href: "/notifications", icon: "notifications", labelKey: "nav.notifications", match: "notifications" },
  { href: "/servers/settings", icon: "settings", labelKey: "nav.serversSettings", match: "servers-settings" },
];

const GLOBAL_NAV: BarnNavItem[] = [
  { href: "/overview", icon: "overview", labelKey: "nav.overview", match: "overview" },
  { href: "/sites", icon: "projects", labelKey: "nav.sites", match: "sites" },
  { href: "/servers", icon: "servers", labelKey: "nav.servers", match: "servers" },
  { href: "/servers/events", icon: "events", labelKey: "nav.serverEvents", match: "servers-events" },
  { href: "/servers/settings", icon: "settings", labelKey: "nav.serversSettings", match: "servers-settings" },
];

const SERVERS_NAV: BarnNavItem = {
  href: "/servers",
  icon: "servers",
  labelKey: "nav.servers",
  match: "servers",
};

export function resolveBarnNavigation(
  isMaster: boolean,
  target: BarnTarget | null,
): BarnNavItem[] {
  if (!isMaster) return LOCAL_NAV;
  if (!target) return GLOBAL_NAV;
  if (target.kind === "local") {
    return [LOCAL_NAV[0], SERVERS_NAV, ...LOCAL_NAV.slice(1)];
  }

  if (target.kind === "barn") {
    // Remote panel via Master→node proxy: panel pages only.
    // Servers list is a Master control-plane view (All Barns / local Master).
    return LOCAL_NAV.filter((item) => item.match !== "servers-settings");
  }

  // Monitoring agent: node overview only (no fleet Servers list).
  return [
    {
      href: `/servers/${target.id}`,
      icon: "overview",
      labelKey: "nav.overview",
      match: `server-${target.id}`,
    },
  ];
}
