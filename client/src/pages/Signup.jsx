import React, { useState, useRef, useEffect, useCallback } from 'react';
import ThemeToggle from '../components/ThemeToggle.jsx';

export default function Signup() {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [providers, setProviders] = useState({ github: false, google: false, atproto: false });
  const [turnstileToken, setTurnstileToken] = useState('');
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [showAtpModal, setShowAtpModal] = useState(false);
  const [atpHandle, setAtpHandle] = useState('');
  const [atpError, setAtpError] = useState('');
  const turnstileRef = useRef(null);

  useEffect(() => {
    fetch('/api/auth/me', { credentials: 'same-origin' })
      .then((r) => r.json())
      .then((d) => {
        if (d?.authenticated) window.location.href = '/dashboard';
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    fetch('/api/auth/providers')
      .then((r) => r.json())
      .then((data) => {
        setProviders(data);
        if (data.turnstileSiteKey) {
          setTurnstileSiteKey(data.turnstileSiteKey);
        }
      })
      .catch(() => {});
  }, []);

  const handleAtprotoClick = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setShowAtpModal(true);
    setAtpError('');
  }, []);

  function startAtproto() {
    if (!atpHandle.trim()) {
      setAtpError('Enter your handle, e.g., yourname.bsky.social');
      return;
    }
    setAtpError('');
    const url = `/auth/atproto?handle=${encodeURIComponent(atpHandle.trim())}`;
    window.location.href = url;
  }

  useEffect(() => {
    if (!turnstileSiteKey) return;

    if (!document.getElementById('cf-turnstile-script')) {
      const script = document.createElement('script');
      script.id = 'cf-turnstile-script';
      script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js';
      script.async = true;
      document.head.appendChild(script);
    }
  }, [turnstileSiteKey]);

  useEffect(() => {
    if (!turnstileSiteKey || !turnstileRef.current) return;

    const renderWidget = () => {
      if (window.turnstile && turnstileRef.current) {
        turnstileRef.current.innerHTML = '';
        window.turnstile.render(turnstileRef.current, {
          sitekey: turnstileSiteKey,
          callback: (token) => setTurnstileToken(token),
          'expired-callback': () => setTurnstileToken(''),
          'error-callback': () => setTurnstileToken('')
        });
      }
    };

    if (window.turnstile) {
      renderWidget();
    } else {
      const script = document.getElementById('cf-turnstile-script');
      script?.addEventListener('load', renderWidget);
    }
  }, [turnstileSiteKey]);

  async function onSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ username, email, password, turnstileToken })
      });

      if (!res.ok) {
        const data = await res.json().catch(() => null);

        const errorMessages = {
          'email-required': 'Please enter your email address.',
          'invalid-email-format': 'Please enter a valid email address.',
          'email-too-long': 'Email address is too long.',
          'email-already-registered': 'This email is already registered. Try logging in instead.',
          'password-required': 'Please enter a password.',
          'password-too-short': 'Password must be at least 8 characters.',
          'password-too-long': 'Password is too long.',
          'username-required': 'Please enter a username.',
          'username-too-short': 'Username must be at least 3 characters.',
          'username-too-long': 'Username must be 20 characters or less.',
          'username-invalid-characters': 'Username can only contain letters, numbers, and underscores.',
          'username-already-taken': 'This username is already taken. Please choose another.',
          'captcha-failed': 'Captcha verification failed. Please try again.'
        };
        const errCode = data?.error || 'registration-failed';
        setError(errorMessages[errCode] || data?.message || 'Signup failed');

        if (window.turnstile && turnstileRef.current) {
          window.turnstile.reset(turnstileRef.current);
        }
        setTurnstileToken('');
        return;
      }
      window.location.href = '/dashboard';
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-theme-toggle">
        <ThemeToggle />
      </div>
      <div className="auth-card">
        <h1>Create account</h1>
        <p className="subtitle">Start deploying in seconds</p>

        {(providers.github || providers.google || providers.atproto) && (
          <>
            <div className="oauth-providers">
              {providers.github && (
                <a className="oauth-btn" href="/auth/github" title="Sign up with GitHub">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                  </svg>
                  <span>GitHub</span>
                </a>
              )}
              {providers.google && (
                <a className="oauth-btn" href="/auth/google" title="Sign up with Google">
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
                  <span>Google</span>
                </a>
              )}
              {providers.atproto && (
                <button type="button" className="oauth-btn" onClick={handleAtprotoClick} title="Sign up with Bluesky">
                  <svg width="20" height="20" viewBox="0 0 600 530" fill="currentColor">
                    <path d="m135.72 44.03c66.496 49.921 138.02 151.14 164.28 205.46 26.262-54.316 97.782-155.54 164.28-205.46 47.98-36.021 125.72-63.892 125.72 24.795 0 17.712-10.155 148.79-16.111 170.07-20.703 73.984-96.144 92.854-163.25 81.433 117.3 19.964 147.14 86.092 82.697 152.22-122.39 125.59-175.91-31.511-189.63-71.766-2.514-7.3797-3.6904-10.832-3.7077-7.8964-0.0174-2.9357-1.1937 0.51669-3.7077 7.8964-13.714 40.255-67.233 197.36-189.63 71.766-64.444-66.128-34.605-132.26 82.697-152.22-67.108 11.421-142.55-7.4491-163.25-81.433-5.9562-21.282-16.111-152.36-16.111-170.07 0-88.687 77.742-60.816 125.72-24.795z" />
                  </svg>
                  <span>Bluesky</span>
                </button>
              )}
            </div>
            <div className="auth-divider">or continue with email</div>
          </>
        )}

        <form className="form" onSubmit={onSubmit}>
          <input
            className="input"
            type="text"
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            pattern="[a-zA-Z0-9_]{3,20}"
            title="3-20 characters, letters, numbers, and underscores only"
            required
            disabled={loading}
          />
          <input
            className="input"
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            disabled={loading}
          />
          <input
            className="input"
            type="password"
            placeholder="Password (min 8 chars)"
            minLength={8}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            disabled={loading}
          />

          {turnstileSiteKey && <div ref={turnstileRef} className="turnstile-container" />}

          {error && <div className="error">{error}</div>}

          <button
            className="btn primary"
            type="submit"
            disabled={loading || (turnstileSiteKey && !turnstileToken)}
            style={{ width: '100%' }}
          >
            {loading ? 'Creating account...' : 'Create account'}
          </button>
        </form>

        <div className="auth-footer">
          Already have an account? <a href="/login">Log in</a>
        </div>
      </div>

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
              <div className="modalTitle">Sign up with ATProto</div>
              <button className="iconBtn" onClick={() => setShowAtpModal(false)} aria-label="Close ATProto dialog">
                ✕
              </button>
            </div>
            <div className="modalBody">
              <div className="field">
                <div className="label">Handle</div>
                <input
                  className="input"
                  placeholder="yourname.bsky.social"
                  value={atpHandle}
                  onChange={(e) => setAtpHandle(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      startAtproto();
                    }
                  }}
                  autoFocus
                />
                <div className="muted" style={{ marginTop: 6 }}>
                  We'll resolve your PDS from the handle and redirect you to its OAuth.
                </div>
              </div>
              {atpError && <div className="error">{atpError}</div>}
              <div className="modalActions">
                <button className="btn ghost" type="button" onClick={() => setShowAtpModal(false)}>
                  Cancel
                </button>
                <button className="btn primary" type="button" onClick={startAtproto} disabled={!atpHandle.trim()}>
                  Continue
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
