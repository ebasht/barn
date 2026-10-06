"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { BarnSwitcher } from "@/components/BarnSwitcher";
import { BrandLogo } from "@/components/BrandLogo";
import { useLogout } from "@/components/AuthGate";
import { resolveBarnNavigation, type BarnNavIcon } from "@/lib/barn-navigation";
import { useBarnScope } from "@/lib/barn-scope-context";
import { useI18n } from "@/lib/i18n/context";

function NavIcon({ name }: { name: BarnNavIcon }) {
  const paths: Record<BarnNavIcon, React.ReactNode> = {
    overview: <><rect x="4" y="4" width="6" height="6" rx="1"/><rect x="14" y="4" width="6" height="6" rx="1"/><rect x="4" y="14" width="6" height="6" rx="1"/><rect x="14" y="14" width="6" height="6" rx="1"/></>,
    projects: <><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></>,
    servers: <><rect x="3" y="4" width="18" height="6" rx="2"/><rect x="3" y="14" width="18" height="6" rx="2"/><path d="M7 7h.01M7 17h.01"/></>,
    database: <><ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v7c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 12v7c0 1.7 3.6 3 8 3s8-1.3 8-3v-7"/></>,
    backup: <><path d="M5 5h12l2 3v11H5z"/><path d="M9 5v5h6V5M9 19v-5h6v5"/></>,
    payments: <><rect x="3" y="5" width="18" height="14" rx="2"/><path d="M3 10h18M7 15h3"/></>,
    notifications: <><path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4"/></>,
    settings: <><circle cx="12" cy="12" r="3"/><path d="M19 12a7 7 0 0 0-.1-1l2-1.5-2-3.5-2.4 1A7 7 0 0 0 15 6l-.3-2.5h-4L10.4 6a7 7 0 0 0-1.5 1L6.5 6 4.5 9.5l2 1.5a7 7 0 0 0 0 2l-2 1.5 2 3.5 2.4-1a7 7 0 0 0 1.5 1l.3 2.5h4L15 18a7 7 0 0 0 1.5-1l2.4 1 2-3.5-2-1.5a7 7 0 0 0 .1-1z"/></>,
    events: <><path d="M4 19V5M4 19h16"/><path d="m7 15 4-4 3 2 5-6"/></>,
  };
  return <svg className="nav-icon" viewBox="0 0 24 24" aria-hidden="true">{paths[name]}</svg>;
}

function isActive(pathname: string, match: string): boolean {
  if (match.startsWith("server-")) return pathname === `/servers/${match.slice(7)}`;
  if (match === "servers") {
    return pathname === "/servers" || pathname.startsWith("/servers/new");
  }
  if (match === "servers-events") return pathname.startsWith("/servers/events");
  if (match === "servers-settings") return pathname.startsWith("/servers/settings");
  return pathname === `/${match}` || pathname.startsWith(`/${match}/`);
}

export function Nav() {
  const logout = useLogout();
  const { t } = useI18n();
  const { activeBarn, isMasterMode, scopeHref } = useBarnScope();
  const pathname = usePathname() || "";
  const [menuOpen, setMenuOpen] = useState(false);
  const links = useMemo(() => resolveBarnNavigation(isMasterMode, activeBarn), [activeBarn, isMasterMode]);

  useEffect(() => setMenuOpen(false), [pathname]);
  useEffect(() => {
    if (!menuOpen) return;
    const onKey = (event: KeyboardEvent) => event.key === "Escape" && setMenuOpen(false);
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = previous;
      window.removeEventListener("keydown", onKey);
    };
  }, [menuOpen]);

  const close = () => setMenuOpen(false);
  const renderedLinks = (mobile: boolean) => (
    <div className={`nav-primary ${mobile ? "nav-primary-mobile" : "nav-primary-desktop"}`}>
      {links.map((item) => {
        const active = isActive(pathname, item.match);
        const className = `nav-link${active ? " nav-link-active" : ""}`;
        const content = <><NavIcon name={item.icon} />{t(item.labelKey)}</>;
        return item.external ? (
          <a key={item.href} href={item.href} className={className} target="_blank" rel="noopener noreferrer" onClick={close}>{content}</a>
        ) : (
          <Link key={`${item.href}-${item.match}`} href={isMasterMode && item.match === "servers" ? "/servers?barn=all" : scopeHref(item.href)} className={className} aria-current={active ? "page" : undefined} onClick={close}>{content}</Link>
        );
      })}
    </div>
  );

  return (
    <nav className={`nav${menuOpen ? " nav-menu-open" : ""}${isMasterMode ? " nav-master" : ""}`}>
      <div className="nav-bar">
        <Link href={isMasterMode ? "/overview?barn=all" : "/sites"} className="nav-brand" onClick={close}>
          <BrandLogo showVersion showServerIP />
        </Link>
        <button type="button" className="nav-toggle btn btn-secondary" aria-expanded={menuOpen} aria-controls="nav-menu" onClick={() => setMenuOpen((open) => !open)}>{t("nav.menu")}</button>
      </div>
      {menuOpen && <button type="button" className="nav-backdrop" aria-label={t("nav.closeMenu")} onClick={close} />}
      <div id="nav-menu" className={`nav-links${menuOpen ? " nav-links-open" : ""}`}>
        <div className="nav-drawer-head">
          <span className="nav-drawer-title">{t("nav.menu")}</span>
          <button type="button" className="btn btn-secondary nav-drawer-close" onClick={close}>{t("nav.closeMenu")}</button>
        </div>
        {isMasterMode && <BarnSwitcher onSelect={close} />}
        {renderedLinks(false)}
        {renderedLinks(true)}
        <div className="nav-actions">
          <button type="button" className="btn btn-secondary nav-logout" onClick={() => { close(); logout(); }}>{t("nav.logout")}</button>
        </div>
      </div>
    </nav>
  );
}
