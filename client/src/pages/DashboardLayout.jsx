// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

import React, { useEffect, useMemo, useState } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import {
  LayoutDashboard,
  Settings,
  LogOut,
  Zap,
  Globe,
  AlertTriangle,
  User,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Plus,
  Menu,
  X,
  Book
} from 'lucide-react';
import md5 from 'blueimp-md5';
import ThemeToggle from '../components/ThemeToggle.jsx';
import MillyLogo from '../components/MillyLogo.jsx';

function useApi() {
  return async (path, init) => {
    const res = await fetch(path, {
      ...init,
      credentials: 'same-origin',
      headers: {
        'content-type': 'application/json',
        ...(init?.headers || {})
      }
    });
    const ct = res.headers.get('content-type') || '';
    const data = ct.includes('application/json') ? await res.json().catch(() => null) : await res.text();
    if (!res.ok) {
      const msg = typeof data === 'string' ? data : data?.error || res.statusText;
      throw new Error(msg);
    }
    return data;
  };
}

function getGravatarUrl(email, size = 40) {
  if (!email) return null;
  const hash = md5(email.trim().toLowerCase());
  return `https://www.gravatar.com/avatar/${hash}?s=${size}&d=mp`;
}

function getAvatar(user, size = 40) {
  if (user?.avatarUrl) return user.avatarUrl;
  return getGravatarUrl(user?.email, size);
}

function UserDropdown({ user, onLogout }) {
  const [open, setOpen] = useState(false);
  const avatar = getAvatar(user);
  const displayName = user?.username || user?.email || 'User';

  return (
    <div className="userDropdown">
      <button className="userDropdownTrigger" onClick={() => setOpen(!open)}>
        {avatar ? (
          <img src={avatar} alt="" className="avatar" />
        ) : (
          <div className="avatarPlaceholder">
            <User size={20} />
          </div>
        )}
        <div className="userInfo">
          <span className="userName">{displayName}</span>
          <span className="userEmail">{user?.email}</span>
        </div>
        <ChevronDown size={16} className={`chevron ${open ? 'open' : ''}`} />
      </button>

      {open && (
        <>
          <div className="dropdownOverlay" onClick={() => setOpen(false)} />
          <div className="dropdownMenu">
            <Link className="dropdownItem" to="/dashboard/account" onClick={() => setOpen(false)}>
              <Settings size={16} />
              Account Settings
            </Link>
            <button className="dropdownItem" onClick={onLogout}>
              <LogOut size={16} />
              Logout
            </button>
          </div>
        </>
      )}
    </div>
  );
}

