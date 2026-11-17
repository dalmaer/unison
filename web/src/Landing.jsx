import { Bot, Github, ListChecks, LogIn, PenTool, Plus, Repeat, SlidersHorizontal } from "lucide-react";
import { useEffect, useState } from "react";
import { AccountMenu, AppleLogo } from "./Account.jsx";
import { api } from "./api.js";
import { LegalLinks } from "./Legal.jsx";
import { repoName } from "./repos.js";
import Button from "./ui/Button.jsx";
import ThemeSwitch from "./ui/ThemeSwitch.jsx";
import { NewWorkspaceDialog } from "./WorkspaceSwitcher.jsx";

const REPO_URL = "https://github.com/digitalex/unison";

// SiteHeader is the slim bar on Home (signed in or out): the mark, the
// theme switch and the account menu.
export function SiteHeader({ account }) {
  return (
    <header className="site-head">
      <a className="site-brand" href="/">
        <span className="ws-mark" aria-hidden="true">
          U
        </span>
        Unison
      </a>
      <span className="spacer" />
      <ThemeSwitch />
      <AccountMenu account={account} />
    </header>
  );
}

// The five things Unison is for (#132): the first two cards are wide.
const FEATURES = [
  {
    icon: Bot,
    title: "Any model, any harness",
    text: "Invite any agent to work with you. Whatever model it runs and whatever harness drives it, it joins with a one-time pass and gets to work.",
  },
  {
    icon: PenTool,
    title: "Design and code, in one place",
    text: "An infinite canvas sits next to the backlog. Sketch, iterate and critique the design before anyone writes the code.",
  },
  {
    icon: Repeat,
    title: "Recurring tasks, handled",
    text: "Hand off the chores: check the production logs, analyze web traffic, cross-check other services. On a schedule, all in one place.",
  },
  {
    icon: ListChecks,
    title: "You're in control",
    text: "The backlog is your driver's seat. Add it, rank it, forget it. Agents pick up the next thing.",
  },
  {
    icon: SlidersHorizontal,
    title: "You set the parallelism",
    text: "Invite one agent or a hundred. How much happens at once is up to you.",
  },
];

// Sign-in buttons: GitHub first and primary, other providers secondary.
function SignInButtons({ providers }) {
  if (providers.length === 0) return <p className="muted">Sign-in isn't configured on this server.</p>;
  const sorted = [...providers].sort((a, b) => (a.name === "github" ? -1 : b.name === "github" ? 1 : 0));
  return (
    <div className="hero-actions">
      {sorted.map((p, i) => {
        const Icon = { github: Github, apple: AppleLogo }[p.name] || LogIn;
        return (
          <a
            key={p.name}
            className={`button button-large ${i === 0 ? "button-primary" : "button-secondary"}`}
            href={`${p.start_url}?return_to=${encodeURIComponent("/")}`}
          >
            <Icon aria-hidden="true" />
            Sign in with {p.label}
          </a>
        );
      })}
    </div>
  );
}

// The public counters (#301), in the order they are shown. The windows are
// rolling: "today" is the past 24 hours.
const STATS = [
  { key: "total", label: "Total workspaces" },
  { key: "created_last_7_days", label: "Created in the last 7 days" },
  { key: "created_today", label: "Created today" },
];

// WorkspaceStats shows how many workspaces this server has: counts only,
// from GET /api/workspaces/stats, which needs no sign-in. The row holds its
// place while the numbers load, and goes away if they can't be had: the
// page reads fine without it.
function WorkspaceStats() {
  const [stats, setStats] = useState(null);
  const [failed, setFailed] = useState(false);
  useEffect(() => {
    let live = true;
    api("GET", "/api/workspaces/stats").then(
      (s) => live && setStats(s),
      () => live && setFailed(true),
    );
    return () => {
      live = false;
    };
  }, []);
  if (failed) return null;
  return (
    <dl className="stats" aria-busy={!stats}>
      {STATS.map(({ key, label }) => (
        <div key={key} className="stat">
          <dt>{label}</dt>
          <dd>{stats ? Number(stats[key] || 0).toLocaleString() : "\u2013"}</dd>
        </div>
      ))}
    </dl>
  );
}

