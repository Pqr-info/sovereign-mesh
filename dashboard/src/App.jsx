import React, { useState, useEffect } from 'react';
import './App.css';

function App() {
  const [activeTab, setActiveTab] = useState('android-ui');
  
  // Simulated UI layouts
  const [androidLayouts, setAndroidLayouts] = useState([
    {
      id: 'login-screen',
      name: 'Agent Login View',
      xml: `<?xml version="1.0" encoding="utf-8"?>
<LinearLayout xmlns:android="http://schemas.android.com/apk/res/android"
    android:layout_width="match_parent"
    android:layout_height="match_parent"
    android:orientation="vertical"
    android:gravity="center"
    android:background="#16171d"
    android:padding="24dp">
    
    <ImageView
        android:id="@+id/logo"
        android:layout_width="120dp"
        android:layout_height="120dp"
        android:src="@drawable/sovereign_logo"
        android:layout_marginBottom="32dp" />

    <TextView
        android:id="@+id/title"
        android:layout_width="wrap_content"
        android:layout_height="wrap_content"
        android:text="Sovereign Swarm Core"
        android:textColor="#f3f4f6"
        android:textSize="28sp"
        android:textStyle="bold"
        android:layout_marginBottom="8dp" />

    <TextView
        android:id="@+id/subtitle"
        android:layout_width="wrap_content"
        android:layout_height="wrap_content"
        android:text="Authenticate nodes in the 2026 Sovereign Mesh"
        android:textColor="#9ca3af"
        android:textSize="14sp"
        android:layout_marginBottom="40dp" />

    <EditText
        android:id="@+id/node_address"
        android:layout_width="match_parent"
        android:layout_height="56dp"
        android:hint="Sovereign Node Address (81 chars)"
        android:textColorHint="#4b5563"
        android:textColor="#f3f4f6"
        android:background="@drawable/edittext_glass"
        android:paddingHorizontal="16dp"
        android:layout_marginBottom="16dp"
        android:inputType="text" />

    <Button
        android:id="@+id/btn_authenticate"
        android:layout_width="match_parent"
        android:layout_height="56dp"
        android:text="CONNECT IDENTITY"
        android:textColor="#ffffff"
        android:background="@drawable/button_glow"
        android:textStyle="bold" />
</LinearLayout>`
    },
    {
      id: 'dashboard-screen',
      name: 'Agent Live Dashboard',
      xml: `<?xml version="1.0" encoding="utf-8"?>
<RelativeLayout xmlns:android="http://schemas.android.com/apk/res/android"
    android:layout_width="match_parent"
    android:layout_height="match_parent"
    android:background="#16171d"
    android:padding="16dp">

    <TextView
        android:id="@+id/header_title"
        android:layout_width="wrap_content"
        android:layout_height="wrap_content"
        android:text="Active Swarm Overview"
        android:textColor="#f3f4f6"
        android:textSize="20sp"
        android:textStyle="bold" />

    <TextView
        android:id="@+id/status_pill"
        android:layout_width="wrap_content"
        android:layout_height="wrap_content"
        android:layout_alignParentRight="true"
        android:text="ACTIVE SWARM"
        android:textColor="#10b981"
        android:textSize="12sp"
        android:background="@drawable/pill_green"
        android:paddingHorizontal="8dp"
        android:paddingVertical="4dp" />

    <ScrollView
        android:layout_width="match_parent"
        android:layout_height="match_parent"
        android:layout_below="@id/header_title"
        android:layout_marginTop="24dp">

        <LinearLayout
            android:layout_width="match_parent"
            android:layout_height="wrap_content"
            android:orientation="vertical">

            <!-- Telemetry Card -->
            <LinearLayout
                android:layout_width="match_parent"
                android:layout_height="wrap_content"
                android:background="#1f2028"
                android:orientation="vertical"
                android:padding="16dp"
                android:layout_marginBottom="16dp">
                
                <TextView
                    android:layout_width="wrap_content"
                    android:layout_height="wrap_content"
                    android:text="GSR Stability Score"
                    android:textColor="#9ca3af"
                    android:textSize="12sp" />
                <TextView
                    android:id="@+id/stability_value"
                    android:layout_width="wrap_content"
                    android:layout_height="wrap_content"
                    android:text="98.42%"
                    android:textColor="#c084fc"
                    android:textSize="32sp"
                    android:textStyle="bold" />
            </LinearLayout>
        </LinearLayout>
    </ScrollView>
</RelativeLayout>`
    }
  ]);

  const [selectedLayout, setSelectedLayout] = useState(androidLayouts[0]);
  const [telemetry, setTelemetry] = useState({
    stability_score: 0.9842,
    entropy_level: 0.052,
    theta: 0.441,
    agentic_weight: 0.892,
    ethical_variance: 0.001,
    active_agents: 3,
    jailed_agents: 0
  });

  const [logs, setLogs] = useState([
    { time: '02:26:10', msg: 'INITIALIZING STARBIRTH PROTOCOL (SBP-001) - 2026 Swarm...', type: 'info' },
    { time: '02:26:10', msg: 'OUROBOROS SENTINEL: Watchdog daemon activated.', type: 'info' },
    { time: '02:26:15', msg: 'SENTINEL ALERT: Process "web_portal" has flatlined! Initiating Resurrection...', type: 'warn' },
    { time: '02:26:16', msg: 'HEALED: Process "web_portal" is back in stable flight path.', type: 'success' }
  ]);

  // Fetch metrics from remote VM controller periodically
  useEffect(() => {
    // Check local server dashboard endpoints or GCP endpoint if mapped
    // For local dev sandbox, mock or pull if running
  }, []);

  return (
    <div className="mesh-app">
      <header className="mesh-header">
        <div className="logo-section">
          <div className="mesh-logo"></div>
          <span className="logo-text">Sovereign Mesh Test Suite</span>
        </div>
        <nav className="mesh-nav">
          <button 
            className={`nav-btn ${activeTab === 'android-ui' ? 'active' : ''}`}
            onClick={() => setActiveTab('android-ui')}
          >
            Android UI Layout Tester
          </button>
          <button 
            className={`nav-btn ${activeTab === 'controller-metrics' ? 'active' : ''}`}
            onClick={() => setActiveTab('controller-metrics')}
          >
            Sovereign Controller Telemetry
          </button>
        </nav>
        <div className="billing-badge">
          Identity: alan@w-isp.net
        </div>
      </header>

      <main className="mesh-main">
        {activeTab === 'android-ui' && (
          <div className="android-layout-tester animate-fade-in">
            <div className="sidebar-layouts">
              <h3>Android UI Layouts</h3>
              <div className="layouts-list">
                {androidLayouts.map(layout => (
                  <button 
                    key={layout.id}
                    className={`layout-item-btn ${selectedLayout.id === layout.id ? 'active' : ''}`}
                    onClick={() => setSelectedLayout(layout)}
                  >
                    <span className="layout-name">{layout.name}</span>
                    <span className="layout-type">XML Layout</span>
                  </button>
                ))}
              </div>
            </div>

            <div className="layout-editor-pane">
              <div className="editor-header">
                <h4>Layout Source Code (AXML)</h4>
                <button className="compile-btn" onClick={() => alert('Layout Compiled successfully into APK Binary!')}>
                  Compile Layout
                </button>
              </div>
              <textarea 
                className="xml-editor"
                value={selectedLayout.xml}
                onChange={(e) => {
                  const updated = { ...selectedLayout, xml: e.target.value };
                  setSelectedLayout(updated);
                  setAndroidLayouts(androidLayouts.map(l => l.id === selectedLayout.id ? updated : l));
                }}
              />
            </div>

            <div className="layout-preview-pane">
              <div className="preview-header">
                <h4>Visual Canvas (Android Device Render)</h4>
                <div className="preview-device-info">Target: Pixel 8 (API 34)</div>
              </div>
              <div className="device-frame">
                <div className="device-screen">
                  {selectedLayout.id === 'login-screen' ? (
                    <div className="simulated-login">
                      <div className="simulated-logo-circle">🧬</div>
                      <h2 className="sim-title">Sovereign Swarm Core</h2>
                      <p className="sim-subtitle">Authenticate nodes in the 2026 Sovereign Mesh</p>
                      <input 
                        type="text" 
                        className="sim-input" 
                        placeholder="Sovereign Node Address (81 chars)" 
                        defaultValue="addr12345678901234567890123456789012345678901234567890123456789012345678901234567"
                      />
                      <button className="sim-button">CONNECT IDENTITY</button>
                    </div>
                  ) : (
                    <div className="simulated-dashboard">
                      <div className="sim-db-header">
                        <span className="sim-db-title">Active Swarm Overview</span>
                        <span className="sim-db-badge">ACTIVE SWARM</span>
                      </div>
                      <div className="sim-db-card">
                        <span className="sim-card-label">GSR Stability Score</span>
                        <span className="sim-card-value">{(telemetry.stability_score * 100).toFixed(2)}%</span>
                      </div>
                      <div className="sim-db-stats">
                        <div className="sim-stat-box">
                          <span className="sim-stat-val">{telemetry.active_agents}</span>
                          <span className="sim-stat-lbl">Active Agents</span>
                        </div>
                        <div className="sim-stat-box">
                          <span className="sim-stat-val">{telemetry.agentic_weight}</span>
                          <span className="sim-stat-lbl">Agentic Weight</span>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'controller-metrics' && (
          <div className="controller-metrics-view animate-fade-in">
            <div className="metrics-summary-grid">
              <div className="metric-card stability">
                <span className="m-label">Stability Index (GSR)</span>
                <span className="m-val">{(telemetry.stability_score * 100).toFixed(2)}%</span>
                <div className="m-bar"><div className="m-bar-inner" style={{ width: `${telemetry.stability_score * 100}%` }}></div></div>
              </div>
              <div className="metric-card entropy">
                <span className="m-label">Entropy Level</span>
                <span className="m-val">{telemetry.entropy_level.toFixed(4)}</span>
                <div className="m-bar"><div className="m-bar-inner" style={{ width: `${telemetry.entropy_level * 100}%`, backgroundColor: '#10b981' }}></div></div>
              </div>
              <div className="metric-card agentic">
                <span className="m-label">Agentic Weight (Alpha)</span>
                <span className="m-val">{telemetry.agentic_weight.toFixed(3)}</span>
                <div className="m-bar"><div className="m-bar-inner" style={{ width: `${telemetry.agentic_weight * 100}%`, backgroundColor: '#f59e0b' }}></div></div>
              </div>
              <div className="metric-card validators">
                <span className="m-label">Jailed Agents</span>
                <span className="m-val">{telemetry.jailed_agents}</span>
                <div className="m-bar"><div className="m-bar-inner" style={{ width: '0%', backgroundColor: '#ef4444' }}></div></div>
              </div>
            </div>

            <div className="mesh-terminal-section">
              <div className="terminal-header">
                <span>Ouroboros Sentinel Console & Live Events</span>
                <span className="pulse-green">● Connected</span>
              </div>
              <div className="terminal-body">
                {logs.map((log, index) => (
                  <div key={index} className={`terminal-line ${log.type}`}>
                    <span className="t-time">[{log.time}]</span>
                    <span className="t-msg">{log.msg}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
