import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import confetti from "canvas-confetti";
import { Scene3D } from "./components/Scene3D";

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

  useEffect(() => {
    if (completedCommands.length === commands.length) {
      const duration = 3 * 1000;
      const animationEnd = Date.now() + duration;
      const defaults = {
        startVelocity: 30,
        spread: 360,
        ticks: 60,
        zIndex: 0,
        colors: ["#2496ED", "#086DD7", "#00BBFF", "#75AADB", "#ffffff"],
      };

      const randomInRange = (min: number, max: number) =>
        Math.random() * (max - min) + min;

      const interval = setInterval(function () {
        const timeLeft = animationEnd - Date.now();

        if (timeLeft <= 0) {
          return clearInterval(interval);
        }

        const particleCount = 50 * (timeLeft / duration);
        confetti({
          ...defaults,
          particleCount,
          origin: { x: randomInRange(0.1, 0.3), y: Math.random() - 0.2 },
        });
        confetti({
          ...defaults,
          particleCount,
          origin: { x: randomInRange(0.7, 0.9), y: Math.random() - 0.2 },
        });
      }, 250);
    }
  }, [completedCommands.length]);

  const [stars, setStars] = useState<number | null>(null);

  useEffect(() => {
    fetch("https://api.github.com/repos/k1lgor/container-diet")
      .then((res) => res.json())
      .then((data) => setStars(data.stargazers_count))
      .catch((err) => console.error("Failed to fetch stars:", err));
  }, []);

  return (
    <div className="app-container">
      {/* Docker grid overlay */}
      <div className="scanlines"></div>

      {/* 3D Scene Background */}
      <Scene3D />

      {/* Ambient glow orbs */}
      <div className="background-decoration">
        <div className="glow-orb glow-orb-top"></div>
        <div className="glow-orb glow-orb-bottom"></div>
      </div>

      {/* ─── Header ─── */}
      <header className="header">
        <motion.div
          className="header-content"
          initial={{ y: -80, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
        >
          <div className="logo">
            <span className="logo-whale">🐳</span>
            <h2 className="logo-text">
              Container <span className="logo-highlight">Diet</span>
            </h2>
            <span className="logo-version">v0.3.0</span>
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
        </motion.div>
      </header>

      {/* ─── Main Content ─── */}
      <main className="main-content">
        {/* Hero Section */}
        <section className="hero">
          <div className="hero-content">
            {/* Left: Hero Text */}
            <motion.div
              className="hero-text"
              initial={{ x: -60, opacity: 0 }}
              animate={{ x: 0, opacity: 1 }}
              transition={{
                duration: 0.8,
                delay: 0.2,
                ease: [0.22, 1, 0.36, 1],
              }}
            >
              <div className="version-badge">
                <span className="status-dot"></span>
                <span>v0.3.0 — Now Available</span>
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
                <motion.button
                  whileHover={{ scale: 1.04 }}
                  whileTap={{ scale: 0.96 }}
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
                </motion.button>
                <motion.button
                  whileHover={{ scale: 1.04 }}
                  whileTap={{ scale: 0.96 }}
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
                </motion.button>
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
                <div className="feature-item">
                  <span className="check-icon">✓</span>
                  Docker & Podman
                </div>
              </div>
            </motion.div>

            {/* Right: Terminal Window */}
            <motion.div
              className="terminal-wrapper"
              initial={{ x: 60, opacity: 0 }}
              animate={{ x: 0, opacity: 1 }}
              transition={{
                duration: 0.8,
                delay: 0.4,
                ease: [0.22, 1, 0.36, 1],
              }}
            >
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

                  <AnimatePresence>
                    {completedCommands.length === commands.length && (
                      <motion.div
                        className="ai-response"
                        initial={{ opacity: 0, y: 16 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.5 }}
                      >
                        <div className="warning-box">
                          <span className="warning-label">⚠ WARNING:</span> Your
                          Python layer is carrying extra weight! 🐘 Using
                          `python:3.9` instead of `3.9-slim` is like wearing
                          lead boots for a sprint.
                        </div>
                        <div className="suggestion-box">
                          <span className="suggestion-label">
                            ✓ SUGGESTION:
                          </span>{" "}
                          Switch to a slim base or use a multi-stage build to
                          purge build tools like `gcc` and `make`.
                        </div>
                        <div className="autofix-box">
                          <div className="autofix-header">
                            <span className="autofix-label">
                              🛠️ AUTO-FIX GENERATED
                            </span>
                            <span className="autofix-path">
                              Dockerfile.diet
                            </span>
                          </div>
                          <pre className="code-snippet">
                            <span className="keyword">FROM</span>{" "}
                            python:3.9-slim <span className="keyword">AS</span>{" "}
                            final
                            <span className="comment">
                              {"\n"}# Purged 220MB of build tools! 📉
                            </span>
                            {"\n"}
                            <span className="keyword">RUN</span> apt-get update
                            && apt-get install -y --no-install-recommends \
                            {"\n"}
                            {"    "}
                            <span className="highlight">libpq5</span> &&{" "}
                            <span className="keyword">rm</span> -rf
                            /var/lib/apt/lists/*{"\n"}
                            <span className="keyword">USER</span> appuser
                          </pre>
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>
              </div>
            </motion.div>
          </div>
        </section>

        {/* ─── Features Section ─── */}
        <motion.section
          className="section features-section"
          id="features"
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
        >
          <div className="section-header">
            <h2 className="section-title">AI-Powered Container Optimization</h2>
            <p className="section-description">
              Container Diet analyzes your Docker images and Dockerfiles to
              provide actionable, sassy-but-helpful optimization advice powered
              by GPT-4o.
            </p>
          </div>

          <div className="features-grid">
            {[
              {
                icon: "🧠",
                iconClass: "purple",
                title: "AI-Driven Analysis",
                desc: "Uses OpenAI (GPT-4o) to provide human-level, context-aware optimization tips. Get entertaining, roast-style feedback that makes optimization fun.",
              },
              {
                icon: "🐳",
                iconClass: "cyan",
                title: "Flexible Image Sources",
                desc: "Analyze local daemon images, pull remote images with --remote, or auto-pull missing local images with --pull-missing. Works with Docker and Podman.",
              },
              {
                icon: "🛡️",
                iconClass: "purple",
                title: "Security Focused",
                desc: "Detects root user violations, exposed secrets, unnecessary packages, and permission issues. Keep your containers secure by default.",
              },
              {
                icon: "🛠️",
                iconClass: "accent",
                title: "Auto-Fix Generation",
                desc: "Automatically generate an optimized Dockerfile (Dockerfile.diet) based on analysis. Purge build tools, optimize layers, and apply security best practices in one command.",
                highlighted: true,
              },
            ].map((feature, i) => (
              <motion.div
                key={i}
                className={`feature-card${feature.highlighted ? " highlighted" : ""}`}
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: i * 0.1 }}
              >
                <div className={`feature-icon ${feature.iconClass}`}>
                  {feature.icon}
                </div>
                <h3 className="feature-title">{feature.title}</h3>
                <p className="feature-description">{feature.desc}</p>
              </motion.div>
            ))}
          </div>
        </motion.section>

        {/* ─── Stats Section ─── */}
        <motion.section
          className="section stats-section"
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
        >
          <div className="stats-grid">
            {[
              { value: "GPT-4o", label: "AI Model", colorClass: "cyan" },
              {
                value: "Local/Remote",
                label: "Image Sources",
                colorClass: "white",
              },
              {
                value: "Docker + Podman",
                label: "Platform Support",
                colorClass: "white",
              },
              {
                value: "Open Source",
                label: "MIT License",
                colorClass: "purple",
              },
            ].map((stat, i) => (
              <motion.div
                key={i}
                className="stat-card"
                initial={{ opacity: 0, scale: 0.95 }}
                whileInView={{ opacity: 1, scale: 1 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: i * 0.08 }}
              >
                <span className={`stat-value ${stat.colorClass}`}>
                  {stat.value}
                </span>
                <span className="stat-label">{stat.label}</span>
              </motion.div>
            ))}
          </div>
        </motion.section>
      </main>

      {/* ─── Footer ─── */}
      <footer className="footer">
        <div className="footer-content">
          <div className="footer-logo">
            <span>🐳</span>
            <span className="logo-text">
              Container <span className="logo-highlight">Diet</span>
            </span>
          </div>
          <div className="footer-copyright">
            © 2026 Container Diet — Built for speed and security.
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;
