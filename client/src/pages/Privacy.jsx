// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

import React from 'react';
import { Link } from 'react-router-dom';
import ThemeToggle from '../components/ThemeToggle.jsx';

export default function Privacy() {
  return (
    <div className="legal-page">
      <div className="legal-theme-toggle">
        <ThemeToggle />
      </div>
      <div className="legal-card">
        <div className="legal-header">
          <h1>Privacy Policy</h1>
          <span className="legal-date">Last updated: December 19, 2025</span>
        </div>

        <div className="legal-content">
          <p>
            This Privacy Policy explains what information boop.cat (the "Service") collects, how we use it, and how it
            may be processed by third-party infrastructure providers.
          </p>

          <h2>1. Information we collect</h2>
          <p>The Service may collect account and usage information such as:</p>
          <ul>
            <li>Email address and username (for authentication and account management).</li>
            <li>OAuth identifiers (when you sign in with a third‑party provider).</li>
            <li>IP addresses are hashed and stored for security and abuse prevention.</li>
            <li>Basic request metadata (e.g., User-Agent) for rate limiting.</li>
            <li>Repository URLs and deployment configuration you provide to create deployments.</li>
            <li>Deployment logs generated during build and upload (to help you debug deployments).</li>
            <li>Files you deploy (static site assets) and related identifiers needed to serve your site.</li>
          </ul>

          <h2>2. How we use information</h2>
          <p>
            We use information to operate the Service, authenticate users, secure the platform, prevent abuse, and
            support deployments.
          </p>

          <h2>3. Where your deployed sites are stored and served</h2>
          <p>
            When you deploy a site, your static build output is uploaded to Backblaze B2 object storage and served
            globally via Cloudflare. Cloudflare also stores limited routing metadata in Cloudflare Workers KV to map
            your subdomain to your current deployment.
          </p>

          <h2>4. Cookies and authentication</h2>
          <p>
            The Service uses an HTTP-only session cookie to keep you signed in. Session data may be stored on our
            servers for account authentication and security.
          </p>

          <h2>5. Sharing</h2>
          <p>
            We do not sell your personal information. We may share information with service providers as needed to
            operate the Service (for example, email delivery), or when required by law.
          </p>

          <p>
            Our infrastructure providers (acting as service providers/processors) may process data necessary to deliver
            the Service, including:
          </p>
          <ul>
            <li>Cloudflare (CDN/Workers/KV) for request handling, caching, routing, and delivery of deployed sites.</li>
            <li>Backblaze B2 for storage of deployed static site files.</li>
            <li>
              Hetzner for hosting the main Service application servers (dashboard/API), including build and deployment
              orchestration.
            </li>
          </ul>

          <h2>6. Data retention</h2>
          <p>
            We retain account and deployment data for as long as necessary to provide the Service and for security and
            operational purposes. When you delete a project, we attempt to remove its associated deployments, deployed
            files, and routing metadata.
          </p>

          <h2>7. Security</h2>
          <p>We take reasonable measures to protect data, but no method of transmission or storage is 100% secure.</p>

          <h2>8. Contact</h2>
          <p>
            If you have questions about this policy, contact <a href="mailto:hello@boop.cat">hello@boop.cat</a>.
          </p>
        </div>

        <div className="legal-footer">
          <Link to="/">← Back to home</Link>
        </div>
      </div>
    </div>
  );
}
