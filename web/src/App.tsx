import { useState, useEffect } from "react";

function App() {
  const [text, setText] = useState("");
  const fullText = "container-diet analyze my-app:latest";

  useEffect(() => {
    let index = 0;
    const interval = setInterval(() => {
      setText(fullText.slice(0, index));
      index++;
      if (index > fullText.length) {
        clearInterval(interval);
      }
    }, 100);
    return () => clearInterval(interval);
  }, []);

  const [stars, setStars] = useState<number | null>(null);

  useEffect(() => {
    fetch("https://api.github.com/repos/k1lgor/container-diet")
      .then((res) => res.json())
      .then((data) => setStars(data.stargazers_count))
      .catch((err) => console.error("Failed to fetch stars:", err));
  }, []);

  return (
    <div>
      <div className="scanlines"></div>

      {/* Product Hunt Badge */}
      <a
        href="https://www.producthunt.com/products/container-diet?embed=true&utm_source=badge-featured&utm_medium=badge&utm_source=badge-container&#0045;diet"
        target="_blank"
        rel="noopener noreferrer"
        style={{
          position: "absolute",
          top: "1rem",
          left: "1rem",
          zIndex: 100,
        }}
      >
        <img
          src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=1041412&theme=light&t=1764229267121"
          alt="Container Diet - Slim down your Docker images with AI-powered sass. 🐳💅 | Product Hunt"
          style={{ width: "250px", height: "54px" }}
          width="250"
          height="54"
        />
      </a>

      {/* GitHub Stars Badge */}
      <a
        href="https://github.com/k1lgor/container-diet"
        target="_blank"
        rel="noopener noreferrer"
        style={{
          position: "absolute",
          top: "1rem",
          right: "1rem",
          display: "flex",
          alignItems: "center",
          gap: "0.5rem",
          padding: "0.5rem 1rem",
          background: "rgba(0, 0, 0, 0.6)",
          border: "1px solid var(--primary-color)",
          borderRadius: "4px",
          color: "#fff",
          textDecoration: "none",
          fontSize: "0.9rem",
          backdropFilter: "blur(4px)",
          zIndex: 100,
        }}
      >
        <svg
          height="20"
          viewBox="0 0 16 16"
          version="1.1"
          width="20"
          aria-hidden="true"
          fill="currentColor"
        >
          <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
        </svg>
        <span>Star on GitHub</span>
        {stars !== null && (
          <span
            style={{
              background: "#333",
              padding: "0.1rem 0.4rem",
              borderRadius: "10px",
              fontSize: "0.8rem",
              fontWeight: "bold",
            }}
          >
            {stars}
          </span>
        )}
      </a>

      {/* Hero Section */}
      <section className="hero">
        <h1>Container Diet</h1>
        <p
          style={{
            maxWidth: "600px",
            margin: "0 auto 2rem",
            fontSize: "1.2rem",
          }}
        >
          OPTIMIZE YOUR DOCKER IMAGES WITH{" "}
          <span style={{ color: "var(--primary-color)" }}>AI-POWERED</span>{" "}
          PRECISION.
        </p>

        <div style={{ display: "flex", gap: "1.5rem", marginTop: "2rem" }}>
          <button
            onClick={() =>
              window.open(
                "https://github.com/k1lgor/container-diet/releases",
                "_blank"
              )
            }
          >
            Download CLI
          </button>
          <button
            className="secondary"
            onClick={() =>
              window.open("https://github.com/k1lgor/container-diet", "_blank")
            }
          >
            View Source
          </button>
        </div>

        <div className="terminal-window">
          <div className="terminal-header">
            <div className="dot red"></div>
            <div className="dot yellow"></div>
            <div className="dot green"></div>
          </div>
          <div className="terminal-body">
            <div className="command-line">
              <span className="prompt">➜</span>
              <span className="command">
                {text}
                <span className="cursor"></span>
              </span>
            </div>
            {text.length === fullText.length && (
              <div
                style={{ marginTop: "1rem", animation: "fadeIn 0.5s ease-out" }}
              >
                <div style={{ color: "#a0a0a0" }}>
                  Analyzing image structure...
                </div>
                <div style={{ color: "#a0a0a0" }}>Fetching layer data...</div>
                <br />
                <div style={{ color: "var(--primary-color)" }}>
                  [AI ANALYSIS COMPLETE]
                </div>
                <div style={{ marginTop: "0.5rem" }}>
                  <span style={{ color: "#ffbd2e" }}>⚠ WARNING:</span> Layer 3
                  (150MB) contains 'gcc' and 'make'.
                  <br />
                  <span style={{ color: "#27c93f" }}>✓ SUGGESTION:</span> Use
                  multi-stage build to remove build tools.
                  <br />
                  <span style={{ color: "#27c93f" }}>✓ SUGGESTION:</span>{" "}
                  Combine RUN instructions on lines 4-6.
                </div>
              </div>
            )}
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="section">
        <h2 className="text-center">System Capabilities</h2>
        <div className="features-grid">
          <div className="feature-card">
            <div className="feature-icon">🧠</div>
            <h3>AI-Driven Analysis</h3>
            <p>
              Leverages advanced LLMs to understand your container context and
              provide human-level optimization advice.
            </p>
          </div>
          <div className="feature-card">
            <div className="feature-icon">⚡</div>
            <h3>Instant Feedback</h3>
            <p>
              Local daemon integration means no pushing to registries. Analyze
              images directly from your machine.
            </p>
          </div>
          <div className="feature-card">
            <div className="feature-icon">🛡️</div>
            <h3>Security First</h3>
            <p>
              Detects non-root user violations, exposed secrets, and unnecessary
              package managers in production images.
            </p>
          </div>
        </div>
      </section>

      <footer
        style={{
          textAlign: "center",
          padding: "4rem 0",
          color: "#666",
          borderTop: "1px solid #222",
        }}
      >
        <p>CONTAINER DIET CLI v0.1.0</p>
      </footer>
    </div>
  );
}

export default App;
