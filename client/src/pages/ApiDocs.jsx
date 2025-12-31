// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

import React from 'react';
import { useNavigate, useOutletContext } from 'react-router-dom';
import { Book, Key, Terminal, Copy, Check, ExternalLink } from 'lucide-react';

export default function ApiDocs() {
  const { me } = useOutletContext();
  const navigate = useNavigate();
  const [copiedIndex, setCopiedIndex] = React.useState(null);

  const baseUrl = window.location.origin;

  const copyToClipboard = (text, index) => {
    navigator.clipboard.writeText(text);
    setCopiedIndex(index);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  const endpoints = [
    {
      method: 'GET',
      path: '/api/v1/sites',
      description: 'List all your sites',
      example: `curl -H "Authorization: Bearer YOUR_API_KEY" \\
  ${baseUrl}/api/v1/sites`,
      response: `{
  "sites": [
    {
      "id": "abc123",
      "name": "my-site",
      "domain": "my-site.boop.cat",
      "git": { "url": "https://github.com/user/repo", "branch": "main" },
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ]
}`
    },
    {
      method: 'GET',
      path: '/api/v1/sites/{id}',
      description: 'Get details for a specific site',
      example: `curl -H "Authorization: Bearer YOUR_API_KEY" \\
  ${baseUrl}/api/v1/sites/YOUR_SITE_ID`,
      response: `{
  "id": "abc123",
  "name": "my-site",
  "domain": "my-site.boop.cat",
  "git": { "url": "https://github.com/user/repo", "branch": "main" },
  "buildCommand": "npm run build",
  "outputDir": "dist",
  "createdAt": "2024-01-01T00:00:00Z"
}`
    },
    {
      method: 'POST',
      path: '/api/v1/sites/{id}/deploy',
      description: 'Trigger a new deployment for a site. Returns immediately with deployment details.',
      example: `curl -X POST -H "Authorization: Bearer YOUR_API_KEY" \\
  ${baseUrl}/api/v1/sites/YOUR_SITE_ID/deploy`,
      response: `{
  "id": "dep_xyz789",
  "siteId": "abc123",
  "status": "building",
  "createdAt": "2024-01-01T12:00:00Z",
  "commitSha": "abc1234",
  "commitMessage": "Initial commit",
  "commitAuthor": "User"
}`
    },
    {
      method: 'POST',
      path: '/api/v1/sites/{id}/deploy?wait=true',
      description: 'Trigger a deployment and stream build logs in real-time.',
      example: `curl -X POST -H "Authorization: Bearer YOUR_API_KEY" \\
  ${baseUrl}/api/v1/sites/YOUR_SITE_ID/deploy?wait=true`,
      response: `[2024-01-01T12:00:00Z] Cloning repository...
[2024-01-01T12:00:02Z] Installing dependencies...
[2024-01-01T12:00:10Z] Building project...
[2024-01-01T12:00:20Z] Uploading to edge storage...
[2024-01-01T12:00:22Z] Deployment successful!`
    },
    {
      method: 'GET',
      path: '/api/v1/sites/{id}/deployments',
      description: 'List all deployments for a site',
      example: `curl -H "Authorization: Bearer YOUR_API_KEY" \\
  ${baseUrl}/api/v1/sites/YOUR_SITE_ID/deployments`,
      response: `{
  "deployments": [
    {
      "id": "dep_xyz789",
      "siteId": "abc123",
      "status": "active",
      "createdAt": "2024-01-01T12:00:00Z",
      "url": "https://my-site.boop.cat"
    }
  ]
}`
    }
  ];

  return (
    <div className="page">
      <div className="pageHeader">
        <div>
          <div className="h">API Documentation</div>
          <div className="muted">Use the API to deploy sites programmatically from CI/CD pipelines.</div>
        </div>
      </div>

      {}
      <div className="panel">
        <div className="panelTitle">
          <Book size={18} style={{ marginRight: 8, color: '#e88978' }} />
          Overview
        </div>
        <div className="muted" style={{ marginBottom: 16, lineHeight: 1.6 }}>
          The boop.cat API allows you to manage sites and trigger deployments programmatically. This is useful for CI/CD
          pipelines, GitHub Actions, and custom deployment workflows.
        </div>
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
          <div
            style={{
              padding: '8px 16px',
              background: 'var(--bg-secondary)',
              borderRadius: 8,
              fontSize: 13
            }}
          >
            <strong>Base URL:</strong> <code>{baseUrl}/api/v1</code>
          </div>
          <div
            style={{
              padding: '8px 16px',
              background: 'var(--bg-secondary)',
              borderRadius: 8,
              fontSize: 13
            }}
          >
            <strong>Rate Limit:</strong> 100 requests / 15 minutes
          </div>
        </div>
      </div>

      {}
      <div className="panel" style={{ marginTop: 8 }}>
        <div className="panelTitle">
          <Key size={18} style={{ marginRight: 8, color: '#e88978' }} />
          Authentication
        </div>
        <div className="muted" style={{ marginBottom: 16, lineHeight: 1.6 }}>
          All API requests require authentication using an API key. Create API keys in your{' '}
          <a
            href="#"
            onClick={(e) => {
              e.preventDefault();
              navigate('/dashboard/account');
            }}
            style={{ color: 'var(--primary)' }}
          >
            Account Settings
          </a>
          .
        </div>

        <div style={{ marginBottom: 16 }}>
          <div style={{ fontWeight: 600, marginBottom: 8, fontSize: 14 }}>Request Header</div>
          <div
            style={{
              background: 'var(--bg-tertiary, #1e1e1e)',
              padding: 16,
              borderRadius: 8,
              fontFamily: 'monospace',
              fontSize: 13,
              color: 'var(--code-text, #e0e0e0)',
              overflowX: 'auto'
            }}
          >
            Authorization: Bearer sk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
          </div>
        </div>

        <div className="notice" style={{ marginBottom: 0 }}>
          <strong>Important:</strong> API keys have full access to your account. Keep them secret and never expose them
          in client-side code.
        </div>
      </div>

      {}
      <div className="panel" style={{ marginTop: 8 }}>
        <div className="panelTitle">
          <Terminal size={18} style={{ marginRight: 8, color: '#e88978' }} />
          Endpoints
        </div>

        {endpoints.map((endpoint, i) => (
          <div
            key={i}
            style={{
              borderBottom: i < endpoints.length - 1 ? '1px solid var(--border)' : 'none',
              paddingBottom: i < endpoints.length - 1 ? 24 : 0,
              marginBottom: i < endpoints.length - 1 ? 24 : 0
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
              <span
                style={{
                  padding: '4px 10px',
                  borderRadius: 6,
                  fontSize: 12,
                  fontWeight: 700,
                  fontFamily: 'monospace',
                  background: endpoint.method === 'POST' ? 'rgba(34, 197, 94, 0.15)' : 'rgba(59, 130, 246, 0.15)',
                  color: endpoint.method === 'POST' ? '#22c55e' : '#3b82f6'
                }}
              >
                {endpoint.method}
              </span>
              <code style={{ fontSize: 14, fontWeight: 500 }}>{endpoint.path}</code>
            </div>
            <div className="muted" style={{ marginBottom: 12 }}>
              {endpoint.description}
            </div>

            <div style={{ marginBottom: 12 }}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  marginBottom: 6
                }}
              >
                <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}>Example Request</span>
                <button
                  className="iconBtn"
                  style={{
                    width: 28,
                    height: 28,
                    borderRadius: '50%',
                    background: 'var(--bg-secondary)',
                    border: 'none'
                  }}
                  onClick={() => copyToClipboard(endpoint.example, `example-${i}`)}
                >
                  {copiedIndex === `example-${i}` ? <Check size={14} /> : <Copy size={14} />}
                </button>
              </div>
              <pre
                style={{
                  background: 'var(--bg-tertiary, #1e1e1e)',
                  padding: 16,
                  borderRadius: 8,
                  fontFamily: 'monospace',
                  fontSize: 12,
                  color: 'var(--code-text, #e0e0e0)',
                  overflowX: 'auto',
                  margin: 0,
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word'
                }}
              >
                {endpoint.example}
              </pre>
            </div>

            <div>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  marginBottom: 6
                }}
              >
                <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}>Example Response</span>
                <button
                  className="iconBtn"
                  style={{
                    width: 28,
                    height: 28,
                    borderRadius: '50%',
                    background: 'var(--bg-secondary)',
                    border: 'none'
                  }}
                  onClick={() => copyToClipboard(endpoint.response, `response-${i}`)}
                >
                  {copiedIndex === `response-${i}` ? <Check size={14} /> : <Copy size={14} />}
                </button>
              </div>
              <pre
                style={{
                  background: 'var(--bg-tertiary, #1e1e1e)',
                  padding: 16,
                  borderRadius: 8,
                  fontFamily: 'monospace',
                  fontSize: 12,
                  color: 'var(--code-text, #e0e0e0)',
                  overflowX: 'auto',
                  margin: 0
                }}
              >
                {endpoint.response}
              </pre>
            </div>
          </div>
        ))}
      </div>

      {}
      <div className="panel" style={{ marginTop: 8 }}>
        <div className="panelTitle">
          <ExternalLink size={18} style={{ marginRight: 8, color: '#e88978' }} />
          GitHub Actions Example
        </div>
        <div className="muted" style={{ marginBottom: 16, lineHeight: 1.6 }}>
          Here's an example workflow to trigger a deployment on every push to main:
        </div>

        <div style={{ position: 'relative' }}>
          <button
            className="iconBtn"
            style={{
              position: 'absolute',
              top: 8,
              right: 8,
              width: 28,
              height: 28,
              borderRadius: '50%',
              background: 'var(--bg-secondary)',
              border: 'none'
            }}
            onClick={() =>
              copyToClipboard(
                `name: Deploy to boop.cat

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Deployment
        run: |
          curl -X POST \\
            -H "Authorization: Bearer \${{ secrets.BOOP_CAT_API_KEY }}" \\
            ${baseUrl}/api/v1/sites/YOUR_SITE_ID/deploy?wait=true`,
                'github-actions'
              )
            }
          >
            {copiedIndex === 'github-actions' ? <Check size={14} /> : <Copy size={14} />}
          </button>
          <pre
            style={{
              background: 'var(--bg-tertiary, #1e1e1e)',
              padding: 16,
              borderRadius: 8,
              fontFamily: 'monospace',
              fontSize: 12,
              color: 'var(--code-text, #e0e0e0)',
              overflowX: 'auto',
              margin: 0
            }}
          >
            {`name: Deploy to boop.cat

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Deployment
        run: |
          curl -X POST \\
            -H "Authorization: Bearer \${{ secrets.BOOP_CAT_API_KEY }}" \\
            ${baseUrl}/api/v1/sites/YOUR_SITE_ID/deploy?wait=true`}
          </pre>
        </div>

        <div className="notice" style={{ marginTop: 16, marginBottom: 0 }}>
          <strong>Tip:</strong> Store your API key as a repository secret named <code>BOOP_CAT_API_KEY</code> in your
          GitHub repository settings.
        </div>
      </div>

      {}
      <div className="panel" style={{ marginTop: 8 }}>
        <div className="panelTitle">
          <ExternalLink size={18} style={{ marginRight: 8, color: '#fc6d26' }} />
          GitLab CI Example
        </div>
        <div className="muted" style={{ marginBottom: 16, lineHeight: 1.6 }}>
          For GitLab CI/CD, add this to your <code>.gitlab-ci.yml</code>:
        </div>

        <div style={{ position: 'relative' }}>
          <button
            className="iconBtn"
            style={{
              position: 'absolute',
              top: 8,
              right: 8,
              width: 28,
              height: 28,
              borderRadius: '50%',
              background: 'var(--bg-secondary)',
              border: 'none'
            }}
            onClick={() =>
              copyToClipboard(
                `deploy:
  stage: deploy
  image: curlimages/curl
  script:
    - curl -X POST -H "Authorization: Bearer $BOOP_CAT_API_KEY" ${baseUrl}/api/v1/sites/YOUR_SITE_ID/deploy?wait=true
  only:
    - main`,
                'gitlab-ci'
              )
            }
          >
            {copiedIndex === 'gitlab-ci' ? <Check size={14} /> : <Copy size={14} />}
          </button>
          <pre
            style={{
              background: 'var(--bg-tertiary, #1e1e1e)',
              padding: 16,
              borderRadius: 8,
              fontFamily: 'monospace',
              fontSize: 12,
              color: 'var(--code-text, #e0e0e0)',
              overflowX: 'auto',
              margin: 0
            }}
          >
            {`deploy:
  stage: deploy
  image: curlimages/curl
  script:
    - curl -X POST -H "Authorization: Bearer $BOOP_CAT_API_KEY" ${baseUrl}/api/v1/sites/YOUR_SITE_ID/deploy?wait=true
  only:
    - main`}
          </pre>
        </div>

        <div className="notice" style={{ marginTop: 16, marginBottom: 0 }}>
          <strong>Tip:</strong> Add <code>BOOP_CAT_API_KEY</code> as a variable in{' '}
          <strong>
            Settings {'>'} CI/CD {'>'} Variables.
          </strong>
        </div>
      </div>

      {}
      <div className="panel" style={{ marginTop: 8 }}>
        <div className="panelTitle">
          <Terminal size={18} style={{ marginRight: 8, color: '#68a063' }} />
          Node.js Script
        </div>
        <div className="muted" style={{ marginBottom: 16, lineHeight: 1.6 }}>
          A simple Node.js script to trigger deployments:
        </div>

        <div style={{ position: 'relative' }}>
          <button
            className="iconBtn"
            style={{
              position: 'absolute',
              top: 8,
              right: 8,
              width: 28,
              height: 28,
              borderRadius: '50%',
              background: 'var(--bg-secondary)',
              border: 'none'
            }}
            onClick={() =>
              copyToClipboard(
                `const siteId = 'YOUR_SITE_ID';
const apiKey = process.env.BOOP_CAT_API_KEY;

async function deploy() {
  const res = await fetch(\`${baseUrl}/api/v1/sites/\${siteId}/deploy\`, {
    method: 'POST',
    headers: {
      'Authorization': \`Bearer \${apiKey}\`
    }
  });
  
  const data = await res.json();
  console.log('Deploy triggered:', data);
}

deploy().catch(console.error);`,
                'node-script'
              )
            }
          >
            {copiedIndex === 'node-script' ? <Check size={14} /> : <Copy size={14} />}
          </button>
          <pre
            style={{
              background: 'var(--bg-tertiary, #1e1e1e)',
              padding: 16,
              borderRadius: 8,
              fontFamily: 'monospace',
              fontSize: 12,
              color: 'var(--code-text, #e0e0e0)',
              overflowX: 'auto',
              margin: 0
            }}
          >
            {`const siteId = 'YOUR_SITE_ID';
const apiKey = process.env.BOOP_CAT_API_KEY;

async function deploy() {
  const res = await fetch(\`${baseUrl}/api/v1/sites/\${siteId}/deploy\`, {
    method: 'POST',
    headers: {
      'Authorization': \`Bearer \${apiKey}\`
    }
  });
  
  const data = await res.json();
  console.log('Deploy triggered:', data);
}

deploy().catch(console.error);`}
          </pre>
        </div>
      </div>

      {}
      <div className="panel" style={{ marginTop: 8 }}>
        <div className="panelTitle">
          <ExternalLink size={18} style={{ marginRight: 8, color: '#8b5cf6' }} />
          Tangled CI Example
        </div>
        <div className="muted" style={{ marginBottom: 16, lineHeight: 1.6 }}>
          For Tangled.org (Spindle), create only <code>.tangled/workflows/deploy.yml</code>:
        </div>

        <div style={{ position: 'relative' }}>
          <button
            className="iconBtn"
            style={{
              position: 'absolute',
              top: 8,
              right: 8,
              width: 28,
              height: 28,
              borderRadius: '50%',
              background: 'var(--bg-secondary)',
              border: 'none'
            }}
            onClick={() =>
              copyToClipboard(
                `when:
  - event: ["push"]
    branch: ["main"]

engine: "nixery"

clone:
  skip: true

dependencies:
  nixpkgs:
    - curl

steps:
  - name: "Trigger Deploy"
    command: |
      curl -X POST \\
        -H "Authorization: Bearer $BOOP_CAT_API_KEY" \\
        ${baseUrl}/api/v1/sites/YOUR_SITE_ID/deploy?wait=true`,
                'tangled-ci'
              )
            }
          >
            {copiedIndex === 'tangled-ci' ? <Check size={14} /> : <Copy size={14} />}
          </button>
          <pre
            style={{
              background: 'var(--bg-tertiary, #1e1e1e)',
              padding: 16,
              borderRadius: 8,
              fontFamily: 'monospace',
              fontSize: 12,
              color: 'var(--code-text, #e0e0e0)',
              overflowX: 'auto',
              margin: 0
            }}
          >
            {`when:
  - event: ["push"]
    branch: ["main"]

engine: "nixery"

clone:
  skip: true

dependencies:
  nixpkgs:
    - curl

steps:
  - name: "Trigger Deploy"
    command: |
      curl -X POST \\
        -H "Authorization: Bearer $BOOP_CAT_API_KEY" \\
        ${baseUrl}/api/v1/sites/YOUR_SITE_ID/deploy?wait=true`}
          </pre>
          <div className="notice" style={{ marginTop: 16, marginBottom: 0 }}>
            <strong>Tip:</strong> Add <code>BOOP_CAT_API_KEY</code> as a secret in{' '}
            <strong>
              settings {'>'} pipelines {'>'} secrets.
            </strong>
          </div>
        </div>
      </div>
    </div>
  );
}
