// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useOutletContext, useParams } from 'react-router-dom';
import {
  AlertTriangle,
  FolderX,
  Zap,
  Server,
  Plug,
  AlertCircle,
  HelpCircle,
  FileQuestion,
  Mail,
  Ban,
  X,
  Lightbulb,
  FileText,
  ExternalLink,
  Check,
  Edit,
  Eye,
  EyeOff,
  Trash2,
  Plus,
  Loader2
} from 'lucide-react';

function Toast({ message, onClose }) {
  useEffect(() => {
    const timer = setTimeout(onClose, 3000);
    return () => clearTimeout(timer);
  }, [onClose]);

  if (!message) return null;

  return (
    <div className="toast">
      <Check size={16} />
      {message}
    </div>
  );
}

const ERROR_MESSAGES = {
  'site-not-found': {
    title: 'Site Not Found',
    message: "The site you're looking for doesn't exist or you don't have access to it.",
    Icon: FileQuestion
  },
  DIRECTORY_NOT_FOUND: {
    title: 'Directory Not Found',
    message: "The specified directory doesn't exist in the repository.",
    Icon: FolderX
  },
  NEXT_JS_DETECTED: {
    title: 'Dynamic Framework Detected',
    message: 'This appears to be a Next.js project which requires a Node.js server.',
    Icon: Zap
  },
  NUXT_DETECTED: {
    title: 'Dynamic Framework Detected',
    message: 'This appears to be a Nuxt.js project which requires a Node.js server.',
    Icon: Zap
  },
  REMIX_DETECTED: {
    title: 'Dynamic Framework Detected',
    message: 'This appears to be a Remix project which requires a Node.js server.',
    Icon: Zap
  },
  SERVER_FRAMEWORK_DETECTED: {
    title: 'Server Application Detected',
    message: 'This appears to be a server application, not a static site.',
    Icon: Server
  },
  API_ROUTES_DETECTED: {
    title: 'API Routes Detected',
    message: 'This project contains API routes which require server-side execution.',
    Icon: Plug
  },
  DYNAMIC_CONTENT_DETECTED: {
    title: 'Dynamic Content Detected',
    message: "This project requires server-side features that aren't supported.",
    Icon: AlertTriangle
  },
  UNKNOWN_PROJECT_TYPE: {
    title: 'Unknown Project Type',
    message: 'Could not detect a supported static site framework.',
    Icon: HelpCircle
  },
  'not-a-static-site': {
    title: 'Not a Static Site',
    message: "This project doesn't appear to be a static website.",
    Icon: FileQuestion
  },
  'dynamic-site-detected': {
    title: 'Dynamic Site Detected',
    message: 'This project requires server-side features.',
    Icon: Zap
  },
  'email-not-verified': {
    title: 'Email Not Verified',
    message: 'Please verify your email address before deploying.',
    Icon: Mail
  },
  'too-many-requests': {
    title: 'Too Many Requests',
    message: "You're making too many requests. Please wait a moment.",
    Icon: Ban
  }
};

function EnvVarModal({ open, onClose, onSave, initialKey, initialValue, isEdit }) {
  const [key, setKey] = React.useState(initialKey || '');
  const [value, setValue] = React.useState(initialValue || '');

  React.useEffect(() => {
    if (open) {
      setKey(initialKey || '');
      setValue(initialValue || '');
    }
  }, [open, initialKey, initialValue]);

  if (!open) return null;

  const handleSave = () => {
    if (!key.trim()) return;
    onSave(key.trim(), value);
    setKey('');
    setValue('');
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && e.metaKey) {
      handleSave();
    }
  };

  return (
    <div className="modalOverlay" onClick={onClose}>
      <div className="modal envModal" onClick={(e) => e.stopPropagation()}>
        <div className="modalHeader">
          <div className="modalTitle">
            {isEdit ? <Edit size={18} /> : <Plus size={18} />}
            {isEdit ? 'Edit Variable' : 'Add Variable'}
          </div>
          <button className="iconBtn" onClick={onClose}>
            <X size={20} />
          </button>
        </div>
        <div className="modalBody" onKeyDown={handleKeyDown}>
          <div className="field">
            <div className="label">Name</div>
            <input
              className="input envKeyInput"
              placeholder="MY_API_KEY"
              value={key}
              onChange={(e) => setKey(e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, ''))}
              autoFocus={!isEdit}
              disabled={isEdit}
              style={isEdit ? { opacity: 0.7, cursor: 'not-allowed' } : {}}
            />
          </div>
          <div className="field">
            <div className="label">Value</div>
            <textarea
              className="textarea envValueInput"
              placeholder="Enter value..."
              value={value}
              onChange={(e) => setValue(e.target.value)}
              rows={4}
              autoFocus={isEdit}
            />
            <div className="muted" style={{ fontSize: 11, marginTop: 6 }}>
              Press ⌘+Enter to save
            </div>
          </div>
        </div>
        <div className="modalActions">
          <button className="btn ghost" onClick={onClose}>
            Cancel
          </button>
          <button className="btn primary" onClick={handleSave} disabled={!key.trim()}>
            {isEdit ? 'Update' : 'Add'}
          </button>
        </div>
      </div>
    </div>
  );
}

