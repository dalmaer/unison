import {
  ChevronDown,
  ChevronRight,
  FolderGit2,
  GitPullRequest,
  Globe,
  LayoutGrid,
  Lightbulb,
  Link2,
  ListTodo,
  MessageCircle,
  MessageSquare,
  Plus,
  Repeat,
  Settings,
} from "lucide-react";

const PREVIEW_HUMANS = [
  { emoji: "🦊", name: "Alex", state: "active" },
  { emoji: "🦜", name: "Zack", state: "dormant" },
];

const PREVIEW_AGENTS = [
  { emoji: "🦅", name: "Kestrel", model: "Gemini" },
  { emoji: "🦈", name: "Mako", model: "Gemini" },
];

const PREVIEW_IN_PROGRESS = [
  { id: 327, pos: 1, title: "Live theme-reactive workspace preview on landing page", reporter: "Tern", assignee: "🦅 Kestrel", pr: "#174", comments: 2, meta: "11h left" },
  { id: 307, pos: 2, title: "Split sidebar roster into Humans and Agents", reporter: "Tern", assignee: "🦈 Mako", pr: "#175", comments: 1, meta: "10h left" },
];

const PREVIEW_OPEN = [
  { id: 308, pos: 3, title: "Add styled participant @mention chips and autocomplete", reporter: "Tern", comments: 3, meta: "2d ago" },
  { id: 310, pos: 4, title: "Paginate chat messages and activity with upward scroll", reporter: "Tern", meta: "2d ago" },
  { id: 315, pos: 5, title: "Add first-use interactive tooltip tour for new members", reporter: "Alex", meta: "3d ago" },
];

// HeroShot renders a non-interactive, theme-reactive preview of a Unison
// workspace using the app's real .shell, .sidebar, and Backlog .screen markup (#327).
export default function HeroShot() {
  return (
    <div className="shell hero-preview-shell" aria-hidden="true">
      <nav className="sidebar">
        <div className="sidebar-head">
          <span className="ws-mark">U</span>
          <span className="ws-name">
            <span className="ws-switcher-trigger">
              Unison
              <ChevronDown className="lucide" />
            </span>
          </span>
          <span className="live live-on" />
        </div>
        <div className="sidebar-content">
          <ul className="nav-list">
            <li>
              <span className="nav-link">
                <Lightbulb className="lucide" />
                Ideas
              </span>
            </li>
            <li>
              <span className="nav-link" aria-current="page">
                <ListTodo className="lucide" />
                Backlog
                <span className="nav-count">5</span>
              </span>
            </li>
            <li>
              <span className="nav-link">
                <Repeat className="lucide" />
                Recurring
              </span>
            </li>
            <li>
              <span className="nav-link">
                <MessageCircle className="lucide" />
                Chat
                <span className="nav-unread" />
              </span>
            </li>
          </ul>

          <section className="sidebar-section">
            <div className="nav-heading nav-heading-button">
              <ChevronDown className="lucide" />
              Humans
              <span className="nav-heading-count">2</span>
            </div>
            <ul className="roster">
              {PREVIEW_HUMANS.map((p) => (
                <li key={p.name} className={`roster-row ${p.state}`}>
                  <span className="roster-emoji">{p.emoji}</span>
                  <span className="roster-name">{p.name}</span>
                  <span className="presence" />
                </li>
              ))}
            </ul>
            <span className="nav-link nav-link-action">
              <Plus className="lucide" />
              Invite
            </span>
          </section>

          <section className="sidebar-section">
            <div className="nav-heading nav-heading-button">
              <ChevronDown className="lucide" />
              Agents
              <span className="nav-heading-count">2</span>
            </div>
            <ul className="roster">
              {PREVIEW_AGENTS.map((p) => (
                <li key={p.name} className="roster-row active">
                  <span className="roster-emoji">{p.emoji}</span>
                  <span className="roster-name">
                    {p.name} <span className="roster-model">{p.model}</span>
                  </span>
                  <span className="presence" />
                </li>
              ))}
            </ul>
            <span className="nav-link nav-link-action">
              <Plus className="lucide" />
              Invite
            </span>
          </section>

          <section className="sidebar-section">
            <div className="nav-heading">Links</div>
            <ul className="nav-list">
              <li>
                <span className="nav-link">
                  <Globe className="lucide" />
                  Product
                </span>
              </li>
              <li>
                <span className="nav-link">
                  <LayoutGrid className="lucide" />
                  Canvas
                </span>
              </li>
              <li>
                <span className="nav-link">
                  <FolderGit2 className="lucide" />
                  digitalex/unison
                </span>
              </li>
              <li>
                <span className="nav-link">
                  <Link2 className="lucide" />
                  Design doc
                </span>
              </li>
            </ul>
            <span className="nav-link nav-link-action">
              <Plus className="lucide" />
              Add link
            </span>
          </section>
        </div>
        <div className="sidebar-foot">
          <span className="nav-link">
            <Settings className="lucide" />
            Settings
          </span>
        </div>
      </nav>

      <main className="shell-main panel">
        <section className="screen">
          <header className="screen-head">
            <h2 className="screen-title">Backlog</h2>
            <div className="seg view-switch">
              <button type="button" tabIndex={-1} aria-pressed="true">
                List
              </button>
              <button type="button" tabIndex={-1} aria-pressed="false">
                Board
              </button>
            </div>
            <span className="spacer" />
            <span className="button button-primary button-small">
              <Plus className="lucide" />
              <span className="label">New item</span>
            </span>
          </header>

          <div className="backlog-scroll">
            <div className="group">
              <div className="group-head">
                In progress <span className="count">2</span>
              </div>
              {PREVIEW_IN_PROGRESS.map((it) => (
                <div key={it.id} className="item in_progress">
                  <div className="row">
                    <span className="rank">
                      <span className="rank-number">{it.pos}</span>
                    </span>
                    <span className="status in_progress" />
                    <span className="item-id">#{it.id}</span>
                    <span className="row-title">{it.title}</span>
                    <span className="pr-chips">
                      <span className="chip pr-chip">
                        <GitPullRequest className="lucide" />
                        {it.pr}
                      </span>
                    </span>
                    {it.comments > 0 && (
                      <span className="comment-count">
                        <MessageSquare className="lucide" />
                        {it.comments}
                      </span>
                    )}
                    <span className="item-reporter">by {it.reporter}</span>
                    <span className="chip">{it.assignee}</span>
                    <span className="meta">{it.meta}</span>
                  </div>
                </div>
              ))}
            </div>

            <div className="group">
              <div className="group-head">
                Open <span className="count">3</span>
              </div>
              {PREVIEW_OPEN.map((it) => (
                <div key={it.id} className="item open">
                  <div className="row">
                    <span className="rank">
                      <span className="rank-number">{it.pos}</span>
                    </span>
                    <span className="status open" />
                    <span className="item-id">#{it.id}</span>
                    <span className="row-title">{it.title}</span>
                    {it.comments > 0 && (
                      <span className="comment-count">
                        <MessageSquare className="lucide" />
                        {it.comments}
                      </span>
                    )}
                    <span className="item-reporter">by {it.reporter}</span>
                    <span className="meta">{it.meta}</span>
                  </div>
                </div>
              ))}
            </div>

            <div className="group">
              <div className="group-head group-toggle">
                <ChevronRight className="lucide" />
                Done <span className="count">14</span>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
