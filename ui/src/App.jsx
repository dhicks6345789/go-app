import React, { useState, useEffect } from 'react';

export default function App() {
  const [user, setUser] = useState({ username: 'Loading...', auth_type: 'checking', mode: 'desktop' });
  const [sysInfo, setSysInfo] = useState(null);
  const [items, setItems] = useState([]);
  const [newItemName, setNewItemName] = useState('');
  const [apiResponse, setApiResponse] = useState('Click an API test button to view response...');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchUser();
    fetchSystemInfo();
    fetchItems();
  }, []);

  const fetchUser = async () => {
    try {
      const res = await fetch('/api/v1/user');
      const data = await res.json();
      setUser(data);
    } catch (err) {
      setUser({ username: 'Offline / Disconnected', auth_type: 'none', mode: 'unknown' });
    }
  };

  const fetchSystemInfo = async () => {
    try {
      const res = await fetch('/api/v1/info');
      const data = await res.json();
      setSysInfo(data);
    } catch (err) {
      console.error('Failed to fetch sys info:', err);
    }
  };

  const fetchItems = async () => {
    try {
      const res = await fetch('/api/v1/items');
      const data = await res.json();
      setItems(data.items || []);
    } catch (err) {
      console.error('Failed to fetch items:', err);
    }
  };

  const handleAddItem = async (e) => {
    e.preventDefault();
    if (!newItemName.trim()) return;
    try {
      const res = await fetch('/api/v1/items', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newItemName }),
      });
      const data = await res.json();
      setNewItemName('');
      fetchItems();
      setApiResponse(JSON.stringify(data, null, 2));
    } catch (err) {
      setApiResponse(`Error adding item: ${err.message}`);
    }
  };

  const testEndpoint = async (endpoint) => {
    setLoading(true);
    try {
      const res = await fetch(endpoint);
      const data = await res.json();
      setApiResponse(JSON.stringify(data, null, 2));
    } catch (err) {
      setApiResponse(`Error calling ${endpoint}: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app-container">
      {/* Header */}
      <header className="header">
        <div className="brand">
          <div className="brand-icon">⚡</div>
          <div className="brand-title">
            <h1>Go Self-Contained App</h1>
            <p>Single Executable • Embedded React UI • OpenAPI Docs</p>
          </div>
        </div>
        <div className="user-badge">
          <div className="status-dot"></div>
          <div className="user-info">
            <span className="user-name">👤 {user.username}</span>
            <span className="user-mode">{user.mode} Mode ({user.auth_type})</span>
          </div>
        </div>
      </header>

      {/* Main Grid */}
      <div className="dashboard-grid">
        {/* Hero Section */}
        <div className="card hero-card col-12">
          <div className="hero-content">
            <h2>Welcome to your Go Application Framework</h2>
            <p>
              This self-contained executable bundles backend API endpoints, a modern React single-page frontend,
              and interactive OpenAPI documentation all into a single binary. Works completely offline!
            </p>
            <div className="actions-row">
              <a href="/docs/api" target="_blank" rel="noreferrer" className="btn btn-primary">
                📖 Open OpenAPI Docs (/docs/api)
              </a>
              <button onClick={() => testEndpoint('/api/v1/user')} className="btn btn-secondary">
                🔍 Verify Auth (/api/v1/user)
              </button>
              <button onClick={() => testEndpoint('/api/v1/health')} className="btn btn-secondary">
                💚 Health Check (/api/v1/health)
              </button>
            </div>
          </div>
        </div>

        {/* System Information */}
        <div className="card col-4">
          <div className="card-header">
            <h3 className="card-title">🖥️ System Status</h3>
          </div>
          <div className="status-list">
            <div className="status-item">
              <span className="status-label">Operation Mode</span>
              <span className="status-value">{sysInfo ? sysInfo.mode : 'Loading...'}</span>
            </div>
            <div className="status-item">
              <span className="status-label">Authenticated User</span>
              <span className="status-value">{user.username}</span>
            </div>
            <div className="status-item">
              <span className="status-label">Auth Provider</span>
              <span className="status-value">{user.auth_type}</span>
            </div>
            <div className="status-item">
              <span className="status-label">Go Version</span>
              <span className="status-value">{sysInfo ? sysInfo.go_version : '...'}</span>
            </div>
            <div className="status-item">
              <span className="status-label">OS / Architecture</span>
              <span className="status-value">{sysInfo ? `${sysInfo.os}/${sysInfo.arch}` : '...'}</span>
            </div>
            <div className="status-item">
              <span className="status-label">Uptime</span>
              <span className="status-value">{sysInfo ? sysInfo.uptime : '...'}</span>
            </div>
          </div>
        </div>

        {/* CRUD Items Feature */}
        <div className="card col-8">
          <div className="card-header">
            <h3 className="card-title">📦 Application Resources (/api/v1/items)</h3>
            <span className="badge">{items.length} items</span>
          </div>

          <form onSubmit={handleAddItem} className="item-form">
            <input
              type="text"
              className="input-field"
              placeholder="Enter new item name..."
              value={newItemName}
              onChange={(e) => setNewItemName(e.target.value)}
            />
            <button type="submit" className="btn btn-primary">
              + Add Item
            </button>
          </form>

          <div className="items-list">
            {items.length === 0 ? (
              <p style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '1rem' }}>
                No items found. Create one above!
              </p>
            ) : (
              items.map((item) => (
                <div key={item.id} className="item-row">
                  <div className="item-info">
                    <span className="item-name">{item.name}</span>
                    <span className="item-date">Created by {item.created_by || user.username} at {new Date(item.created_at).toLocaleTimeString()}</span>
                  </div>
                  <span className="badge">ID: {item.id}</span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* API Inspector */}
        <div className="card col-12">
          <div className="card-header">
            <h3 className="card-title">🧪 API Inspector & Response Viewer</h3>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button onClick={() => testEndpoint('/api/v1/health')} className="btn btn-secondary" style={{ fontSize: '0.8rem', padding: '0.4rem 0.8rem' }}>
                GET /health
              </button>
              <button onClick={() => testEndpoint('/api/v1/user')} className="btn btn-secondary" style={{ fontSize: '0.8rem', padding: '0.4rem 0.8rem' }}>
                GET /user
              </button>
              <button onClick={() => testEndpoint('/api/v1/info')} className="btn btn-secondary" style={{ fontSize: '0.8rem', padding: '0.4rem 0.8rem' }}>
                GET /info
              </button>
              <button onClick={() => testEndpoint('/api/v1/items')} className="btn btn-secondary" style={{ fontSize: '0.8rem', padding: '0.4rem 0.8rem' }}>
                GET /items
              </button>
            </div>
          </div>
          <pre className="code-block">
            {loading ? 'Fetching API response...' : apiResponse}
          </pre>
        </div>
      </div>

      {/* Footer */}
      <footer className="footer">
        <p>Go Application Framework • Self-Hosted & Offline Capable • Built with Go & React</p>
      </footer>
    </div>
  );
}