function ErrorModal({ error, onClose }) {
  if (!error) return null;

  const errorInfo = ERROR_MESSAGES[error.code] ||
    ERROR_MESSAGES[error.error] || {
      title: 'Deployment Failed',
      message: error.message || 'An unexpected error occurred.',
      Icon: AlertCircle
    };

  const IconComponent = errorInfo.Icon || AlertCircle;

  return (
    <div className="modalOverlay" onClick={onClose}>
      <div className="modal errorModal" onClick={(e) => e.stopPropagation()}>
        <div className="modalHeader">
          <div className="modalTitle">
            <IconComponent size={20} style={{ marginRight: 8 }} />
            {errorInfo.title}
          </div>
          <button className="iconBtn" onClick={onClose} aria-label="Close">
            <X size={18} />
          </button>
        </div>
        <div className="modalBody">
          <div className="errorContent">
            <p className="errorMessage">{error.message || errorInfo.message}</p>
            {error.suggestion && (
              <div className="errorSuggestion">
                <Lightbulb size={16} style={{ marginRight: 6, flexShrink: 0 }} />
                <div>
                  <strong>Suggestion:</strong>
                  <p style={{ margin: '4px 0 0 0' }}>{error.suggestion}</p>
                </div>
              </div>
            )}
            {error.details && (
              <details className="errorDetails">
                <summary>Technical Details</summary>
                <pre>{JSON.stringify(error.details, null, 2)}</pre>
              </details>
            )}
          </div>
          <div className="modalActions">
            <button className="btn primary" onClick={onClose}>
              Got it
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function LogsModal({ deployment, onClose }) {
  const [logs, setLogs] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!deployment) return;

    let cancelled = false;
    let timer;
    let lastText = null;

    const isTerminal = (status) => {
      const s = String(status || '').toLowerCase();
      return s === 'active' || s === 'ready' || s === 'failed' || s === 'stopped' || s === 'canceled';
    };

    const load = async ({ initial } = {}) => {
      try {
        if (!cancelled && initial) {
          setLoading(true);
          setError('');
        }
        const res = await fetch(`/api/deployments/${deployment.id}/logs`, {
          credentials: 'same-origin'
        });
        const text = await res.text().catch(() => '');
        if (!res.ok) throw new Error(text || 'Failed to fetch logs');
        if (!cancelled) {
          if (text !== lastText) {
            setLogs(text);
            lastText = text;
          }
          if (initial) setLoading(false);
        }
      } catch (e) {
        if (!cancelled) {
          setError(e.message);
          if (initial) setLoading(false);
        }
      }
    };

    load({ initial: true });
    if (!isTerminal(deployment.status)) {
      timer = window.setInterval(() => load({ initial: false }), 2500);
    }

    return () => {
      cancelled = true;
      if (timer) window.clearInterval(timer);
    };
  }, [deployment?.id]);

  if (!deployment) return null;

  return (
    <div className="modalOverlay" onClick={onClose}>
      <div className="modal logsModal" onClick={(e) => e.stopPropagation()}>
        <div className="modalHeader">
          <div className="modalTitle">
            <FileText size={20} style={{ marginRight: 8 }} />
            Deployment Logs
          </div>
          <button className="iconBtn" onClick={onClose} aria-label="Close">
            <X size={18} />
          </button>
        </div>
        <div className="modalBody">
          <div className="logsInfo">
            <span className="badge">{deployment.status}</span>
            <span className="muted">{deployment.createdAt}</span>
          </div>
          {loading && <div className="logsLoading">Loading logs...</div>}
          {error && <div className="error">{error}</div>}
          {!loading && !error && <pre className="logsContent">{logs || 'No logs available'}</pre>}
        </div>
      </div>
    </div>
  );
}

