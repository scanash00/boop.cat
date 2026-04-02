// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

import React from 'react';
import { Link, useOutletContext } from 'react-router-dom';
import MillyLogo from '../components/MillyLogo.jsx';

const PROJECT_COLORS = [
  '#e88978', '#6ba3e8', '#72d2cf', '#9b87f5', '#f5a524',
  '#e87298', '#4ecdc4', '#a78bfa', '#fb923c', '#34d399',
];

function getProjectColor(name) {
  let hash = 0;
  for (let i = 0; i < (name || '').length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return PROJECT_COLORS[Math.abs(hash) % PROJECT_COLORS.length];
}

export default function DashboardHome() {
  const { me, sites } = useOutletContext();
  const siteLimitReached = sites.length >= 10;

  return (
    <div className="page">
      <div className="pageHeader" style={{ padding: '20px 0 40px', display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between' }}>
        <div>
          <h1 style={{ fontSize: '3rem', fontWeight: 800, margin: '0 0 12px', letterSpacing: '-0.03em', lineHeight: 1 }}>
            Your <span style={{ color: 'var(--accent)' }}>Websites</span>
          </h1>
          <div className="muted" style={{ fontSize: '1.1rem' }}>Manage deployments and environment variables.</div>
        </div>
        <div className="topActions">
          {siteLimitReached ? (
            <button className="btn primary" disabled style={{ padding: '16px 28px', fontSize: '16px' }}>
              + New website
            </button>
          ) : (
            <Link to="/dashboard/new" className="btn primary" style={{ padding: '16px 28px', fontSize: '16px' }}>
              + New website
            </Link>
          )}
        </div>
      </div>

      {me && me.emailVerified === false ? (
        <div className="notice">
          <strong>Verify your email</strong> to deploy your websites. Check your inbox for the verification link.
        </div>
      ) : null}

      {siteLimitReached && (
        <div className="notice">
          <strong>Limit reached:</strong> You have 10 projects. Delete a project to create a new one.
        </div>
      )}

      <div className="bentoGrid">
        {sites.map((s, index) => {
          const color = getProjectColor(s.name);
          const letter = (s.name || '?')[0].toUpperCase();
          const isHero = index === 0 && sites.length > 0;

          return (
            <Link key={s.id} className={`projectCard ${isHero ? 'bentoHero' : 'bentoNormal'}`} to={`/dashboard/site/${s.id}`}>
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: isHero ? 24 : 16 }}>
                <div
                  style={{
                    width: isHero ? 76 : 52,
                    height: isHero ? 76 : 52,
                    borderRadius: isHero ? 18 : 12,
                    background: color,
                    color: '#fff',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 900,
                    fontSize: isHero ? 34 : 24,
                    flexShrink: 0,
                    border: '3px solid var(--card-text)',
                    boxShadow: '3px 3px 0 var(--card-text)',
                    letterSpacing: '-0.04em',
                    fontFamily: 'inherit',
                  }}
                >
                  {letter}
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div className="projectTitle" style={{ fontSize: isHero ? 24 : 18, marginBottom: isHero ? 8 : 4 }}>
                    {s.name}
                  </div>
                  <div className="muted" style={{ fontSize: isHero ? 15 : 13, wordBreak: 'break-all' }}>
                    {s.domain ? s.domain : s.git?.url?.replace('https://github.com/', '')}
                  </div>
                </div>
              </div>
              <div className="projectMeta" style={{ marginTop: 'auto', paddingTop: isHero ? 24 : 16 }}>
                <span className="chip" style={{ fontSize: isHero ? 13 : 11 }}>{s.git?.branch || 'main'}</span>
                {s.git?.subdir && <span className="chip" style={{ fontSize: isHero ? 13 : 11 }}>{s.git.subdir}</span>}
              </div>
            </Link>
          );
        })}

        {sites.length === 0 && (
          <div className="panel" style={{ gridColumn: '1 / -1', border: '3px dashed var(--card-text)', background: 'var(--card-bg-solid)', boxShadow: '6px 6px 0 var(--card-text)', opacity: 0.85 }}>
            <div style={{ textAlign: 'center', padding: '64px 24px' }}>
              <div style={{ marginBottom: 32, animation: 'float 4s ease-in-out infinite', cursor: 'default' }}>
                <MillyLogo size={96} />
              </div>
              <h3 style={{ fontSize: '2rem', fontWeight: 800, marginBottom: 12, letterSpacing: '-0.02em' }}>No websites yet</h3>
              <div
                className="muted"
                style={{ marginBottom: 32, maxWidth: 420, marginLeft: 'auto', marginRight: 'auto', lineHeight: 1.6, fontSize: '1.1rem' }}
              >
                Deploy blazing fast static sites from any Git repository. Connect your GitHub or paste a URL to get started in seconds.
              </div>
              <Link to="/dashboard/new" className="btn primary" style={{ padding: '16px 32px', fontSize: '16px' }}>
                Deploy your first site
              </Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
