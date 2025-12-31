import React from 'react';
import { Link } from 'react-router-dom';
import ThemeToggle from '../components/ThemeToggle.jsx';

export default function Dmca() {
  return (
    <div className="legal-page">
      <div className="legal-theme-toggle">
        <ThemeToggle />
      </div>
      <div className="legal-card">
        <div className="legal-header">
          <h1>DMCA Policy</h1>
          <span className="legal-date">Last updated: December 19, 2025</span>
        </div>

        <div className="legal-content">
          <p>
            boop.cat respects the intellectual property rights of others and expects its users to do the same. In
            accordance with the Digital Millennium Copyright Act of 1998, the text of which may be found on the U.S.
            Copyright Office website at{' '}
            <a href="http://www.copyright.gov/legislation/dmca.pdf" target="_blank" rel="noopener noreferrer">
              http:
            </a>
            , we will respond expeditiously to claims of copyright infringement committed using the Service.
          </p>

          <h2>Reporting Copyright Infringement</h2>
          <p>
            If you are a copyright owner, or are authorized to act on behalf of one, or authorized to act under any
            exclusive right under copyright, please report alleged copyright infringements taking place on or through
            the Service by sending a notice to our designated agent at:
          </p>

          <div className="legal-callout">
            <strong>Email:</strong> <a href="mailto:dmca@boop.cat">dmca@boop.cat</a>
          </div>

          <p>
            Upon receipt of the Notice as described below, we will take whatever action, in our sole discretion, we deem
            appropriate, including removal of the challenged material from the Site.
          </p>

          <h2>What to include</h2>
          <p>Please include the following in your notice:</p>
          <ul>
            <li>Identify the copyrighted work that you claim has been infringed.</li>
            <li>
              Identify the material that you claim is infringing (or to be the subject of infringing activity) and that
              is to be removed or access to which is to be disabled, and information reasonably sufficient to permit us
              to locate the material (e.g., the URL).
            </li>
            <li>Your contact information (email, address, phone number).</li>
            <li>
              A statement that you have a good faith belief that use of the material in the manner complained of is not
              authorized by the copyright owner, its agent, or the law.
            </li>
          </ul>
        </div>

        <div className="legal-footer">
          <Link to="/">← Back to home</Link>
        </div>
      </div>
    </div>
  );
}
