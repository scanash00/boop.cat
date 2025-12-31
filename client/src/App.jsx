import React from 'react';
import { Routes, Route, Link, useNavigate } from 'react-router-dom';
import Login from './pages/Login.jsx';
import Signup from './pages/Signup.jsx';
import ResetPassword from './pages/ResetPassword.jsx';
import DashboardLayout from './pages/DashboardLayout.jsx';
import DashboardHome from './pages/DashboardHome.jsx';
import DashboardSite from './pages/DashboardSite.jsx';
import NewSite from './pages/NewSite.jsx';
import Account from './pages/Account.jsx';
import Tos from './pages/Tos.jsx';
import Privacy from './pages/Privacy.jsx';
import Dmca from './pages/Dmca.jsx';
import ApiDocs from './pages/ApiDocs.jsx';
import ThemeToggle from './components/ThemeToggle.jsx';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error) {
    return { error };
  }

  render() {
    if (this.state.error) {
      return (
        <div className="center">
          <div style={{ fontSize: 64, marginBottom: 16 }}>😵</div>
          <h1 className="title" style={{ fontSize: 36 }}>
            Something went wrong
          </h1>
          <div className="muted" style={{ maxWidth: 480, marginBottom: 24 }}>
            {String(this.state.error?.message || this.state.error)}
          </div>
          <div className="buttons">
            <button className="btn primary" onClick={() => window.location.reload()}>
              Refresh page
            </button>
            <a className="btn ghost" href="/">
              Go home
            </a>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}

function Landing() {
  const [authChecked, setAuthChecked] = React.useState(false);
  const [isAuthed, setIsAuthed] = React.useState(false);
  const [navHidden, setNavHidden] = React.useState(false);
  const lastScrollY = React.useRef(0);

  React.useEffect(() => {
    fetch('/api/auth/me', { credentials: 'same-origin' })
      .then((r) => r.json())
      .then((d) => {
        setIsAuthed(Boolean(d?.authenticated));
        setAuthChecked(true);
      })
      .catch(() => {
        setIsAuthed(false);
        setAuthChecked(true);
      });
  }, []);

  React.useEffect(() => {
    const handleScroll = () => {
      const currentScrollY = window.scrollY;
      if (currentScrollY > lastScrollY.current && currentScrollY > 80) {
        setNavHidden(true);
      } else {
        setNavHidden(false);
      }
      lastScrollY.current = currentScrollY;
    };
    window.addEventListener('scroll', handleScroll, { passive: true });
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  return (
    <div className="landing-page">
      {}
      <nav className={`navbar-frame ${navHidden ? 'nav-hidden' : ''}`}>
        <div className="navbar-content">
          <Link to="/" className="navbar-logo">
            <img src="/milly.png" alt="" width="28" height="28" style={{ imageRendering: 'pixelated' }} />
            <span>boop.cat</span>
          </Link>
          <div className="navbar-buttons">
            <ThemeToggle />
            {authChecked && isAuthed ? (
              <Link to="/dashboard" className="glass-btn">
                <svg
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <rect x="3" y="3" width="7" height="7" />
                  <rect x="14" y="3" width="7" height="7" />
                  <rect x="14" y="14" width="7" height="7" />
                  <rect x="3" y="14" width="7" height="7" />
                </svg>
                <span>Dashboard</span>
              </Link>
            ) : (
              <>
                <Link to="/login" className="glass-btn">
                  <svg
                    width="18"
                    height="18"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" />
                    <polyline points="10 17 15 12 10 7" />
                    <line x1="15" y1="12" x2="3" y2="12" />
                  </svg>
                  <span>Log in</span>
                </Link>
                <Link to="/signup" className="glass-btn accent">
                  <svg
                    width="18"
                    height="18"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                    <circle cx="8.5" cy="7" r="4" />
                    <line x1="20" y1="8" x2="20" y2="14" />
                    <line x1="23" y1="11" x2="17" y2="11" />
                  </svg>
                  <span>Sign up</span>
                </Link>
              </>
            )}
          </div>
        </div>
      </nav>

      {}
      <main className="page-frame">
        {}
        <section className="hero-section">
          <div className="hero-icon">
            <img src="/milly.png" alt="" width="80" height="80" style={{ imageRendering: 'pixelated' }} />
          </div>
          <h1>Static hosting, simplified</h1>
          <p className="hero-desc">
            Push your code. We handle the rest. Free static site hosting powered by open source.
          </p>
          <div className="hero-buttons">
            {authChecked && isAuthed ? (
              <Link to="/dashboard" className="glass-btn accent large">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="3" y="3" width="7" height="7" />
                  <rect x="14" y="3" width="7" height="7" />
                  <rect x="14" y="14" width="7" height="7" />
                  <rect x="3" y="14" width="7" height="7" />
                </svg>
                <span>Go to Dashboard</span>
              </Link>
            ) : (
              <>
                <Link to="/signup" className="glass-btn accent large">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M5 12h14" />
                    <path d="M12 5v14" />
                  </svg>
                  <span>Start for free</span>
                </Link>
                <Link to="/login" className="glass-btn large">
                  <span>I have an account</span>
                </Link>
              </>
            )}
          </div>
        </section>

        {}
        <section className="content-section">
          <div className="features-row">
            <div className="block-card" style={{ '--card-color': '#B8D4E8' }}>
              <div className="card-icon-wrap">
                <svg
                  width="28"
                  height="28"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
                </svg>
              </div>
              <h3>Instant deploys</h3>
              <p>Connect your Git repo and every push goes live automatically.</p>
            </div>

            <div className="block-card" style={{ '--card-color': '#D4E8B8' }}>
              <div className="card-icon-wrap">
                <svg
                  width="28"
                  height="28"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
                  <line x1="8" y1="21" x2="16" y2="21" />
                  <line x1="12" y1="17" x2="12" y2="21" />
                </svg>
              </div>
              <h3>Any framework</h3>
              <p>Vite, Next.js, Astro, Hugo, if it outputs HTML, it works.</p>
            </div>

            <div className="block-card" style={{ '--card-color': '#E8D4B8' }}>
              <div className="card-icon-wrap">
                <svg
                  width="28"
                  height="28"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                  <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                </svg>
              </div>
              <h3>SSL included</h3>
              <p>Every site gets HTTPS. No configuration needed.</p>
            </div>

            <div className="block-card" style={{ '--card-color': '#E8B8D4' }}>
              <div className="card-icon-wrap">
                <svg
                  width="28"
                  height="28"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <circle cx="12" cy="12" r="10" />
                  <line x1="2" y1="12" x2="22" y2="12" />
                  <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                </svg>
              </div>
              <h3>Edge delivery</h3>
              <p>Your sites cached globally for fast loads everywhere.</p>
            </div>
          </div>
        </section>

        {}
        <section className="content-section">
          <h2 className="section-title">How it works</h2>
          <div className="steps-row">
            <div className="step-card" style={{ '--card-color': '#C8E0F0' }}>
              <span className="step-num">1</span>
              <h3>Add your repo</h3>
              <p>Paste a GitHub, GitLab, or any public Git URL</p>
            </div>
            <div className="step-arrow">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M5 12h14M12 5l7 7-7 7" />
              </svg>
            </div>
            <div className="step-card" style={{ '--card-color': '#D0F0C8' }}>
              <span className="step-num">2</span>
              <h3>Set build options</h3>
              <p>Choose your build command and output folder</p>
            </div>
            <div className="step-arrow">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M5 12h14M12 5l7 7-7 7" />
              </svg>
            </div>
            <div className="step-card" style={{ '--card-color': '#F0D0C8' }}>
              <span className="step-num">3</span>
              <h3>You're live</h3>
              <p>Get a URL instantly, add custom domains later</p>
            </div>
          </div>
        </section>

        {}
        <section className="content-section">
          <h2 className="section-title">Works with your stack</h2>
          <div className="frameworks-row">
            {[
              { name: 'React', color: '#61DAFB20' },
              { name: 'Vue', color: '#42b88320' },
              { name: 'Svelte', color: '#FF3E0020' },
              { name: 'Astro', color: '#FF5D0120' },
              { name: 'Next.js', color: '#00000015' },
              { name: 'Vite', color: '#646CFF20' }
            ].map((fw) => (
              <div key={fw.name} className="framework-chip" style={{ background: fw.color }}>
                {fw.name}
              </div>
            ))}
          </div>
        </section>
      </main>

      {}
      <footer className="site-footer">
        <div className="footer-inner">
          <span>© 2025 boop.cat</span>
          <div className="footer-links">
            <a
              href="https://tangled.org/scanash.com/boop.cat/"
              target="_blank"
              rel="noopener noreferrer"
              className="tangled-footer-link"
            >
              <img
                src="https://assets.tangled.network/tangled_dolly_face_only_black_on_trans.svg"
                className="tangled-icon light-only"
                alt="Tangled"
              />
              <img
                src="https://assets.tangled.network/tangled_dolly_face_only_white_on_trans.svg"
                className="tangled-icon dark-only"
                alt="Tangled"
              />
              <span>Tangled</span>
            </a>
            <Link to="/tos">Terms</Link>
            <Link to="/privacy">Privacy</Link>
            <Link to="/dmca">DMCA</Link>
            <a href="https://ko-fi.com/P5P1VMR1D" target="_blank" rel="noopener noreferrer">
              Donate
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default function App() {
  return (
    <ErrorBoundary>
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="/tos" element={<Tos />} />
        <Route path="/privacy" element={<Privacy />} />
        <Route path="/dmca" element={<Dmca />} />

        <Route path="/dashboard" element={<DashboardLayout />}>
          <Route index element={<DashboardHome />} />
          <Route path="new" element={<NewSite />} />
          <Route path="site/:id" element={<DashboardSite />} />
          <Route path="account" element={<Account />} />
          <Route path="api-docs" element={<ApiDocs />} />
        </Route>
      </Routes>
    </ErrorBoundary>
  );
}