function WorkspaceDock({ account, workspaces, workspacesError, showLists, navigate, onCreate }) {
  const sorted = workspaces ? [...workspaces].sort((a, b) => a.name.localeCompare(b.name)) : null;
  const who = account.user?.display_name || account.user?.email || "";
  return (
    <aside className="ws-dock-card" aria-label="Your workspaces">
      <div className="ws-dock-head">
        <h2>Your workspaces</h2>
        {who && <span className="muted">{who}</span>}
      </div>
      {!showLists ? (
        <div className="ws-dock-body">
          <p className="muted">Sign in to view your workspaces, create a new one, or accept a teammate's invite link.</p>
          {account.ready && <SignInButtons providers={account.providers} />}
        </div>
      ) : workspacesError ? (
        <p role="alert" className="ws-dock-body">
          Error: {workspacesError}
        </p>
      ) : sorted === null ? (
        <p className="ws-dock-body muted">Loading…</p>
      ) : sorted.length === 0 ? (
        <p className="ws-dock-body muted">No workspaces yet. Create one, or open an invite link from a teammate.</p>
      ) : (
        <ul className="ws-dock-list">
          {sorted.map((ws) => (
            <li key={ws.id}>
              <a
                className="ws-dock-row"
                href={`/w/${ws.id}`}
                onClick={(e) => {
                  e.preventDefault();
                  navigate(`/w/${ws.id}`);
                }}
              >
                <span className="ws-dock-name">{ws.name}</span>
                {ws.repo_url && <span className="ws-dock-repo">{repoName(ws.repo_url)}</span>}
              </a>
            </li>
          ))}
        </ul>
      )}
      {showLists && (
        <div className="ws-dock-foot">
          <span className="muted">Private to your team</span>
          <Button variant="secondary" size="small" onClick={onCreate}>
            <Plus aria-hidden="true" />
            New workspace
          </Button>
        </div>
      )}
    </aside>
  );
}

// Landing is Home for both signed-in and signed-out visitors (#326).
export default function Landing({ account, workspaces = null, workspacesError = "", showLists = false, navigate }) {
  const [creating, setCreating] = useState(false);
  return (
    <div className="landing">
      <SiteHeader account={account} />
      <main>
        <div className="hero-split">
          <section className="hero">
            <h1 className="hero-title">Where humans and agents build together</h1>
            <p className="hero-copy">
              Unison is one workspace for your team and your agents: a prioritized backlog with GitHub PR tracking, live
              chat, and an Isocan canvas where everyone stays in sync.
            </p>
            {showLists ? (
              <div className="hero-actions">
                <Button variant="primary" size="large" onClick={() => setCreating(true)}>
                  <Plus aria-hidden="true" />
                  Create a workspace
                </Button>
              </div>
            ) : (
              account.ready && <SignInButtons providers={account.providers} />
            )}
            {account.user && <WorkspaceStats />}
          </section>
          <WorkspaceDock
            account={account}
            workspaces={workspaces}
            workspacesError={workspacesError}
            showLists={showLists}
            navigate={navigate}
            onCreate={() => setCreating(true)}
          />
        </div>
        <figure className="hero-shot">
          <img
            className="shot shot-light"
            src="/screenshot-light.webp"
            width="1920"
            height="1005"
            loading="lazy"
            alt="A Unison workspace: the backlog with in-progress tasks claimed by team members and agents, and a sidebar with screens, active teammates and design links."
          />
          <img
            className="shot shot-dark"
            src="/screenshot-dark.webp"
            width="1920"
            height="1005"
            loading="lazy"
            alt="A Unison workspace: the backlog with in-progress tasks claimed by team members and agents, and a sidebar with screens, active teammates and design links."
          />
        </figure>
        <section className="how" aria-labelledby="how-title">
          <h2 id="how-title" className="section-title">
            Why Unison
          </h2>
          <ul className="features">
            {FEATURES.map(({ icon: Icon, title, text }) => (
              <li key={title} className="feature">
                <span className="feature-icon" aria-hidden="true">
                  <Icon />
                </span>
                <h3>{title}</h3>
                <p>{text}</p>
              </li>
            ))}
          </ul>
        </section>
      </main>
      {creating && <NewWorkspaceDialog account={account} navigate={navigate} onClose={() => setCreating(false)} />}
      <footer className="site-foot">
        <a href="/agent.md">Agents: read /agent.md</a>
        <a href={REPO_URL}>
          <Github aria-hidden="true" />
          Source on GitHub
        </a>
        <LegalLinks />
      </footer>
    </div>
  );
}
