// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

import React, { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import ThemeToggle from '../components/ThemeToggle.jsx';

const ERROR_MESSAGES = {
  'missing-token': 'Reset link is invalid.',
  'invalid-token': 'Reset link is invalid or has already been used.',
  'already-used': 'This reset link has already been used.',
  expired: 'This reset link has expired. Please request a new one.',
  'password-required': 'Please enter a new password.',
  'password-too-short': 'Password must be at least 8 characters.',
  'password-too-long': 'Password is too long.',
  'user-not-found': 'User not found.',
  'reset-failed': 'Failed to reset password. Please try again.'
};

export default function ResetPassword() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token');

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (!token) {
      setError('missing-token');
    }
  }, [token]);

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');

    if (password !== confirmPassword) {
      setError('Passwords do not match.');
      return;
    }

    if (password.length < 8) {
      setError('password-too-short');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch('/api/auth/reset-password', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ token, newPassword: password })
      });

      const data = await res.json().catch(() => null);

      if (!res.ok) {
        const errCode = data?.error || 'reset-failed';
        setError(ERROR_MESSAGES[errCode] || errCode);
        return;
      }

      setSuccess(true);
    } finally {
      setLoading(false);
    }
  }

  if (success) {
    return (
      <div className="auth-page">
        <div className="auth-theme-toggle">
          <ThemeToggle />
        </div>
        <div className="auth-card">
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 64, marginBottom: 16 }}>✅</div>
            <h1>Password reset!</h1>
            <p className="subtitle" style={{ marginBottom: 24 }}>
              Your password has been successfully changed.
            </p>
            <button className="btn primary" onClick={() => navigate('/login')} style={{ width: '100%' }}>
              Log in with new password
            </button>
          </div>
        </div>
      </div>
    );
  }

  const displayError = ERROR_MESSAGES[error] || error;

  return (
    <div className="auth-page">
      <div className="auth-theme-toggle">
        <ThemeToggle />
      </div>
      <div className="auth-card">
        <h1>Set new password</h1>
        <p className="subtitle">Enter your new password below.</p>

        <form className="form" onSubmit={handleSubmit}>
          <input
            className="input"
            type="password"
            placeholder="New password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            disabled={loading || !token}
          />
          <input
            className="input"
            type="password"
            placeholder="Confirm new password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            minLength={8}
            disabled={loading || !token}
          />

          {displayError && <div className="error">{displayError}</div>}

          <button
            className="btn primary"
            type="submit"
            disabled={loading || !token || !password || !confirmPassword}
            style={{ width: '100%' }}
          >
            {loading ? 'Resetting...' : 'Reset password'}
          </button>
        </form>

        <div className="auth-footer">
          <a href="/login">Back to login</a>
        </div>
      </div>
    </div>
  );
}