function Sidebar({ sites, selectedId, collapsed, onToggle, user, onLogout, mobileOpen, onMobileClose }) {
  const avatar = getAvatar(user, 64);
  const displayName = user?.username || user?.email || 'User';
  const location = useLocation();

  const isMobile = typeof window !== 'undefined' && window.innerWidth <= 768;
  const isCollapsed = isMobile ? false : collapsed;

  useEffect(() => {
    if (mobileOpen) {
      onMobileClose();
    }
  }, [location.pathname]);

  return (
    <>
      {}
      <div className={`sidebarOverlay ${mobileOpen ? 'visible' : ''}`} onClick={onMobileClose} />

      <aside className={`sidebar ${isCollapsed ? 'collapsed' : ''} ${mobileOpen ? 'mobileOpen' : ''}`}>
        <div className="sidebarHeader">
          <Link to="/" className="brand">
            <MillyLogo size={24} />
            {!isCollapsed && <span>boop.cat</span>}
          </Link>
        </div>

        <nav className="sidebarNav">
          <Link className="sidebarLink" to="/dashboard" title="All websites">
            <LayoutDashboard size={20} />
            {!isCollapsed && <span>All websites</span>}
          </Link>
          <Link className="sidebarLink" to="/dashboard/account" title="Settings">
            <Settings size={20} />
            {!isCollapsed && <span>Settings</span>}
          </Link>
          <Link className="sidebarLink" to="/dashboard/api-docs" title="API Docs">
            <Book size={20} />
            {!isCollapsed && <span>API Docs</span>}
          </Link>
        </nav>

        {!isCollapsed && sites.length > 0 && (
          <div className="sidebarSection">
            <div className="sidebarLabel">Projects</div>
            <div className="sidebarProjects">
              {sites.map((s) => (
                <Link
                  key={s.id}
                  className={`sidebarProject ${s.id === selectedId ? 'active' : ''}`}
                  to={`/dashboard/site/${s.id}`}
                >
                  <Globe size={16} />
                  <div className="sidebarProjectInfo">
                    <span className="sidebarProjectName">{s.name}</span>
                    <span className="sidebarProjectUrl">
                      {s.domain || s.git?.url?.replace('https://github.com/', '') || ''}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        )}

        {isCollapsed && sites.length > 0 && (
          <div className="sidebarSection">
            {sites.map((s) => (
              <Link
                key={s.id}
                className={`sidebarLink ${s.id === selectedId ? 'active' : ''}`}
                to={`/dashboard/site/${s.id}`}
                title={s.name}
              >
                <Globe size={20} />
              </Link>
            ))}
          </div>
        )}

        <div className="sidebarUser">
          <div className="sidebarUserInfo">
            {avatar ? (
              <img src={avatar} alt="" className="sidebarAvatar" />
            ) : (
              <div className="sidebarAvatarPlaceholder">
                <User size={18} />
              </div>
            )}
            {!isCollapsed && (
              <div className="sidebarUserText">
                <span className="sidebarUserName">{displayName}</span>
                <span className="sidebarUserEmail">{user?.email}</span>
              </div>
            )}
          </div>
          <button className="sidebarLogout" onClick={onLogout} title="Logout">
            <LogOut size={18} />
          </button>
        </div>

        <div className="sidebarFooter">
          <button className="sidebarToggle" onClick={onToggle} title={isCollapsed ? 'Expand' : 'Collapse'}>
            {isCollapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
            {!isCollapsed && <span>Collapse</span>}
          </button>
        </div>
      </aside>
    </>
  );
}

export default function DashboardLayout() {
  const api = useApi();
  const nav = useNavigate();
  const loc = useLocation();

  const [authChecked, setAuthChecked] = useState(false);
  const [me, setMe] = useState(null);
  const [sites, setSites] = useState([]);
  const [error, setError] = useState('');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
    const saved = localStorage.getItem('sidebarCollapsed');
    return saved === 'true';
  });
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const toggleSidebar = () => {
    setSidebarCollapsed((prev) => {
      localStorage.setItem('sidebarCollapsed', String(!prev));
      return !prev;
    });
  };

  const toggleMobileMenu = () => {
    setMobileMenuOpen((prev) => !prev);
  };

  const closeMobileMenu = () => {
    setMobileMenuOpen(false);
  };

  const selectedId = useMemo(() => {
    const m = loc.pathname.match(/\/dashboard\/site\/([^/]+)/);
    return m ? m[1] : null;
  }, [loc.pathname]);

  async function refreshSites() {
    const data = await api('/api/sites');
    const arr = Array.isArray(data) ? data : [];
    setSites(arr);
    return arr;
  }

  useEffect(() => {
    api('/api/auth/me')
      .then((d) => {
        if (!d?.authenticated) {
          nav('/login');
          return;
        }
        setMe(d.user);
        setAuthChecked(true);
      })
      .catch(() => {
        nav('/login');
      });
  }, []);

  useEffect(() => {
    if (!authChecked) return;
    refreshSites().catch((e) => setError(e.message));
  }, [authChecked]);

  async function logout() {
    await api('/api/auth/logout', { method: 'POST' }).catch(() => {});
    nav('/');
  }

  if (!authChecked || !me) {
    return (
      <div className="auth-page">
        <div className="auth-card" style={{ textAlign: 'center' }}>
          <MillyLogo size={48} style={{ marginBottom: 16 }} />
          <h1 style={{ margin: '0 0 8px', fontSize: '1.5rem' }}>Loading Dashboard</h1>
          <p className="muted">Just a moment...</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`dashWrapper ${sidebarCollapsed ? 'sidebarCollapsed' : ''}`}>
      <Sidebar
        sites={sites}
        selectedId={selectedId}
        collapsed={sidebarCollapsed}
        onToggle={toggleSidebar}
        user={me}
        onLogout={logout}
        mobileOpen={mobileMenuOpen}
        onMobileClose={closeMobileMenu}
      />
      <main className="dashMain">
        <div className="dashHeader">
          <button className="mobileMenuBtn" onClick={toggleMobileMenu} aria-label="Toggle menu">
            {mobileMenuOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
          <ThemeToggle />
        </div>
        {error ? <div className="errorBox">{error}</div> : null}
        <Outlet
          context={{
            api,
            me,
            sites,
            refreshSites,
            setError
          }}
        />
      </main>
    </div>
  );
}
