import React from 'react';
import { Link } from 'react-router-dom';
import ThemeToggle from '../components/ThemeToggle.jsx';

export default function Tos() {
  return (
    <div className="legal-page">
      <div className="legal-theme-toggle">
        <ThemeToggle />
      </div>
      <div className="legal-card">
        <div className="legal-header">
          <h1>Terms of Service</h1>
          <span className="legal-date">Last updated: December 18, 2025</span>
        </div>

        <div className="legal-content">
          <p>
            These Terms of Service ("Terms") govern your use of boop.cat (the "Service"). By using the Service, you
            agree to these Terms.
          </p>

          <h2>1. Accounts</h2>
          <p>
            You are responsible for maintaining the security of your account and for all activity that occurs under your
            account.
          </p>

          <h2>2. Acceptable use</h2>
          <p>
            You agree not to use the Service to host, deploy, distribute, or link to content that is illegal, malicious,
            infringes intellectual property, violates privacy, or is intended to harm or disrupt systems (including
            malware, phishing, or abuse of third‑party services).
          </p>

          <h2>3. Deployments and content</h2>
          <p>
            You retain responsibility for the repositories you connect and the content you deploy. We may suspend or
            remove deployments that violate these Terms or applicable law.
          </p>

          <h2>4. Rate limits and availability</h2>
          <p>
            The Service may enforce rate limits, quotas, and other restrictions. The Service is provided on an "as is"
            and "as available" basis and may change or be discontinued.
          </p>

          <h2>5. Termination</h2>
          <p>
            We may suspend or terminate access to the Service at any time for violations of these Terms, suspected
            abuse, or security reasons.
          </p>

          <h2>6. Disclaimer</h2>
          <p>
            To the maximum extent permitted by law, we disclaim all warranties and will not be liable for any indirect,
            incidental, or consequential damages arising from your use of the Service.
          </p>

          <h2>7. Contact</h2>
          <p>
            Questions about these Terms can be directed to <a href="mailto:hello@boop.cat">hello@boop.cat</a>.
          </p>
        </div>

        <div className="legal-footer">
          <Link to="/">← Back to home</Link>
        </div>
      </div>
    </div>
  );
}
