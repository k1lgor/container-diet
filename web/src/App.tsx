import { useState, useEffect } from "react";

function App() {
  const [commandIndex, setCommandIndex] = useState(0);
  const [currentText, setCurrentText] = useState("");
  const [completedCommands, setCompletedCommands] = useState<number[]>([]);

  const commands = [
    {
      text: "container-diet analyze python:3.9 --auto-fix",
      output: [
        "🔍 Scanning image: python:3.9",
        "Pulling from remote...",
        "Analyzing remote image: python:3.9",
        "",
        "📊 IMAGE SUMMARY",
        "📦 Image: python:3.9",
        "📊 Total Size: 386.32 MB",
        "🍰 Layers: 7",
        "",
        "🤖 [AI ANALYSIS]",
        "🚢 Asking the Container Dietician for insights...",
      ],
    },
  ];

  useEffect(() => {
    if (commandIndex >= commands.length) return;

    const command = commands[commandIndex];
    let charIndex = 0;

    const typeInterval = setInterval(() => {
      if (charIndex <= command.text.length) {
        setCurrentText(command.text.slice(0, charIndex));
        charIndex++;
      } else {
        clearInterval(typeInterval);
        setTimeout(() => {
          setCompletedCommands((prev) => [...prev, commandIndex]);
          if (commandIndex < commands.length - 1) {
            setCommandIndex(commandIndex + 1);
            setCurrentText("");
          }
        }, 500);
      }
    }, 60);

    return () => clearInterval(typeInterval);
  }, [commandIndex]);

  const [stars, setStars] = useState<number | null>(null);

  useEffect(() => {
    fetch("https://api.github.com/repos/k1lgor/container-diet")
      .then((res) => res.json())
      .then((data) => setStars(data.stargazers_count))
      .catch((err) => console.error("Failed to fetch stars:", err));
  }, []);

  return (
    <div className="app-container">
      <div className="scanlines"></div>
      <div className="background-decoration">
        <div className="glow-orb glow-orb-top"></div>
        <div className="glow-orb glow-orb-bottom"></div>
      </div>

      {/* Header */}
      <header className="header">
        <div className="header-content">
          <div className="logo">
            <h2 className="logo-text">
              Container <span className="logo-highlight">Diet</span>
            </h2>
          </div>

          <div className="header-badges">
            {/* Product Hunt Badge */}
            <a
              href="https://www.producthunt.com/products/container-diet"
              target="_blank"
              rel="noopener noreferrer"
              className="ph-badge"
            >
              <img
                src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=1041412&theme=dark&t=1764229267121"
                alt="Container Diet on Product Hunt"
                width="200"
                height="43"
              />
            </a>

            {/* GitHub Stars */}
            <a
              href="https://github.com/k1lgor/container-diet"
              target="_blank"
              rel="noopener noreferrer"
              className="github-badge"
            >
              <svg
                height="16"
                viewBox="0 0 16 16"
                width="16"
                fill="currentColor"
              >
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
              </svg>
              <span>Star</span>
              {stars !== null && <span className="star-count">{stars}</span>}
            </a>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <main className="main-content">
        <section className="hero">
          <div className="hero-content">
            <div className="hero-text">
              <div className="version-badge">
                <span className="status-dot"></span>
                <span className="version-text">v0.3.0 Now Available</span>
              </div>

              <h1 className="hero-title">
                Slim Down Your <br />
                <span className="hero-gradient">Containers.</span>
              </h1>

              <p className="hero-description">
                The ultimate CLI tool to analyze, optimize, and secure your
                Docker images in seconds. Reduce bloat by up to 90% without
                breaking your application.
              </p>

              <div className="cta-buttons">
                <button
                  className="btn btn-primary"
                  onClick={() =>
                    window.open(
                      "https://github.com/k1lgor/container-diet/releases",
                      "_blank",
                    )
                  }
                >
                  <span className="btn-icon">⬇</span>
                  Download CLI
                </button>
                <button
                  className="btn btn-secondary"
                  onClick={() =>
                    window.open(
                      "https://github.com/k1lgor/container-diet",
                      "_blank",
                    )
                  }
                >
                  <span className="btn-icon">{"</>"}</span>
                  View on GitHub
                </button>
              </div>

              <div className="features-list">
                <div className="feature-item">
                  <span className="check-icon">✓</span>
                  Open Source
                </div>
                <div className="feature-item">
                  <span className="check-icon">✓</span>
                  CI/CD Integrated
                </div>
              </div>
            </div>

            {/* Terminal Window */}
            <div className="terminal-wrapper">
              <div className="terminal-glow"></div>
              <div className="terminal-window">
                <div className="terminal-header">
                  <div className="terminal-dots">
                    <div className="dot red"></div>
                    <div className="dot yellow"></div>
                    <div className="dot green"></div>
                  </div>
                  <div className="terminal-title">bash — container-diet</div>
                  <div className="terminal-spacer"></div>
                </div>

                <div className="terminal-body">
                  {completedCommands.map((cmdIdx) => {
                    const cmd = commands[cmdIdx];
                    return (
                      <div key={cmdIdx}>
                        <div className="command-line">
                          <span className="prompt">$</span>
                          <span className="command-text">{cmd.text}</span>
                        </div>
                        {cmd.output.map((line, i) => (
                          <div key={i} className="command-output">
                            {line}
                          </div>
                        ))}
                      </div>
                    );
                  })}

                  {commandIndex < commands.length && (
                    <div className="command-line">
                      <span className="prompt">$</span>
                      <span className="command-text">
                        {currentText}
                        <span className="cursor"></span>
                      </span>
                    </div>
                  )}

                  {completedCommands.length === commands.length && (
                    <div className="ai-response">
                      <div className="warning-box">
                        <span className="warning-label">⚠ WARNING:</span> Your
                        Python layer is carrying extra weight! 🐘 Using
                        `python:3.9` instead of `3.9-slim` is like wearing lead
                        boots for a sprint.
                      </div>
                      <div className="suggestion-box">
                        <span className="suggestion-label">✓ SUGGESTION:</span>{" "}
                        Switch to a slim base or use a multi-stage build to
                        purge build tools like `gcc` and `make`.
                      </div>
                      <div className="autofix-box">
                        <div className="autofix-header">
                          <span className="autofix-label">
                            🛠️ AUTO-FIX GENERATED
                          </span>
                          <span className="autofix-path">Dockerfile.diet</span>
                        </div>
                        <pre className="code-snippet">
                          {`FROM python:3.9-slim AS final
# Purged 220MB of build tools! 📉
RUN apt-get update && apt-get install -y --no-install-recommends \\
    libpq5 && rm -rf /var/lib/apt/lists/*
USER appuser`}
                        </pre>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Features Section */}
        <section className="section features-section" id="features">
          <div className="section-header">
            <h2 className="section-title">AI-Powered Container Optimization</h2>
            <p className="section-description">
              Container Diet analyzes your Docker images and Dockerfiles to
              provide actionable, sassy-but-helpful optimization advice powered
              by GPT-4o.
            </p>
          </div>

          <div className="features-grid">
            <div className="feature-card">
              <div className="feature-icon purple">🧠</div>
              <h3 className="feature-title">AI-Driven Analysis</h3>
              <p className="feature-description">
                Uses OpenAI (GPT-4o) to provide human-level, context-aware
                optimization tips. Get entertaining, roast-style feedback that
                makes optimization fun.
              </p>
            </div>

            <div className="feature-card">
              <div className="feature-icon cyan">🐳</div>
              <h3 className="feature-title">Flexible Image Sources</h3>
              <p className="feature-description">
                Analyze local daemon images, pull remote images with --remote,
                or auto-pull missing local images with --pull-missing. Works
                with Docker and Podman.
              </p>
            </div>

            <div className="feature-card">
              <div className="feature-icon purple">🛡️</div>
              <h3 className="feature-title">Security Focused</h3>
              <p className="feature-description">
                Detects root user violations, exposed secrets, unnecessary
                packages, and permission issues. Keep your containers secure by
                default.
              </p>
            </div>

            <div className="feature-card highlighted">
              <div className="feature-icon accent">🛠️</div>
              <h3 className="feature-title">Auto-Fix Generation</h3>
              <p className="feature-description">
                Automatically generate an optimized Dockerfile (Dockerfile.diet)
                based on analysis. Purge build tools, optimize layers, and apply
                security best practices in one command.
              </p>
            </div>
          </div>
        </section>

        {/* Stats Section */}
        <section className="section stats-section">
          <div className="stats-grid">
            <div className="stat-card">
              <span className="stat-value cyan">GPT-4o</span>
              <span className="stat-label">AI Model</span>
            </div>
            <div className="stat-card">
              <span className="stat-value white">Local/Remote</span>
              <span className="stat-label">Image Sources</span>
            </div>
            <div className="stat-card">
              <span className="stat-value white">Docker + Podman</span>
              <span className="stat-label">Platform Support</span>
            </div>
            <div className="stat-card">
              <span className="stat-value purple">Open Source</span>
              <span className="stat-label">MIT License</span>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="footer">
        <div className="footer-content">
          <div className="footer-copyright">
            © 2026 Container Diet. Built for speed and security.
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;
