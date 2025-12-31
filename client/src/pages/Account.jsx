import React, { useEffect, useState } from 'react';
import { useNavigate, useOutletContext, useSearchParams } from 'react-router-dom';
import { Mail, Lock, AlertTriangle, Link2, Unlink, Key, Copy, Check, Trash2 } from 'lucide-react';

const ERROR_MESSAGES = {
  'email-required': 'Please enter your email address.',
  'invalid-email-format': 'Please enter a valid email address.',
  'email-too-long': 'Email address is too long.',
  'email-already-registered': 'This email is already in use.',
  'password-required': 'Please enter your password.',
  'password-too-short': 'Password must be at least 8 characters.',
  'password-too-long': 'Password is too long.',
  'invalid-password': 'Current password is incorrect.',
  'password-not-set': 'No password set. You signed up with OAuth.',
  'user-not-found': 'User not found.',
  'cannot-unlink-only-auth': 'Cannot unlink your only login method. Set a password first.',
  'account-not-found': 'Linked account not found.'
};

const PROVIDER_INFO = {
  github: {
    name: 'GitHub',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
      </svg>
    ),
    color: '#333'
  },
  google: {
    name: 'Google',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24">
        <path
          fill="#4285F4"
          d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
        />
        <path
          fill="#34A853"
          d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        />
        <path
          fill="#FBBC05"
          d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
        />
        <path
          fill="#EA4335"
          d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        />
      </svg>
    ),
    color: '#4285F4'
  },
  atproto: {
    name: 'ATProto (Bluesky)',
    icon: (
      <svg width="20" height="20" viewBox="0 0 600 530" fill="currentColor">
        <path d="m135.72 44.03c66.496 49.921 138.02 151.14 164.28 205.46 26.262-54.316 97.782-155.54 164.28-205.46 47.98-36.021 125.72-63.892 125.72 24.795 0 17.712-10.155 148.79-16.111 170.07-20.703 73.984-96.144 92.854-163.25 81.433 117.3 19.964 147.14 86.092 82.697 152.22-122.39 125.59-175.91-31.511-189.63-71.766-2.514-7.3797-3.6904-10.832-3.7077-7.8964-0.0174-2.9357-1.1937 0.51669-3.7077 7.8964-13.714 40.255-67.233 197.36-189.63 71.766-64.444-66.128-34.605-132.26 82.697-152.22-67.108 11.421-142.55-7.4491-163.25-81.433-5.9562-21.282-16.111-152.36-16.111-170.07 0-88.687 77.742-60.816 125.72-24.795z" />
      </svg>
    ),
    color: '#0085ff'
  }
};