export default function DashboardSite() {
  const { id } = useParams();
  const nav = useNavigate();
  const { api, me, sites, refreshSites, setError } = useOutletContext();

  const site = useMemo(() => sites.find((s) => s.id === id) || null, [sites, id]);

  const [deployments, setDeployments] = useState([]);
  const [envDraft, setEnvDraft] = useState('');
  const [settingsDraft, setSettingsDraft] = useState({
    name: '',
    gitUrl: '',
    branch: 'main',
    subdir: '',
    domain: '',
    buildCommand: '',
    outputDir: ''
  });
  const [tab, setTab] = useState('deployments');
  const [envSubTab, setEnvSubTab] = useState('styled');
  const [deployError, setDeployError] = useState(null);
  const [deploying, setDeploying] = useState(false);
  const [logsDeployment, setLogsDeployment] = useState(null);
  const [customDomains, setCustomDomains] = useState([]);
  const [customDomainInput, setCustomDomainInput] = useState('');
  const [customDomainLoading, setCustomDomainLoading] = useState(false);
  const [customDomainError, setCustomDomainError] = useState('');
  const [toast, setToast] = useState(null);
  const [visibleEnvKeys, setVisibleEnvKeys] = useState(new Set());
  const [envModalOpen, setEnvModalOpen] = useState(false);
  const [editingEnv, setEditingEnv] = useState(null); // { key, value } for edit mode
  const [pollingDomains, setPollingDomains] = useState(new Set());

  const repoInfo = useMemo(() => {
    const gitUrl = site?.git?.url || '';
    const match = gitUrl.match(/github\.com[/:]([^/]+)\/([^/.]+)/i);
    if (!match) return null;
    const [, owner, repo] = match;
    return { owner, repo };
  }, [site]);

  const formatTimestamp = (iso) => {
    if (!iso) return '';
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return iso;
    return new Intl.DateTimeFormat(undefined, {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      timeZoneName: 'short'
    }).format(d);
  };

  const getCommitAuthorName = (author) => {
    if (!author) return null;
    const match = author.match(/^([^<]+)</);
    return match ? match[1].trim() : author.trim();
  };

  const getCommitAvatar = (author, deployment) => {
    if (deployment?.commitAvatar) {
      return deployment.commitAvatar;
    }

    const emailMatch = author?.match(/<([^>]+)>/);
    const email = emailMatch ? emailMatch[1] : null;

    if (email && email.includes('noreply.github.com')) {
      const parts = email.split('@')[0].split('+');
      const username = parts.length > 1 ? parts[1] : parts[0];
      return `https://github.com/${username}.png?size=64`;
    }

    if (email) {
      const hash = Array.from(new TextEncoder().encode(email.trim().toLowerCase())).reduce((hash, c) => {
        hash = (hash << 5) - hash + c;
        return hash & hash;
      }, 0);
      return `https://www.gravatar.com/avatar/${Math.abs(hash)}?d=mp&s=64`;
    }

    const fallbackName = getCommitAuthorName(author) || repoInfo?.owner || 'Committer';
    return `https://api.dicebear.com/7.x/initials/svg?backgroundColor=6ba3e8,9b87f5,72d2cf&fontWeight=700&size=64&seed=${encodeURIComponent(fallbackName)}`;
  };

  const activeDeployment = useMemo(() => {
    return deployments.find((d) => d.status === 'building' || d.status === 'running') || null;
  }, [deployments]);

  const [config, setConfig] = useState({ deliveryMode: '', edgeRootDomain: '' });

  const edgeOnly = String(config?.deliveryMode || '').toLowerCase() === 'edge' && Boolean(config?.edgeRootDomain);

  function toEdgeLabel(value) {
    const v = String(value || '')
      .trim()
      .toLowerCase();
    const root = String(config?.edgeRootDomain || '')
      .trim()
      .toLowerCase();
    if (!v || !root) return v;
    if (v.endsWith(`.${root}`)) return v.slice(0, -(root.length + 1));
    return v;
  }

  useEffect(() => {
    if (!site) return;
    setEnvDraft(site.envText || '');
    setSettingsDraft({
      name: site.name || '',
      gitUrl: site.git?.url || '',
      branch: site.git?.branch || 'main',
      subdir: site.git?.subdir || '',
      domain: edgeOnly ? toEdgeLabel(site.domain || '') : site.domain || '',
      buildCommand: site.buildCommand || '',
      outputDir: site.outputDir || ''
    });
  }, [site?.id, edgeOnly, config?.edgeRootDomain]);

  async function refreshDeployments() {
    const data = await api(`/api/sites/${encodeURIComponent(id)}/deployments`);
    setDeployments(Array.isArray(data) ? data : []);
  }

  async function deleteProject() {
    if (!confirm('Delete this project and all its deployments?')) return;
    setError('');
    await api(`/api/sites/${encodeURIComponent(site.id)}`, {
      method: 'DELETE'
    }).catch((e) => {
      setError(e.message);
      throw e;
    });
    await refreshSites();
    nav('/dashboard');
  }

  useEffect(() => {
    if (!site) return;
    refreshDeployments().catch((e) => setError(e.message));

    let timer;
    if (activeDeployment) {
      timer = setInterval(() => {
        refreshDeployments().catch(() => {});
      }, 2000);
    }
    return () => {
      if (timer) clearInterval(timer);
    };
  }, [site?.id, activeDeployment?.status]);

  useEffect(() => {
    if (!site) return;
    (async () => {
      try {
        const data = await api(`/api/sites/${encodeURIComponent(site.id)}/custom-domains`);
        setCustomDomains(Array.isArray(data) ? data : []);
      } catch (e) {
        console.error('Failed to load custom domains:', e);
      }
    })();
  }, [site?.id]);

  useEffect(() => {
    api('/api/config')
      .then((d) => setConfig({ deliveryMode: d?.deliveryMode || '', edgeRootDomain: d?.edgeRootDomain || '' }))
      .catch(() => setConfig({ deliveryMode: '', edgeRootDomain: '' }));
  }, []);

  function parseEnvText(text) {
    const lines = String(text || '').split('\n');
    const parsed = lines.map((line) => {
      if (!line.trim()) return { key: '', value: '' };
      const equalsIdx = line.indexOf('=');
      if (equalsIdx === -1) return { key: line.trim(), value: '' };
      return { key: line.slice(0, equalsIdx).trim(), value: line.slice(equalsIdx + 1) };
    });
    return parsed;
  }

  function serializeEnvEntries(entries) {
    return entries
      .filter((e) => e.key.trim() || e.value.trim())
      .map((e) => `${e.key}=${e.value}`)
      .join('\n');
  }

  const envEntries = useMemo(() => {
    const parsed = parseEnvText(envDraft);
    const nonEmpty = parsed.filter((e) => e.key.trim() || e.value.trim());
    const hasBlank = parsed.some((e) => !e.key.trim() && !e.value.trim());
    if (nonEmpty.length === 0 && !hasBlank) return [{ key: '', value: '' }];
    return parsed;
  }, [envDraft]);

  const envFilledCount = useMemo(() => envEntries.filter((e) => e.key.trim()).length, [envEntries]);

  const toggleEnvVisibility = (key) => {
    setVisibleEnvKeys((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  };

  const openAddEnvModal = () => {
    setEditingEnv(null);
    setEnvModalOpen(true);
  };

  const openEditEnvModal = (key, value) => {
    setEditingEnv({ key, value });
    setEnvModalOpen(true);
  };

  const handleEnvSave = (key, value) => {
    if (editingEnv) {
      setEnvDraft((prev) => {
        const parsed = parseEnvText(prev);
        const next = parsed.map((e) => (e.key === editingEnv.key ? { key, value } : e));
        return serializeEnvEntries(next);
      });
    } else {
      setEnvDraft((prev) => {
        const parsed = parseEnvText(prev);
        const exists = parsed.some((e) => e.key === key);
        if (exists) {
          const next = parsed.map((e) => (e.key === key ? { key, value } : e));
          return serializeEnvEntries(next);
        }
        return serializeEnvEntries([...parsed, { key, value }]);
      });
    }
    setEnvModalOpen(false);
    setEditingEnv(null);
  };

  const removeEnvEntry = (key) => {
    setEnvDraft((prev) => {
      const parsed = parseEnvText(prev);
      const next = parsed.filter((e) => e.key !== key);
      return serializeEnvEntries(next);
    });
    setVisibleEnvKeys((prev) => {
      const next = new Set(prev);
      next.delete(key);
      return next;
    });
  };

  if (!site) {
    return (
      <div className="page">
        <div className="panel" style={{ textAlign: 'center', padding: '48px 24px' }}>
          <img
            src="/milly.png"
            alt=""
            width="56"
            height="56"
            style={{ imageRendering: 'pixelated', marginBottom: 20, opacity: 0.6 }}
          />
          <div className="panelTitle" style={{ fontSize: '1.25rem', marginBottom: 8 }}>
            Project not found
          </div>
          <div className="muted" style={{ marginBottom: 24 }}>
            That project doesn't exist or you don't have access.
          </div>
          <button className="btn primary" onClick={() => nav('/dashboard')}>
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  async function saveEnv() {
    setError('');
    await api(`/api/sites/${encodeURIComponent(site.id)}`, {
      method: 'PATCH',
      body: JSON.stringify({ envText: envDraft })
    }).catch((e) => {
      setError(e.message);
      throw e;
    });
    setToast('Environment variables saved.');
    await refreshSites();
  }

  async function saveSettings() {
    setError('');

    const payload = {
      ...settingsDraft,
      domain: settingsDraft.domain
    };

    await api(`/api/sites/${encodeURIComponent(site.id)}/settings`, {
      method: 'PATCH',
      body: JSON.stringify(payload)
    }).catch((e) => {
      setError(e.message);
      throw e;
    });
    setToast('Settings saved.');
    await refreshSites();
  }

  async function stopDeployment(deploymentId) {
    setError('');
    try {
      await api(`/api/deployments/${encodeURIComponent(deploymentId)}/stop`, {
        method: 'POST'
      });
      setToast('Deployment stopped.');
      await refreshDeployments();
    } catch (e) {
      setError(e.message || 'Failed to stop deployment');
    }
  }

  async function deploy() {
    setError('');
    setDeployError(null);
    setDeploying(true);

    try {
      const res = await fetch(`/api/sites/${encodeURIComponent(site.id)}/deploy`, {
        method: 'POST',
        credentials: 'same-origin',
        headers: { 'content-type': 'application/json' }
      });

      const data = await res.json().catch(() => null);

      if (!res.ok) {
        setDeployError({
          code: data?.error,
          message: data?.message || data?.error || 'Deployment failed',
          suggestion: data?.suggestion,
          details: data?.details
        });
        return;
      }

      await refreshDeployments();
    } catch (e) {
      setDeployError({
        code: 'network-error',
        message: e.message || 'Network error occurred',
        suggestion: 'Please check your internet connection and try again.'
      });
    } finally {
      setDeploying(false);
    }
  }

  async function addCustomDomain() {
    setCustomDomainError('');
    if (!customDomainInput.trim()) {
      setCustomDomainError('Enter a hostname.');
      return;
    }
    setCustomDomainLoading(true);
    try {
      const data = await api(`/api/sites/${encodeURIComponent(site.id)}/custom-domains`, {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ hostname: customDomainInput.trim() })
      });
      setCustomDomains((prev) => [...prev, data]);
      setCustomDomainInput('');
    } catch (e) {
      setCustomDomainError(e.message || 'Failed to add custom domain.');
    } finally {
      setCustomDomainLoading(false);
    }
  }

  async function pollCustomDomain(customDomainId) {
    setPollingDomains((prev) => new Set(prev).add(customDomainId));
    try {
      const data = await api(
        `/api/sites/${encodeURIComponent(site.id)}/custom-domains/${encodeURIComponent(customDomainId)}/poll`,
        { method: 'POST' }
      );
      setCustomDomains((prev) => prev.map((d) => (d.id === customDomainId ? { ...d, ...data } : d)));
      setToast('Domain status updated.');
    } catch (e) {
      console.error('Poll failed:', e);
      setError(e.message || 'Failed to check domain status.');
    } finally {
      setPollingDomains((prev) => {
        const next = new Set(prev);
        next.delete(customDomainId);
        return next;
      });
    }
  }

  async function removeCustomDomain(customDomainId) {
    if (!confirm('Remove this custom domain?')) return;
    try {
      await api(`/api/sites/${encodeURIComponent(site.id)}/custom-domains/${encodeURIComponent(customDomainId)}`, {
        method: 'DELETE'
      });
      setCustomDomains((prev) => prev.filter((d) => d.id !== customDomainId));
    } catch (e) {
      setError(e.message || 'Failed to remove custom domain.');
    }
  }

  return (
    <div className="page">
      <div className="pageHeader">
        <div>
          <div className="crumbs">
            <span className="muted">Dashboard</span>
            <span className="muted">/</span>
            <span className="crumb">{site.name}</span>
          </div>
          <div className="h">{site.name}</div>
          <div className="muted">{site.domain ? site.domain : site.git?.url}</div>
        </div>

        <div className="topActions">
          <button className="btn primary" disabled={deploying || (me && me.emailVerified === false)} onClick={deploy}>
            {deploying ? 'Deploying...' : 'Deploy'}
          </button>

          <button
            className="btn ghost"
            disabled={!activeDeployment || (me && me.emailVerified === false)}
            onClick={() => activeDeployment && stopDeployment(activeDeployment.id)}
            style={{ marginLeft: 10 }}
          >
            Stop
          </button>
        </div>
      </div>

      {me && me.emailVerified === false ? <div className="notice">Verify your email to deploy.</div> : null}

      <ErrorModal error={deployError} onClose={() => setDeployError(null)} />
      <LogsModal deployment={logsDeployment} onClose={() => setLogsDeployment(null)} />
      <EnvVarModal
        open={envModalOpen}
        onClose={() => {
          setEnvModalOpen(false);
          setEditingEnv(null);
        }}
        onSave={handleEnvSave}
        initialKey={editingEnv?.key || ''}
        initialValue={editingEnv?.value || ''}
        isEdit={Boolean(editingEnv)}
      />
      {toast && <Toast message={toast} onClose={() => setToast(null)} />}

      <div className="panel">
        <div className="tabs">
          <button className={tab === 'deployments' ? 'tab active' : 'tab'} onClick={() => setTab('deployments')}>
            Deployments
          </button>
          <button className={tab === 'env' ? 'tab active' : 'tab'} onClick={() => setTab('env')}>
            Env
          </button>
          <button className={tab === 'settings' ? 'tab active' : 'tab'} onClick={() => setTab('settings')}>
            Settings
          </button>
          {edgeOnly ? (
            <button className={tab === 'domains' ? 'tab active' : 'tab'} onClick={() => setTab('domains')}>
              Domains
            </button>
          ) : null}
        </div>

        {tab === 'deployments' ? (
          <>
            <div className="panelTitle" style={{ marginTop: 12 }}>
              Deployments
            </div>
            <div className="muted" style={{ marginBottom: 10 }}>
              Deploy + logs + stop
            </div>
            <div className="deployList">
              {deployments.map((d) => (
                <div key={d.id} className="deployRow">
                  <div className="deployRowMain">
                    <div className="deployRowTop">
                      <span className={`statusPill status-${d.status}`}>
                        {d.status === 'active' ? 'running' : d.status}
                      </span>
                      <span className="muted deployTime">{formatTimestamp(d.createdAt)}</span>
                    </div>

                    <div className="deployLinks">
                      {d.url ? (
                        <a href={d.url} target="_blank" rel="noopener noreferrer" className="muted linkChip">
                          {d.url.replace(/^https?:\/\//, '')}{' '}
                          <ExternalLink size={12} style={{ marginLeft: 4, verticalAlign: 'middle' }} />
                        </a>
                      ) : null}

                      {}
                      {site.domain && (!d.url || !d.url.includes(site.domain)) && (
                        <a
                          href={`https://${site.domain}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="muted linkChip"
                        >
                          {site.domain} <ExternalLink size={12} style={{ marginLeft: 4, verticalAlign: 'middle' }} />
                        </a>
                      )}

                      {(d.status === 'running' || d.status === 'active') &&
                        customDomains
                          .filter((cd) => cd.status === 'active')
                          .filter((cd) => !d.url || !d.url.includes(cd.hostname))
                          .map((cd) => (
                            <a
                              key={cd.id}
                              href={`https://${cd.hostname}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="muted linkChip"
                            >
                              {cd.hostname}{' '}
                              <ExternalLink size={12} style={{ marginLeft: 4, verticalAlign: 'middle' }} />
                            </a>
                          ))}
                    </div>

                    {(d.commitSha || d.commitMessage || d.commitAuthor) && (
                      <div className="deployCommit">
                        {d.commitAuthor && (
                          <div className="commitAuthorInfo">
                            <img
                              className="commitAvatar"
                              src={getCommitAvatar(d.commitAuthor, d)}
                              alt={getCommitAuthorName(d.commitAuthor) || 'Committer'}
                              loading="lazy"
                            />
                            <span className="commitAuthorName">
                              {getCommitAuthorName(d.commitAuthor) || d.commitAuthor}
                            </span>
                          </div>
                        )}
                        {d.commitMessage && (
                          <span className="commitMessage" title={d.commitMessage}>
                            {d.commitMessage}
                          </span>
                        )}
                        {(() => {
                          const shortSha = d.commitSha ? d.commitSha.slice(0, 7) : 'unknown';
                          const commitUrl =
                            repoInfo?.owner && repoInfo?.repo && d.commitSha
                              ? `https://github.com/${repoInfo.owner}/${repoInfo.repo}/commit/${d.commitSha}`
                              : null;
                          return commitUrl ? (
                            <a className="commitBadge" href={commitUrl} target="_blank" rel="noopener noreferrer">
                              {shortSha}
                            </a>
                          ) : (
                            <span className="commitBadge">{shortSha}</span>
                          );
                        })()}
                      </div>
                    )}
                  </div>

                  <div className="actions">
                    <button className="btn ghost" onClick={() => setLogsDeployment(d)}>
                      <FileText size={14} style={{ marginRight: 6 }} />
                      Logs
                    </button>
                  </div>
                </div>
              ))}
              {deployments.length === 0 ? <div className="muted">No deployments yet.</div> : null}
            </div>
          </>
        ) : null}

        {tab === 'env' ? (
          <>
            <div className="envHeader">
              <div className="envHeaderLeft">
                <div className="panelTitle" style={{ margin: 0 }}>
                  Environment Variables
                </div>
              </div>
              <div className="envHeaderRight">
                <button className="btn ghost" onClick={() => setEnvSubTab(envSubTab === 'styled' ? 'raw' : 'styled')}>
                  {envSubTab === 'styled' ? 'Raw Editor' : 'Visual Editor'}
                </button>
              </div>
            </div>

            {envSubTab === 'styled' ? (
              <div className="envContainer">
                <div className="envDescription">
                  <div className="muted">
                    Environment variables are encrypted and injected during the build process.
                  </div>
                </div>

                {envEntries.filter((e) => e.key.trim()).length === 0 ? (
                  <>
                    <div className="envEmptyState">
                      <div className="envEmptyIcon">
                        <FileText size={32} strokeWidth={1.5} />
                      </div>
                      <div className="envEmptyTitle">No variables configured</div>
                      <div className="envEmptyDesc">
                        Add environment variables like API keys, secrets, or configuration values.
                      </div>
                      <button className="btn primary" onClick={openAddEnvModal}>
                        <Plus size={16} /> Add Variable
                      </button>
                    </div>
                    {site.envText && (
                      <div className="envFooter">
                        <div />
                        <button className="btn primary" onClick={saveEnv}>
                          Save Changes
                        </button>
                      </div>
                    )}
                  </>
                ) : (
                  <>
                    <div className="envTable">
                      <div className="envTableHeader">
                        <div className="envTableCell envTableName">Name</div>
                        <div className="envTableCell envTableValue">Value</div>
                        <div className="envTableCell envTableActions"></div>
                      </div>
                      {envEntries
                        .filter((e) => e.key.trim())
                        .map((entry) => (
                          <div key={entry.key} className="envTableRow">
                            <div className="envTableCell envTableName">
                              <code className="envKeyCode">{entry.key}</code>
                            </div>
                            <div className="envTableCell envTableValue">
                              <div className="envValueContainer">
                                {visibleEnvKeys.has(entry.key) ? (
                                  <code className="envValueCode">
                                    {entry.value || <span className="muted">(empty)</span>}
                                  </code>
                                ) : (
                                  <span className="envMasked">
                                    {'•'.repeat(Math.min(entry.value?.length || 12, 24))}
                                  </span>
                                )}
                              </div>
                            </div>
                            <div className="envTableCell envTableActions">
                              <button
                                className="envActionBtn"
                                onClick={() => toggleEnvVisibility(entry.key)}
                                title={visibleEnvKeys.has(entry.key) ? 'Hide value' : 'Show value'}
                              >
                                {visibleEnvKeys.has(entry.key) ? <EyeOff size={16} /> : <Eye size={16} />}
                              </button>
                              <button
                                className="envActionBtn"
                                onClick={() => openEditEnvModal(entry.key, entry.value)}
                                title="Edit"
                              >
                                <Edit size={16} />
                              </button>
                              <button
                                className="envActionBtn envActionBtnDanger"
                                onClick={() => removeEnvEntry(entry.key)}
                                title="Remove"
                              >
                                <Trash2 size={16} />
                              </button>
                            </div>
                          </div>
                        ))}
                    </div>
                    <div className="envFooter">
                      <button className="btn ghost" onClick={openAddEnvModal}>
                        <Plus size={16} /> Add Variable
                      </button>
                      <button className="btn primary" onClick={saveEnv}>
                        Save Changes
                      </button>
                    </div>
                  </>
                )}
              </div>
            ) : (
              <div className="envContainer">
                <div className="envDescription">
                  <div className="muted">
                    Edit variables directly. One per line: <code>KEY=value</code>
                  </div>
                </div>
                <textarea
                  className="textarea envRawTextarea"
                  value={envDraft}
                  onChange={(e) => setEnvDraft(e.target.value)}
                  placeholder="API_KEY=your-api-key&#10;DATABASE_URL=postgres://..."
                  spellCheck={false}
                />
                <div className="envFooter">
                  <button className="btn primary" onClick={saveEnv}>
                    Save Changes
                  </button>
                </div>
              </div>
            )}
          </>
        ) : null}

        {tab === 'domains' && edgeOnly ? (
          <>
            <div className="panelTitle" style={{ marginTop: 12 }}>
              Custom Domains
            </div>
            <div className="muted" style={{ marginBottom: 16 }}>
              Point your own domain to this site.
            </div>

            <div className="row" style={{ marginBottom: 16 }}>
              <div className="field" style={{ flex: 1 }}>
                <div className="label">Hostname</div>
                <input
                  className="input"
                  placeholder="www.example.com"
                  value={customDomainInput}
                  onChange={(e) => setCustomDomainInput(e.target.value)}
                  disabled={customDomainLoading}
                />
              </div>
              <div className="field" style={{ alignSelf: 'flex-end' }}>
                <button
                  className="btn primary"
                  onClick={addCustomDomain}
                  disabled={customDomainLoading || !customDomainInput.trim()}
                >
                  {customDomainLoading ? 'Adding...' : 'Add'}
                </button>
              </div>
            </div>
            {customDomainError && (
              <div className="error" style={{ marginBottom: 16 }}>
                {customDomainError}
              </div>
            )}

            {customDomains.length > 0 && (
              <div className="deployList">
                {customDomains.map((d) => {
                  const txtRecord = d.verificationRecords?.find((r) => r.type === 'txt' || r.type === 'ssl_txt');
                  const isPending = d.status !== 'active';
                  const needsSetup = d.status === 'pending' || d.status === 'pending_validation';

                  return (
                    <div key={d.id} className="deployRow" style={{ flexDirection: 'column', alignItems: 'stretch' }}>
                      <div
                        style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 12 }}
                      >
                        <div>
                          <div className="siteName">{d.hostname}</div>
                          <div className="muted">
                            {d.status === 'active' && <span style={{ color: '#15803d' }}>✓ Active</span>}
                            {d.status === 'pending' && '⏳ Pending verification'}
                            {d.status === 'pending_ssl' &&
                              '⏳ SSL certificate issuing... (this may take a few minutes)'}
                            {d.status === 'pending_validation' && '⏳ Verifying ownership...'}
                          </div>
                        </div>
                        <div className="actions">
                          {isPending && (
                            <button
                              className="btn ghost"
                              disabled={pollingDomains.has(d.id)}
                              onClick={() => pollCustomDomain(d.id)}
                            >
                              {pollingDomains.has(d.id) ? (
                                <>
                                  <Loader2 size={14} className="animate-spin" style={{ marginRight: 6 }} />
                                  Checking...
                                </>
                              ) : (
                                'Check'
                              )}
                            </button>
                          )}
                          <button className="btn danger" onClick={() => removeCustomDomain(d.id)}>
                            Remove
                          </button>
                        </div>
                      </div>

                      {needsSetup && (
                        <div
                          style={{
                            marginTop: 12,
                            padding: 12,
                            background: 'rgba(255,122,0,.04)',
                            borderRadius: 8,
                            border: '1px solid rgba(255,122,0,.1)'
                          }}
                        >
                          <div style={{ fontWeight: 700, fontSize: 13, marginBottom: 8 }}>Setup Instructions</div>

                          {txtRecord && (
                            <div style={{ marginBottom: 12 }}>
                              <div className="muted" style={{ fontSize: 12, marginBottom: 4 }}>
                                1. Add this TXT record to verify ownership:
                              </div>
                              <div
                                style={{
                                  background: 'rgba(0,0,0,.03)',
                                  padding: 8,
                                  borderRadius: 6,
                                  fontSize: 12,
                                  fontFamily: 'monospace'
                                }}
                              >
                                <div>
                                  <strong>Type:</strong> TXT
                                </div>
                                <div>
                                  <strong>Name:</strong> {txtRecord.name}
                                </div>
                                <div style={{ wordBreak: 'break-all' }}>
                                  <strong>Value:</strong> {txtRecord.value}
                                </div>
                              </div>
                            </div>
                          )}

                          <div>
                            <div className="muted" style={{ fontSize: 12, marginBottom: 4 }}>
                              {txtRecord ? '2.' : '1.'} Add a CNAME record:
                            </div>
                            <div
                              style={{
                                background: 'rgba(0,0,0,.03)',
                                padding: 8,
                                borderRadius: 6,
                                fontSize: 12,
                                fontFamily: 'monospace'
                              }}
                            >
                              <div>
                                <strong>Type:</strong> CNAME
                              </div>
                              <div>
                                <strong>Name:</strong>{' '}
                                {d.hostname.split('.').length === 2 ? '@' : d.hostname.split('.')[0]}
                              </div>
                              <div>
                                <strong>Target:</strong> sites.{config.edgeRootDomain}
                              </div>
                            </div>
                          </div>

                          <div className="muted" style={{ fontSize: 12, marginTop: 8 }}>
                            After adding records, wait a few minutes then click "Check".
                          </div>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </>
        ) : null}

        {tab === 'settings' ? (
          <div className="settingsContainer">
            {}
            <div className="settingsSection">
              <div className="settingsSectionHeader">
                <div className="settingsSectionIcon">
                  <FileText size={20} />
                </div>
                <div>
                  <div className="settingsSectionTitle">General</div>
                  <div className="settingsSectionDesc">Basic project information</div>
                </div>
              </div>
              <div className="settingsGrid">
                <div className="settingsField">
                  <label className="settingsLabel">Project Name</label>
                  <input
                    className="input"
                    value={settingsDraft.name}
                    onChange={(e) => setSettingsDraft((s) => ({ ...s, name: e.target.value }))}
                  />
                </div>
                <div className="settingsField">
                  <label className="settingsLabel">Subdomain</label>
                  {edgeOnly && config.edgeRootDomain ? (
                    <div className="settingsInputGroup">
                      <input
                        className="input"
                        placeholder="my-site"
                        value={settingsDraft.domain}
                        onChange={(e) => setSettingsDraft((s) => ({ ...s, domain: e.target.value }))}
                      />
                      <span className="settingsInputSuffix">.{config.edgeRootDomain}</span>
                    </div>
                  ) : (
                    <input
                      className="input"
                      placeholder="my-site"
                      value={settingsDraft.domain}
                      onChange={(e) => setSettingsDraft((s) => ({ ...s, domain: e.target.value }))}
                    />
                  )}
                </div>
              </div>
            </div>

            {}
            <div className="settingsSection">
              <div className="settingsSectionHeader">
                <div className="settingsSectionIcon">
                  <ExternalLink size={20} />
                </div>
                <div>
                  <div className="settingsSectionTitle">Repository</div>
                  <div className="settingsSectionDesc">Source code configuration</div>
                </div>
              </div>
              <div className="settingsFieldFull">
                <label className="settingsLabel">Git URL</label>
                <input
                  className="input"
                  placeholder="https://github.com/user/repo"
                  value={settingsDraft.gitUrl}
                  onChange={(e) => setSettingsDraft((s) => ({ ...s, gitUrl: e.target.value }))}
                />
              </div>
              <div className="settingsGrid">
                <div className="settingsField">
                  <label className="settingsLabel">Branch</label>
                  <input
                    className="input"
                    placeholder="main"
                    value={settingsDraft.branch}
                    onChange={(e) => setSettingsDraft((s) => ({ ...s, branch: e.target.value }))}
                  />
                </div>
                <div className="settingsField">
                  <label className="settingsLabel">Root Directory</label>
                  <input
                    className="input"
                    placeholder="/ (root)"
                    value={settingsDraft.subdir}
                    onChange={(e) => setSettingsDraft((s) => ({ ...s, subdir: e.target.value }))}
                  />
                  <span className="settingsFieldHint">Leave empty for repository root</span>
                </div>
              </div>
            </div>

            {}
            <div className="settingsSection">
              <div className="settingsSectionHeader">
                <div className="settingsSectionIcon">
                  <Zap size={20} />
                </div>
                <div>
                  <div className="settingsSectionTitle">Build & Output</div>
                  <div className="settingsSectionDesc">Configure build process (auto-detected if empty)</div>
                </div>
              </div>
              <div className="settingsGrid">
                <div className="settingsField">
                  <label className="settingsLabel">Build Command</label>
                  <input
                    className="input"
                    placeholder="npm run build"
                    value={settingsDraft.buildCommand}
                    onChange={(e) => setSettingsDraft((s) => ({ ...s, buildCommand: e.target.value }))}
                  />
                </div>
                <div className="settingsField">
                  <label className="settingsLabel">Output Directory</label>
                  <input
                    className="input"
                    placeholder="dist"
                    value={settingsDraft.outputDir}
                    onChange={(e) => setSettingsDraft((s) => ({ ...s, outputDir: e.target.value }))}
                  />
                </div>
              </div>
            </div>

            <div className="settingsActions">
              <button className="btn primary" onClick={saveSettings}>
                Save Settings
              </button>
            </div>

            {}
            <div className="settingsSection settingsDanger">
              <div className="settingsSectionHeader">
                <div className="settingsSectionIcon settingsDangerIcon">
                  <Trash2 size={20} />
                </div>
                <div>
                  <div className="settingsSectionTitle">Danger Zone</div>
                  <div className="settingsSectionDesc">Irreversible actions</div>
                </div>
              </div>
              <div className="settingsDangerContent">
                <div>
                  <div className="settingsDangerTitle">Delete Project</div>
                  <div className="settingsDangerDesc">
                    Permanently delete this project and all its deployments. This action cannot be undone.
                  </div>
                </div>
                <button className="btn danger" onClick={deleteProject}>
                  Delete Project
                </button>
              </div>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}
