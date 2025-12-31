import React, { useEffect, useState } from 'react';
import { Link, useNavigate, useOutletContext } from 'react-router-dom';
import { ArrowLeft, Link as LinkIcon, Github, Lock, Search, Loader2, Book, GitBranch, ArrowRight } from 'lucide-react';

export default function NewSite() {
  const { me, sites, refreshSites, setError } = useOutletContext();
  const navigate = useNavigate();

  const [tab, setTab] = useState('git');
  const [name, setName] = useState('');
  const [gitUrl, setGitUrl] = useState('');
  const [branch, setBranch] = useState('main');
  const [subdir, setSubdir] = useState('');
  const [domain, setDomain] = useState('');
  const [loading, setLoading] = useState(false);
  const [createError, setCreateError] = useState('');

  const [repos, setRepos] = useState([]);
  const [reposLoading, setReposLoading] = useState(false);
  const [githubConnected, setGithubConnected] = useState(false);
  const [reposPage, setReposPage] = useState(1);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [reposError, setReposError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);

  const [config, setConfig] = useState({ deliveryMode: '', edgeRootDomain: '' });

  const [preview, setPreview] = useState(null);
  const [previewStatus, setPreviewStatus] = useState('idle');
  const [previewError, setPreviewError] = useState('');

  const siteLimitReached = sites.length >= 10;

  useEffect(() => {
    fetch('/api/config', { credentials: 'same-origin' })
      .then((r) => r.json())
      .then((d) => setConfig({ deliveryMode: d?.deliveryMode || '', edgeRootDomain: d?.edgeRootDomain || '' }))
      .catch(() => setConfig({ deliveryMode: '', edgeRootDomain: '' }));
  }, []);

  useEffect(() => {
    if (tab === 'github') {
      fetchRepos(1, searchQuery, { background: false });
    }
  }, [tab]);

  useEffect(() => {
    if (tab !== 'github') return;
    const timer = setTimeout(() => {
      fetchRepos(1, searchQuery, { background: true });
    }, 500);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  async function fetchRepos(page = 1, query = '', { background = false } = {}) {
    if (background) {
      setIsSearching(true);
    } else {
      setReposLoading(true);
    }
    setReposError('');
    try {
      const qParam = query ? `&q=${encodeURIComponent(query)}` : '';
      const res = await fetch(`/api/github/repos?page=${page}&per_page=100${qParam}`, { credentials: 'same-origin' });
      const data = await res.json();
      if (!res.ok) {
        setReposError(data.message || 'Failed to fetch repos');
        return;
      }
      setRepos(data.repos || []);
      setGithubConnected(data.githubConnected);
      setHasNextPage(data.hasNextPage);
      setReposPage(page);
    } catch (e) {
      setReposError(e.message || 'Failed to fetch repos');
    } finally {
      if (background) {
        setIsSearching(false);
      } else {
        setReposLoading(false);
      }
    }
  }

  const edgeOnly = String(config?.deliveryMode || '').toLowerCase() === 'edge' && Boolean(config?.edgeRootDomain);
  const edgeRoot = String(config?.edgeRootDomain || '').trim();

  const domainLabel = edgeOnly
    ? String(domain || '')
        .trim()
        .toLowerCase()
        .replace(new RegExp(`\\.${edgeRoot.replace(/[.*+?^${}()|[\\]\\]/g, '\\$&')}$`, 'i'), '')
    : domain;

  useEffect(() => {
    const trimmedUrl = gitUrl.trim();
    const trimmedBranch = branch.trim() || 'main';
    const trimmedSubdir = subdir.trim();

    if (!trimmedUrl) {
      setPreview(null);
      setPreviewStatus('idle');
      setPreviewError('');
      return;
    }

    const controller = new AbortController();
    const timer = setTimeout(async () => {
      setPreviewStatus('loading');
      setPreviewError('');
      setPreview(null);

      try {
        const res = await fetch('/api/sites/preview', {
          method: 'POST',
          headers: { 'content-type': 'application/json' },
          credentials: 'same-origin',
          body: JSON.stringify({ gitUrl: trimmedUrl, branch: trimmedBranch, subdir: trimmedSubdir }),
          signal: controller.signal
        });

        const data = await res.json().catch(() => null);
        if (!res.ok) {
          setPreviewStatus('error');
          setPreviewError(data?.message || 'Unable to analyze repo.');
          return;
        }

        setPreviewStatus('success');
        setPreview(data);
      } catch (e) {
        if (controller.signal.aborted) return;
        setPreviewStatus('error');
        setPreviewError(e?.message || 'Unable to analyze repo.');
      }
    }, 2000);

    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [gitUrl, branch, subdir]);

  async function handleCreate() {
    setLoading(true);
    setCreateError('');
    setError('');

    try {
      const res = await fetch('/api/sites', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({ name, gitUrl, branch, subdir, domain: domainLabel })
      });

      const data = await res.json().catch(() => null);

      if (!res.ok) {
        const errorMessages = {
          'name-required': 'Please enter a project name.',
          'git-url-required': 'Please enter a Git repository URL.',
          'invalid-domain': 'The domain format is invalid.',
          'domain-taken': 'That subdomain is already taken.',
          'site-limit-reached': "You've reached the maximum of 10 projects.",
          'user-required': 'You must be logged in to create a site.'
        };
        const errCode = data?.error || 'site-create-failed';
        setCreateError(errorMessages[errCode] || data?.message || 'Failed to create site.');
        return;
      }

      await refreshSites();
      navigate('/dashboard');
    } catch (e) {
      setCreateError(e?.message || 'Failed to create site.');
    } finally {
      setLoading(false);
    }
  }

  function selectRepo(repo) {
    setGitUrl(repo.cloneUrl);
    setBranch(repo.defaultBranch || 'main');
    if (!name) {
      setName(repo.name);
    }
    setTab('git');
  }

  function renderSignals() {
    if (!preview?.details) return null;
    const {
      staticDepsFound = [],
      staticFilesFound = [],
      dynamicDepsFound = [],
      dynamicFilesFound = []
    } = preview.details;
    const staticSignals = [...staticDepsFound, ...staticFilesFound].slice(0, 4);
    const dynamicSignals = [...dynamicDepsFound, ...dynamicFilesFound].slice(0, 4);

    return (
      <>
        {staticSignals.length > 0 && (
          <div className="compatSignals">
            <div className="compatSignalsLabel">Static clues</div>
            <div className="compatSignalPills">
              {staticSignals.map((sig) => (
                <span key={`static-${sig}`} className="compatPill">
                  {sig}
                </span>
              ))}
            </div>
          </div>
        )}
        {dynamicSignals.length > 0 && (
          <div className="compatSignals warn">
            <div className="compatSignalsLabel">Dynamic blockers</div>
            <div className="compatSignalPills">
              {dynamicSignals.map((sig) => (
                <span key={`dynamic-${sig}`} className="compatPill pillWarn">
                  {sig}
                </span>
              ))}
            </div>
          </div>
        )}
      </>
    );
  }

  return (
    <div className="page">
      <div className="pageHeader">
        <div>
          <div className="crumbs">
            <Link to="/dashboard" className="muted" style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
              <ArrowLeft size={14} />
              Dashboard
            </Link>
          </div>
          <div className="h">New Website</div>
          <div className="muted">Deploy a static site from Git</div>
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

      {createError && <div className="errorBox">{createError}</div>}

      <div className="panel">
        <div className="tabs">
          <button className={tab === 'git' ? 'tab active' : 'tab'} onClick={() => setTab('git')}>
            <LinkIcon size={16} style={{ marginRight: 6 }} />
            Git URL
          </button>
          <button className={tab === 'github' ? 'tab active' : 'tab'} onClick={() => setTab('github')}>
            <Github size={16} style={{ marginRight: 6 }} />
            GitHub
          </button>
        </div>

        {tab === 'git' && (
          <>
            <div className="panelTitle" style={{ marginTop: 16 }}>
              Repository
            </div>
            <div className="muted" style={{ marginBottom: 16 }}>
              Enter a Git URL to deploy
            </div>

            <div className="field">
              <div className="label">Git Repository URL</div>
              <input
                className="input"
                placeholder="https://github.com/username/repo.git"
                value={gitUrl}
                onChange={(e) => setGitUrl(e.target.value)}
                disabled={loading}
              />
              <div className="muted" style={{ marginTop: 6, fontSize: 12 }}>
                Public repos work automatically. For private repos, connect via GitHub OAuth in Account settings.
              </div>
            </div>

            {gitUrl.trim() && (
              <div className="compatCard">
                {previewStatus === 'idle' && <div className="compatHint">Analyzing...</div>}
                {previewStatus === 'loading' && <div className="compatHint">Analyzing repository…</div>}
                {previewStatus === 'error' && (
                  <div className="compatError">{previewError || 'Could not analyze this repo.'}</div>
                )}
                {previewStatus === 'success' && preview && (
                  <>
                    <div className="compatBadgeRow">
                      <div className={`compatBadge ${preview.ok ? '' : 'warn'}`}>
                        {preview.label || 'Static site check'}
                      </div>
                      <div className="compatStatus">{preview.ok ? 'Looks deployable' : 'Needs attention'}</div>
                    </div>
                    <div className="compatHeadline">{preview.headline}</div>
                    {!preview.ok && preview.suggestion && (
                      <div className="compatSuggestion">
                        <strong>Suggestion:</strong> {preview.suggestion}
                      </div>
                    )}
                    {renderSignals()}
                  </>
                )}
              </div>
            )}

            <div className="panelTitle" style={{ marginTop: 24 }}>
              Configuration
            </div>
            <div className="muted" style={{ marginBottom: 16 }}>
              Project settings
            </div>

            <div className="row" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <div className="field">
                <div className="label">Project Name</div>
                <input
                  className="input"
                  placeholder="my-awesome-site"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  disabled={loading}
                />
              </div>
              <div className="field">
                <div className="label">Branch</div>
                <input
                  className="input"
                  placeholder="main"
                  value={branch}
                  onChange={(e) => setBranch(e.target.value)}
                  disabled={loading}
                />
              </div>
            </div>

            <div className="row" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <div className="field">
                <div className="label">Subdir (optional)</div>
                <input
                  className="input"
                  placeholder="packages/web"
                  value={subdir}
                  onChange={(e) => setSubdir(e.target.value)}
                  disabled={loading}
                />
              </div>
              <div className="field">
                <div className="label">Subdomain (optional)</div>
                {edgeOnly ? (
                  <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                    <input
                      className="input"
                      placeholder="your-subdomain"
                      value={domainLabel}
                      onChange={(e) => setDomain(e.target.value)}
                      disabled={loading}
                      style={{ flex: 1 }}
                    />
                    <span className="muted">.{edgeRoot}</span>
                  </div>
                ) : (
                  <input
                    className="input"
                    placeholder="your-subdomain"
                    value={domain}
                    onChange={(e) => setDomain(e.target.value)}
                    disabled={loading}
                  />
                )}
              </div>
            </div>

            <div
              style={{
                display: 'flex',
                justifyContent: 'flex-end',
                gap: 12,
                marginTop: 24,
                paddingTop: 16,
                borderTop: '1px solid var(--divider)'
              }}
            >
              <Link to="/dashboard" className="btn ghost">
                Cancel
              </Link>
              <button
                className="btn primary"
                disabled={siteLimitReached || loading || !name.trim() || !gitUrl.trim()}
                onClick={handleCreate}
              >
                {loading ? 'Creating...' : 'Create Website'}
              </button>
            </div>
          </>
        )}

        {tab === 'github' && (
          <>
            <div className="panelTitle" style={{ marginTop: 16 }}>
              Import from GitHub
            </div>
            <div className="muted" style={{ marginBottom: 16 }}>
              Select a repository to import
            </div>

            {reposLoading ? (
              <div className="repoLoadingState" style={{ textAlign: 'center', padding: '48px 24px' }}>
                Loading repositories...
              </div>
            ) : !githubConnected ? (
              <div style={{ textAlign: 'center', padding: '48px 24px' }}>
                <Github size={48} style={{ marginBottom: 16, opacity: 0.4 }} />
                <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: 8 }}>Connect GitHub</h3>
                <div
                  className="muted"
                  style={{ marginBottom: 24, maxWidth: 320, marginLeft: 'auto', marginRight: 'auto' }}
                >
                  Link your GitHub account to import repositories directly.
                </div>
                <a href="/auth/github" className="btn primary">
                  <Github size={16} style={{ marginRight: 8 }} />
                  Connect GitHub
                </a>
              </div>
            ) : (
              <>
                {reposError && <div className="errorBox">{reposError}</div>}

                <div className="field" style={{ marginBottom: 16 }}>
                  <div style={{ position: 'relative' }}>
                    {isSearching ? (
                      <Loader2
                        size={16}
                        className="animate-spin searchIcon"
                        style={{
                          position: 'absolute',
                          left: 12,
                          top: '50%',
                          transform: 'translateY(-50%)'
                        }}
                      />
                    ) : (
                      <Search
                        size={16}
                        className="searchIcon"
                        style={{
                          position: 'absolute',
                          left: 12,
                          top: '50%',
                          transform: 'translateY(-50%)'
                        }}
                      />
                    )}
                    <input
                      className="input"
                      style={{ paddingLeft: 36 }}
                      placeholder="Search repositories..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                    />
                  </div>
                </div>

                {reposLoading ? (
                  <div className="repoLoadingState" style={{ textAlign: 'center', padding: '48px 24px' }}>
                    Loading repositories...
                  </div>
                ) : (
                  <>
                    <div className="repoGrid">
                      {repos.map((repo) => (
                        <button key={repo.id} className="repoItem" onClick={() => selectRepo(repo)}>
                          <div className="repoItemIcon">
                            <Book size={20} strokeWidth={1.5} />
                          </div>

                          <div className="repoItemContent">
                            <div className="repoItemName">{repo.name}</div>
                            <div className="repoItemDesc">{repo.description || 'No description provided'}</div>
                          </div>

                          <div className="repoItemMetaColumn">
                            <div className="repoItemMetaRow">
                              {repo.private && (
                                <span className="repoItemPrivate">
                                  <Lock size={10} />
                                  Private
                                </span>
                              )}
                              {repo.language && (
                                <span>
                                  <span className="repoLangDot" />
                                  {repo.language}
                                </span>
                              )}
                              <span>
                                <GitBranch size={12} />
                                {repo.defaultBranch}
                              </span>
                            </div>
                            <div className="repoArrow">
                              <ArrowRight size={16} />
                            </div>
                          </div>
                        </button>
                      ))}
                    </div>

                    {repos.length === 0 && (
                      <div className="repoEmptyState" style={{ textAlign: 'center', padding: '32px' }}>
                        No repositories found
                      </div>
                    )}
                  </>
                )}
              </>
            )}
          </>
        )}
      </div>
    </div>
  );
}
