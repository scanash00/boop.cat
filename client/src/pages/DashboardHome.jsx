// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

import React from 'react';
import { Link, useOutletContext } from 'react-router-dom';
import MillyLogo from '../components/MillyLogo.jsx';

export default function DashboardHome() {
  const { me, sites } = useOutletContext();
  const siteLimitReached = sites.length >= 10;

  return (
    <div className="page">
      <div className="pageHeader">
        <div>
          <div className="h">Your Websites</div>
          <div className="muted">Manage deployments and environment variables.</div>
        </div>
        <div className="topActions">
          {siteLimitReached ? (
            <button className="btn primary" disabled>
              + New website
            </button>
          ) : (
            <Link to="/dashboard/new" className="btn primary">
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

      <div className="gridCards">
        {sites.map((s) => (
          <Link key={s.id} className="projectCard" to={`/dashboard/site/${s.id}`}>
            <div className="projectTitle">{s.name}</div>
            <div className="muted" style={{ marginTop: 8, fontSize: 13 }}>
              {s.domain ? s.domain : s.git?.url?.replace('https://github.com/', '')}
            </div>
            <div className="projectMeta">
              <span className="chip">{s.git?.branch || 'main'}</span>
              {s.git?.subdir && <span className="chip">{s.git.subdir}</span>}
            </div>
          </Link>
        ))}

        {sites.length === 0 && (
          <div className="panel" style={{ gridColumn: '1 / -1' }}>
            <div style={{ textAlign: 'center', padding: '32px 24px' }}>
              <MillyLogo size={64} style={{ marginBottom: 20 }} />
              <h3 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: 8 }}>Create your first website</h3>
              <div
                className="muted"
                style={{ marginBottom: 24, maxWidth: 320, marginLeft: 'auto', marginRight: 'auto' }}
              >
                Deploy static sites from any Git repository in seconds.
              </div>
              <Link to="/dashboard/new" className="btn primary">
                + New website
              </Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
