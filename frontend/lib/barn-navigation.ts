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

export function resolveBarnNavigation(
  isMaster: boolean,
  target: BarnTarget | null,
): BarnNavItem[] {
  if (!isMaster) return LOCAL_NAV;
  if (!target) return GLOBAL_NAV;
  if (target.kind === "local") {
    return [
      LOCAL_NAV[0],
      { href: "/servers", icon: "servers", labelKey: "nav.servers", match: "servers" },
      ...LOCAL_NAV.slice(1),
    ];
  }

  const items: BarnNavItem[] = [
    {
      href: `/servers/${target.id}`,
      icon: "overview",
      labelKey: "nav.overview",
      match: `server-${target.id}`,
    },
    { href: "/servers", icon: "servers", labelKey: "nav.servers", match: "servers" },
  ];
  if (target.kind === "barn" && target.baseUrl) {
    items.splice(1, 0, {
      href: `${target.baseUrl.replace(/\/$/, "")}/sites`,
      icon: "projects",
      labelKey: "nav.openRemoteBarn",
      match: "remote-sites",
      external: true,
    });
  }
  return items;
}