export default function Account() {
  const { api, me, setError } = useOutletContext();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const [email, setEmail] = useState('');
  const [currentPasswordEmail, setCurrentPasswordEmail] = useState('');

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');

  const [success, setSuccess] = useState('');
  const [emailError, setEmailError] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [loadingEmail, setLoadingEmail] = useState(false);
  const [loadingPassword, setLoadingPassword] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleteLoading, setDeleteLoading] = useState(false);

  const [linkedAccounts, setLinkedAccounts] = useState([]);
  const [linkedAccountsLoading, setLinkedAccountsLoading] = useState(true);
  const [linkedAccountsError, setLinkedAccountsError] = useState('');
  const [unlinkingId, setUnlinkingId] = useState(null);
  const [providers, setProviders] = useState({ github: false, google: false, atproto: false });
  const [showAtpModal, setShowAtpModal] = useState(false);
  const [atpHandle, setAtpHandle] = useState('');

  const [apiKeys, setApiKeys] = useState([]);
  const [apiKeysLoading, setApiKeysLoading] = useState(true);
  const [apiKeysError, setApiKeysError] = useState('');
  const [newKeyName, setNewKeyName] = useState('');
  const [creatingKey, setCreatingKey] = useState(false);
  const [newlyCreatedKey, setNewlyCreatedKey] = useState(null);
  const [copiedKeyId, setCopiedKeyId] = useState(null);
  const [deletingKeyId, setDeletingKeyId] = useState(null);

  useEffect(() => {
    setEmail(me?.email || '');
  }, [me?.email]);

  useEffect(() => {
    const error = searchParams.get('error');
    if (error === 'already-linked') {
      setLinkedAccountsError('This account is already linked to another user.');

      searchParams.delete('error');
      setSearchParams(searchParams, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  useEffect(() => {
    fetch('/api/auth/providers')
      .then((r) => r.json())
      .then((data) => setProviders(data))
      .catch(() => {});

    fetchLinkedAccounts();

    fetchApiKeys();
  }, []);

  async function fetchApiKeys() {
    setApiKeysLoading(true);
    try {
      const res = await fetch('/api/api-keys', { credentials: 'same-origin' });
      const data = await res.json();
      if (res.ok) {
        setApiKeys(data || []);
      }
    } catch (e) {
      setApiKeysError('Failed to load API keys');
    } finally {
      setApiKeysLoading(false);
    }
  }

  async function createApiKey() {
    if (!newKeyName.trim()) return;
    setCreatingKey(true);
    setApiKeysError('');
    try {
      const res = await fetch('/api/api-keys', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({ name: newKeyName.trim() })
      });
      const data = await res.json();
      if (!res.ok) {
        setApiKeysError(data?.message || 'Failed to create API key');
        return;
      }
      setNewlyCreatedKey(data);
      setNewKeyName('');
      setApiKeys((prev) => [...prev, { id: data.id, name: data.name, prefix: data.prefix, createdAt: data.createdAt }]);
    } catch (e) {
      setApiKeysError('Failed to create API key');
    } finally {
      setCreatingKey(false);
    }
  }

  async function deleteApiKey(keyId) {
    setDeletingKeyId(keyId);
    setApiKeysError('');
    try {
      const res = await fetch(`/api/api-keys/${keyId}`, {
        method: 'DELETE',
        credentials: 'same-origin'
      });
      if (!res.ok) {
        const data = await res.json().catch(() => null);
        setApiKeysError(data?.message || 'Failed to delete API key');
        return;
      }
      setApiKeys((prev) => prev.filter((k) => k.id !== keyId));
      if (newlyCreatedKey?.id === keyId) {
        setNewlyCreatedKey(null);
      }
    } catch (e) {
      setApiKeysError('Failed to delete API key');
    } finally {
      setDeletingKeyId(null);
    }
  }

  function copyToClipboard(text, keyId) {
    navigator.clipboard.writeText(text);
    setCopiedKeyId(keyId);
    setTimeout(() => setCopiedKeyId(null), 2000);
  }

  async function fetchLinkedAccounts() {
    setLinkedAccountsLoading(true);
    try {
      const res = await fetch('/api/account/linked-accounts', { credentials: 'same-origin' });
      const data = await res.json();
      if (res.ok) {
        setLinkedAccounts(data.accounts || []);
      }
    } catch (e) {
      setLinkedAccountsError('Failed to load linked accounts');
    } finally {
      setLinkedAccountsLoading(false);
    }
  }

  async function unlinkAccount(accountId) {
    setUnlinkingId(accountId);
    setLinkedAccountsError('');
    try {
      const res = await fetch(`/api/account/linked-accounts/${accountId}`, {
        method: 'DELETE',
        credentials: 'same-origin'
      });
      const data = await res.json().catch(() => null);

      if (!res.ok) {
        const errCode = data?.error || 'unlink-failed';
        setLinkedAccountsError(ERROR_MESSAGES[errCode] || errCode);
        return;
      }

      setLinkedAccounts((prev) => prev.filter((a) => a.id !== accountId));
      setSuccess('🔗 Account unlinked successfully.');
    } finally {
      setUnlinkingId(null);
    }
  }

  async function changeEmail() {
    setEmailError('');
    setSuccess('');
    setLoadingEmail(true);

    try {
      const res = await fetch('/api/account/email', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({ newEmail: email, currentPassword: currentPasswordEmail })
      });

      const data = await res.json().catch(() => null);

      if (!res.ok) {
        const errCode = data?.error || 'email-change-failed';
        setEmailError(ERROR_MESSAGES[errCode] || data?.message || 'Failed to change email.');
        return;
      }

      setSuccess('📧 Email change requested. Check your inbox to verify your new email.');
      setCurrentPasswordEmail('');
    } finally {
      setLoadingEmail(false);
    }
  }

  async function changePassword() {
    setPasswordError('');
    setSuccess('');
    setLoadingPassword(true);

    try {
      const res = await fetch('/api/account/password', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({ currentPassword, newPassword })
      });

      const data = await res.json().catch(() => null);

      if (!res.ok) {
        const errCode = data?.error || 'password-update-failed';
        setPasswordError(ERROR_MESSAGES[errCode] || data?.message || 'Failed to update password.');
        return;
      }

      setSuccess('🔒 Password updated successfully.');
      setCurrentPassword('');
      setNewPassword('');
    } finally {
      setLoadingPassword(false);
    }
  }

  async function deleteAccount() {
    setDeleteLoading(true);
    try {
      await api('/api/account', { method: 'DELETE' });
      navigate('/');
    } catch (e) {
      setError(e.message);
    } finally {
      setDeleteLoading(false);
    }
  }

  return (
    <div className="page">
      <div className="pageHeader">
        <div>
          <div className="h">Account Settings</div>
          <div className="muted">Manage your email, password, and account preferences.</div>
        </div>
      </div>

      {success && <div className="notice">{success}</div>}

      <div className="grid2">
        <div className="panel">
          <div className="panelTitle">
            <Mail size={18} style={{ marginRight: 8, color: '#e88978' }} />
            Email Address
          </div>
          <div className="muted" style={{ marginBottom: 16, fontSize: 13 }}>
            Changing your email requires verification. We'll send a link to your new address.
          </div>

          {emailError && (
            <div className="errorBox" style={{ marginBottom: 12 }}>
              {emailError}
            </div>
          )}

          <div className="field">
            <div className="label">New Email</div>
            <input
              className="input"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={loadingEmail}
            />
          </div>
          <div className="field">
            <div className="label">Current Password</div>
            <input
              className="input"
              type="password"
              placeholder="Enter your current password"
              value={currentPasswordEmail}
              onChange={(e) => setCurrentPasswordEmail(e.target.value)}
              disabled={loadingEmail}
            />
          </div>
          <div className="panelActions">
            <button
              className="btn primary"
              onClick={changeEmail}
              disabled={loadingEmail || !email || !currentPasswordEmail}
            >
              {loadingEmail ? 'Sending...' : 'Change Email'}
            </button>
          </div>
        </div>

        <div className="panel">
          <div className="panelTitle">
            <Lock size={18} style={{ marginRight: 8, color: '#e88978' }} />
            Password
          </div>
          <div className="muted" style={{ marginBottom: 16, fontSize: 13 }}>
            Use a strong password with at least 8 characters.
          </div>

          {passwordError && (
            <div className="errorBox" style={{ marginBottom: 12 }}>
              {passwordError}
            </div>
          )}

          <div className="field">
            <div className="label">Current Password</div>
            <input
              className="input"
              type="password"
              placeholder="Enter your current password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              disabled={loadingPassword}
            />
          </div>
          <div className="field">
            <div className="label">New Password</div>
            <input
              className="input"
              type="password"
              placeholder="Enter your new password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              minLength={8}
              disabled={loadingPassword}
            />
          </div>
          <div className="panelActions">
            <button
              className="btn primary"
              onClick={changePassword}
              disabled={loadingPassword || !currentPassword || !newPassword || newPassword.length < 8}
            >
              {loadingPassword ? 'Updating...' : 'Update Password'}
            </button>
          </div>
        </div>
      </div>

      {}
      <div className="panel" style={{ marginTop: 8 }}>
        <div className="panelTitle">
          <Link2 size={18} style={{ marginRight: 8, color: '#e88978' }} />
          Linked Accounts
        </div>
        <div className="muted" style={{ marginBottom: 16, fontSize: 13 }}>
          Connect additional login methods to your account. You can log in using any linked account.
        </div>

        {linkedAccountsError && (
          <div className="errorBox" style={{ marginBottom: 12 }}>
            {linkedAccountsError}
          </div>
        )}

        {linkedAccountsLoading ? (
          <div className="muted">Loading linked accounts...</div>
        ) : (
          <>
            {}
            {linkedAccounts.length > 0 && (
              <div style={{ marginBottom: 16 }}>
                {linkedAccounts.map((account) => {
                  const info = PROVIDER_INFO[account.provider] || { name: account.provider, icon: null, color: '#666' };
                  return (
                    <div
                      key={account.id}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        padding: '12px 16px',
                        background: 'var(--bg-secondary)',
                        borderRadius: 8,
                        marginBottom: 8
                      }}
                    >
                      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                        <span style={{ color: info.color }}>{info.icon}</span>
                        <div>
                          <div style={{ fontWeight: 500 }}>{info.name}</div>
                          <div className="muted" style={{ fontSize: 12 }}>
                            {account.displayName && <span>{account.displayName} · </span>}
                            Connected {new Date(account.createdAt).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                      <button
                        className="btn ghost"
                        onClick={() => unlinkAccount(account.id)}
                        disabled={unlinkingId === account.id}
                        style={{ display: 'flex', alignItems: 'center', gap: 6 }}
                      >
                        <Unlink size={14} />
                        {unlinkingId === account.id ? 'Unlinking...' : 'Unlink'}
                      </button>
                    </div>
                  );
                })}
              </div>
            )}

            {}
            {(() => {
              const linkedProviders = new Set(linkedAccounts.map((a) => a.provider));
              const availableProviders = Object.entries(providers)
                .filter(([key, enabled]) => enabled && !linkedProviders.has(key) && key !== 'turnstileSiteKey')
                .map(([key]) => key);

              if (availableProviders.length === 0 && linkedAccounts.length === 0) {
                return <div className="muted">No OAuth providers are configured.</div>;
              }

              if (availableProviders.length === 0) {
                return null;
              }

              return (
                <div>
                  <div className="muted" style={{ marginBottom: 12, fontSize: 13 }}>
                    Link additional accounts:
                  </div>
                  <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                    {availableProviders.map((provider) => {
                      const info = PROVIDER_INFO[provider] || { name: provider, icon: null, color: '#666' };
                      const authUrl = provider === 'atproto' ? null : `/auth/${provider}`;

                      if (provider === 'atproto') {
                        return (
                          <button
                            key={provider}
                            type="button"
                            onClick={() => setShowAtpModal(true)}
                            className="btn ghost"
                            style={{ display: 'flex', alignItems: 'center', gap: 8 }}
                          >
                            <span style={{ color: info.color }}>{info.icon}</span>
                            Link {info.name}
                          </button>
                        );
                      }

                      return (
                        <a
                          key={provider}
                          href={authUrl}
                          className="btn ghost"
                          style={{ display: 'flex', alignItems: 'center', gap: 8 }}
                        >
                          <span style={{ color: info.color }}>{info.icon}</span>
                          Link {info.name}
                        </a>
                      );
                    })}
                  </div>
                </div>
              );
            })()}
          </>
        )}
      </div>

      {}
      <div className="panel" style={{ marginTop: 8 }}>
        <div className="panelTitle">
          <Key size={18} style={{ marginRight: 8, color: '#e88978' }} />
          API Keys
        </div>
        <div className="muted" style={{ marginBottom: 16, fontSize: 13 }}>
          Use API keys to deploy sites from CI/CD pipelines and scripts. Keys are shown only once when created.
        </div>

        {apiKeysError && (
          <div className="errorBox" style={{ marginBottom: 12 }}>
            {apiKeysError}
          </div>
        )}

        {}
        {newlyCreatedKey && (
          <div
            className="panel"
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: 12,
              borderColor: 'rgba(34, 197, 94, 0.3)',
              background: 'rgba(34, 197, 94, 0.08)'
            }}
          >
            <div style={{ fontWeight: 600, display: 'flex', alignItems: 'center', gap: 8 }}>
              <Check size={16} style={{ color: '#22c55e' }} />
              API Key Created
            </div>
            <div className="muted" style={{ fontSize: 13 }}>
              Copy this key now — it won't be shown again.
            </div>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                background: 'var(--code-bg)',
                padding: '8px 12px',
                borderRadius: 6,
                fontFamily: 'monospace',
                fontSize: 14,
                wordBreak: 'break-all',
                border: '1px solid var(--card-border)'
              }}
            >
              <code style={{ flex: 1, color: 'var(--card-text)' }}>{newlyCreatedKey.key}</code>
              <button
                className="iconBtn"
                onClick={() => copyToClipboard(newlyCreatedKey.key, newlyCreatedKey.id)}
                style={{ width: 28, height: 28, borderRadius: '50%', flexShrink: 0 }}
              >
                {copiedKeyId === newlyCreatedKey.id ? <Check size={14} /> : <Copy size={14} />}
              </button>
            </div>
            <button
              className="btn ghost"
              onClick={() => setNewlyCreatedKey(null)}
              style={{ fontSize: 13, alignSelf: 'flex-start' }}
            >
              Dismiss
            </button>
          </div>
        )}

        {apiKeysLoading ? (
          <div className="muted">Loading API keys...</div>
        ) : (
          <>
            {}
            {apiKeys.length > 0 && (
              <div style={{ marginBottom: 16 }}>
                {apiKeys.map((key) => (
                  <div
                    key={key.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      padding: '12px 16px',
                      background: 'var(--bg-secondary)',
                      borderRadius: 8,
                      marginBottom: 8
                    }}
                  >
                    <div>
                      <div style={{ fontWeight: 500 }}>{key.name}</div>
                      <div className="muted" style={{ fontSize: 12, fontFamily: 'monospace' }}>
                        {key.prefix}••••••••
                        {key.lastUsedAt?.Valid && key.lastUsedAt?.String && (
                          <span> · Last used {new Date(key.lastUsedAt.String).toLocaleDateString()}</span>
                        )}
                        {!key.lastUsedAt?.Valid && <span> · Never used</span>}
                      </div>
                    </div>
                    <button
                      className="btn ghost"
                      onClick={() => deleteApiKey(key.id)}
                      disabled={deletingKeyId === key.id}
                      style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--danger)' }}
                    >
                      <Trash2 size={14} />
                      {deletingKeyId === key.id ? 'Deleting...' : 'Delete'}
                    </button>
                  </div>
                ))}
              </div>
            )}

            {apiKeys.length === 0 && !newlyCreatedKey && (
              <div className="muted" style={{ marginBottom: 16 }}>
                No API keys yet. Create one to get started with CLI deployments.
              </div>
            )}

            {}
            <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end' }}>
              <div className="field" style={{ flex: 1, marginBottom: 0 }}>
                <div className="label">Key Name</div>
                <input
                  className="input"
                  placeholder="e.g., GitHub Actions"
                  value={newKeyName}
                  onChange={(e) => setNewKeyName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && newKeyName.trim()) {
                      createApiKey();
                    }
                  }}
                  disabled={creatingKey}
                  maxLength={64}
                />
              </div>
              <button
                className="btn primary"
                onClick={createApiKey}
                disabled={creatingKey || !newKeyName.trim()}
                style={{ marginBottom: 0 }}
              >
                {creatingKey ? 'Creating...' : 'Create Key'}
              </button>
            </div>
          </>
        )}
      </div>

      <div className="panel" style={{ marginTop: 8 }}>
        <div className="panelTitle">
          <AlertTriangle size={18} style={{ marginRight: 8, color: '#dc2626' }} />
          Danger Zone
        </div>
        <div className="muted" style={{ marginBottom: 16, fontSize: 13 }}>
          Once you delete your account, all your projects and deployments will be permanently removed.
        </div>

        {!showDeleteConfirm ? (
          <button className="btn danger" onClick={() => setShowDeleteConfirm(true)}>
            Delete Account
          </button>
        ) : (
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span style={{ fontWeight: 600, color: '#b91c1c' }}>Are you sure?</span>
            <button className="btn danger" onClick={deleteAccount} disabled={deleteLoading}>
              {deleteLoading ? 'Deleting...' : 'Yes, delete my account'}
            </button>
            <button className="btn ghost" onClick={() => setShowDeleteConfirm(false)} disabled={deleteLoading}>
              Cancel
            </button>
          </div>
        )}
      </div>

      {}
      {showAtpModal && (
        <div
          className="modalOverlay"
          onClick={(e) => {
            if (e.target === e.currentTarget) {
              setShowAtpModal(false);
            }
          }}
        >
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modalHeader">
              <div className="modalTitle">Link ATProto Account</div>
              <button className="iconBtn" onClick={() => setShowAtpModal(false)} aria-label="Close">
                ✕
              </button>
            </div>
            <div className="modalBody">
              <div className="field">
                <div className="label">Handle</div>
                <input
                  className="input"
                  placeholder="your-handle.your.pds"
                  value={atpHandle}
                  onChange={(e) => setAtpHandle(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && atpHandle.trim()) {
                      e.preventDefault();
                      window.location.href = `/auth/atproto?handle=${encodeURIComponent(atpHandle.trim())}`;
                    }
                  }}
                  autoFocus
                />
                <div className="muted" style={{ marginTop: 6 }}>
                  Enter your Bluesky handle to link your account.
                </div>
              </div>
              <div className="modalActions">
                <button className="btn ghost" type="button" onClick={() => setShowAtpModal(false)}>
                  Cancel
                </button>
                <button
                  className="btn primary"
                  type="button"
                  onClick={() => {
                    if (atpHandle.trim()) {
                      window.location.href = `/auth/atproto?handle=${encodeURIComponent(atpHandle.trim())}`;
                    }
                  }}
                  disabled={!atpHandle.trim()}
                >
                  Link Account
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
