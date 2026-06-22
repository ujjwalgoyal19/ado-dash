import type { AzureDevOpsClient } from "./azureDevOps.js";
import { ansi, openUrl, setupKeypress, terminalSize, truncate } from "./terminal.js";
import type { AppConfig, DashboardItem, LoadedSection, PullRequestReview } from "./types.js";

interface ReviewState {
  item: DashboardItem;
  loading: boolean;
  selectedFile: number;
  selectedTab: number;
  data?: PullRequestReview;
  error?: string;
}

const REVIEW_TABS = [" Overview", " Activity", " Files Changed"];

export class Dashboard {
  private sections: LoadedSection[];
  private selectedSection = 0;
  private selectedRow = 0;
  private status = "";
  private stopped = false;
  private cleanupKeys?: () => void;
  private refreshTimer?: NodeJS.Timeout;
  private review?: ReviewState;

  constructor(
    private readonly config: AppConfig,
    private readonly client: AzureDevOpsClient,
  ) {
    this.sections = config.sections.map((section) => ({ config: section, items: [], loading: false }));
  }

  async run(): Promise<void> {
    process.on("resize", this.render);
    this.cleanupKeys = setupKeypress((key) => void this.handleKey(key));

    const refreshSeconds = this.config.defaults?.refreshSeconds ?? 120;
    this.refreshTimer = setInterval(() => void this.refreshCurrent(false), refreshSeconds * 1000);

    this.render();
    await this.refreshAll();
  }

  stop(): void {
    if (this.stopped) return;
    this.stopped = true;
    if (this.refreshTimer) clearInterval(this.refreshTimer);
    this.cleanupKeys?.();
    process.off("resize", this.render);
  }

  private handleKey = async (key: string): Promise<void> => {
    if (this.review && key !== "q" && key !== "ctrl+c") {
      this.handleReviewKey(key);
      return;
    }

    if (key === "q" || key === "escape" || key === "ctrl+c") {
      this.stop();
      return;
    }

    if (key === "tab" || key === "l" || key === "right") {
      this.selectedSection = (this.selectedSection + 1) % this.sections.length;
      this.selectedRow = 0;
      this.render();
      if (!this.current.items.length && !this.current.loading) await this.refreshCurrent(false);
      return;
    }

    if (key === "shift+tab" || key === "h" || key === "left") {
      this.selectedSection = (this.selectedSection - 1 + this.sections.length) % this.sections.length;
      this.selectedRow = 0;
      this.render();
      if (!this.current.items.length && !this.current.loading) await this.refreshCurrent(false);
      return;
    }

    if (key === "down" || key === "j") {
      this.selectedRow = Math.min(this.selectedRow + 1, Math.max(0, this.current.items.length - 1));
      this.render();
      return;
    }

    if (key === "up" || key === "k") {
      this.selectedRow = Math.max(this.selectedRow - 1, 0);
      this.render();
      return;
    }

    if (key === "r") {
      await this.refreshCurrent(true);
      return;
    }

    if (key === "shift+r") {
      await this.refreshAll();
      return;
    }

    if (key === "o" || key === "enter") {
      this.openSelected("default");
      return;
    }

    if (key === "v") {
      await this.openReview();
      return;
    }
  };

  private openSelected(mode: "default" | "review"): void {
    const item = this.current.items[this.selectedRow];
    const url = mode === "review" ? item?.reviewUrl : item?.url;
    if (url) {
      openUrl(url);
      this.status = mode === "review" ? `Opened review ${url}` : `Opened ${url}`;
    } else {
      this.status = mode === "review" ? "No review URL for selected item" : "No URL for selected item";
    }
    this.render();
  }

  private async openReview(): Promise<void> {
    const item = this.current.items[this.selectedRow];
    if (!item) return;
    if (item.kind !== "pullRequest") {
      this.status = "Local review is available for pull requests only";
      this.render();
      return;
    }

    this.review = { item, loading: true, selectedFile: 0, selectedTab: 0 };
    this.status = `Loading local review for ${item.id}…`;
    this.render();

    try {
      this.review.data = await this.client.getPullRequestReview(item);
      this.review.error = undefined;
      this.status = `Loaded local review for ${item.id}`;
    } catch (error) {
      this.review.error = error instanceof Error ? error.message : String(error);
      this.status = `Failed loading local review for ${item.id}`;
    } finally {
      this.review.loading = false;
      this.render();
    }
  }

  private handleReviewKey(key: string): void {
    if (!this.review) return;

    if (key === "escape" || key === "b" || key === "v") {
      this.review = undefined;
      this.render();
      return;
    }

    if (key === "right" || key === "l" || key === "]") {
      this.review.selectedTab = (this.review.selectedTab + 1) % REVIEW_TABS.length;
      this.render();
      return;
    }

    if (key === "left" || key === "h" || key === "[") {
      this.review.selectedTab = (this.review.selectedTab - 1 + REVIEW_TABS.length) % REVIEW_TABS.length;
      this.render();
      return;
    }

    const files = this.review.data?.files ?? [];
    if (this.review.selectedTab === 2 && (key === "down" || key === "j")) {
      this.review.selectedFile = Math.min(this.review.selectedFile + 1, Math.max(0, files.length - 1));
      this.render();
      return;
    }

    if (this.review.selectedTab === 2 && (key === "up" || key === "k")) {
      this.review.selectedFile = Math.max(this.review.selectedFile - 1, 0);
      this.render();
      return;
    }

    if (key === "o" || key === "enter") {
      this.openSelected("review");
    }
  }

  private get current(): LoadedSection {
    return this.sections[this.selectedSection]!;
  }

  private async refreshAll(): Promise<void> {
    await Promise.all(this.sections.map((_section, index) => this.refreshSection(index, false)));
  }

  private async refreshCurrent(force: boolean): Promise<void> {
    await this.refreshSection(this.selectedSection, force);
  }

  private async refreshSection(index: number, force: boolean): Promise<void> {
    const section = this.sections[index]!;
    if (section.loading) return;
    if (!force && section.items.length > 0 && section.lastLoadedAt) return;

    section.loading = true;
    section.error = undefined;
    this.status = `Loading ${section.config.title}…`;
    this.render();

    try {
      const limit = this.config.defaults?.limit ?? 20;
      const items = section.config.type === "pullRequests"
        ? await this.client.listPullRequests(section.config, limit)
        : await this.client.listWorkItems(section.config, limit);

      section.items = items;
      section.lastLoadedAt = new Date();
      this.selectedRow = Math.min(this.selectedRow, Math.max(0, section.items.length - 1));
      this.status = `Loaded ${items.length} item(s) for ${section.config.title}`;
    } catch (error) {
      section.error = error instanceof Error ? error.message : String(error);
      this.status = `Failed loading ${section.config.title}`;
    } finally {
      section.loading = false;
      this.render();
    }
  }

  private render = (): void => {
    if (this.stopped) return;
    const { columns, rows } = terminalSize();
    const width = Math.max(columns, 80);
    const height = Math.max(rows, 20);
    if (this.review) {
      process.stdout.write(ansi.clear + this.renderReview(width, height));
      return;
    }

    const current = this.current;
    const bodyHeight = Math.max(5, height - 10);
    const selected = current.items[this.selectedRow];

    const lines: string[] = [];
    lines.push(`${ansi.bold}${ansi.cyan}ado-dash${ansi.reset} ${ansi.dim}${this.config.project}${ansi.reset}`);
    lines.push(this.renderTabs(width));
    lines.push("─".repeat(width));

    if (current.error) {
      lines.push(`${ansi.red}Error:${ansi.reset} ${current.error}`.slice(0, width));
    } else if (current.loading && current.items.length === 0) {
      lines.push(`${ansi.yellow}Loading…${ansi.reset}`);
    } else if (current.items.length === 0) {
      lines.push(`${ansi.dim}No items.${ansi.reset}`);
    } else {
      lines.push(this.renderHeader(width));
      for (let i = 0; i < Math.min(bodyHeight, current.items.length); i++) {
        lines.push(this.renderRow(current.items[i]!, i, width));
      }
    }

    while (lines.length < height - 5) lines.push("");
    lines.push("─".repeat(width));
    lines.push(this.renderPreview(selected, width));
    lines.push("─".repeat(width));
    lines.push(this.renderFooter(current));

    process.stdout.write(ansi.clear + lines.slice(0, height).join("\n"));
  };

  private renderReview(width: number, height: number): string {
    const review = this.review!;
    const lines: string[] = [];

    lines.push(this.previewHeader(width, `${review.item.subtitle ?? this.config.project} · ${review.item.id}`));
    lines.push(...this.previewTitle(width, review.item.title));
    lines.push(this.reviewMetaLine(review.item, width));
    lines.push("");
    lines.push(this.reviewTabs(width, review.selectedTab));
    lines.push("─".repeat(width));

    if (review.loading) {
      lines.push(`${ansi.yellow}Loading PR details…${ansi.reset}`);
    } else if (review.error) {
      lines.push(`${ansi.red}Error:${ansi.reset} ${review.error}`.slice(0, width));
    } else if (review.data) {
      const bodyHeight = Math.max(4, height - lines.length - 3);
      const body = this.renderReviewTab(review, width, bodyHeight);
      lines.push(...body);
    }

    while (lines.length < height - 2) lines.push("");
    lines.push("─".repeat(width));
    lines.push(`${ansi.dim}[/]/h/l tabs • j/k move files • b/esc/v back • enter/o browser review • q quit | ${this.status}${ansi.reset}`);
    return lines.slice(0, height).join("\n");
  }

  private renderReviewTab(review: ReviewState, width: number, height: number): string[] {
    switch (review.selectedTab) {
      case 1:
        return this.renderReviewActivity(review.data!, width, height);
      case 2:
        return this.renderReviewFiles(review, width, height);
      default:
        return this.renderReviewOverview(review.data!, width, height);
    }
  }

  private renderReviewOverview(data: PullRequestReview, width: number, height: number): string[] {
    const lines: string[] = [];
    lines.push(`${ansi.bold}${ansi.cyan} Changes${ansi.reset}`);
    lines.push(this.box(width, [
      `${ansi.dim}${ansi.reset} ${data.files.length} files changed`,
      `${ansi.dim}${ansi.reset} ${data.threads.length} discussion thread(s)`,
      `${ansi.dim}${ansi.reset} iteration ${data.iterationId ?? "?"}`,
    ]));
    lines.push("");
    lines.push(`${ansi.bold}${ansi.cyan} Files${ansi.reset}`);
    for (const file of data.files.slice(0, Math.max(0, height - lines.length - 1))) {
      lines.push(this.fileLine(file, width));
    }
    if (data.files.length === 0) lines.push(`${ansi.dim}No changed files returned by Azure DevOps.${ansi.reset}`);
    return lines.slice(0, height);
  }

  private renderReviewActivity(data: PullRequestReview, width: number, height: number): string[] {
    const lines: string[] = [];
    if (data.threads.length === 0) return [`${ansi.dim}No PR discussion threads.${ansi.reset}`];

    for (const thread of data.threads) {
      const location = thread.filePath ? `${thread.filePath}${thread.line ? `:${thread.line}` : ""}` : "Conversation";
      lines.push(`${ansi.yellow}#${thread.id}${ansi.reset} ${this.statusPill(thread.status ?? "active")} ${truncate(location, width - 18)}`);
      for (const comment of thread.comments) {
        lines.push(`  ${ansi.cyan}${truncate(comment.author ?? "unknown", 20)}${ansi.reset} ${truncate(oneLine(comment.content), width - 24)}`);
        if (lines.length >= height) return lines;
      }
      lines.push("");
      if (lines.length >= height) return lines;
    }
    return lines.slice(0, height);
  }

  private renderReviewFiles(review: ReviewState, width: number, height: number): string[] {
    const files = review.data?.files ?? [];
    if (files.length === 0) return [`${ansi.dim}No changed files.${ansi.reset}`];

    const lines: string[] = [];
    for (let i = 0; i < Math.min(height, files.length); i++) {
      const line = this.fileLine(files[i]!, width);
      lines.push(i === review.selectedFile ? `${ansi.inverse}${line}${ansi.reset}` : line);
    }
    return lines;
  }

  private previewHeader(width: number, text: string): string {
    return `${ansi.inverse}${truncate(` ${text}`, width)}${ansi.reset}`;
  }

  private previewTitle(width: number, title: string): string[] {
    const inner = truncate(title, Math.max(1, width - 2));
    return [
      `${ansi.inverse}${" ".repeat(width)}${ansi.reset}`,
      `${ansi.inverse} ${ansi.bold}${inner}${ansi.reset}${ansi.inverse}${" ".repeat(Math.max(0, width - inner.length - 1))}${ansi.reset}`,
      `${ansi.inverse}${" ".repeat(width)}${ansi.reset}`,
    ];
  }

  private reviewMetaLine(item: DashboardItem, width: number): string {
    const status = this.statusPill(item.state ?? "active");
    const author = item.author ? `${ansi.dim}by${ansi.reset} ${item.author}` : "";
    const branches = item.subtitle ? `${ansi.dim}${item.subtitle}${ansi.reset}` : "";
    return truncate(` ${status}  ${branches}  ${author}`, width);
  }

  private reviewTabs(width: number, selected: number): string {
    const rendered = REVIEW_TABS.map((tab, index) => {
      const label = ` ${tab} `;
      return index === selected ? `${ansi.inverse}${label}${ansi.reset}` : `${ansi.dim}${label}${ansi.reset}`;
    }).join(" ");
    return rendered.slice(0, width + REVIEW_TABS.length * 12);
  }

  private box(width: number, body: string[]): string {
    const innerWidth = Math.max(4, width - 4);
    const top = `╭${"─".repeat(innerWidth + 2)}╮`;
    const middle = body.map((line) => `│ ${truncate(line, innerWidth)} │`).join("\n");
    const bottom = `╰${"─".repeat(innerWidth + 2)}╯`;
    return `${top}\n${middle}\n${bottom}`;
  }

  private fileLine(file: PullRequestReview["files"][number], width: number): string {
    const icon = changeIcon(file.changeType);
    const change = truncate(file.changeType.toLowerCase(), 10);
    return truncate(` ${icon} ${change} ${file.path}`, width);
  }

  private statusPill(status: string): string {
    const normalized = status.toLowerCase();
    const color = normalized.includes("active") || normalized.includes("open")
      ? ansi.green
      : normalized.includes("closed") || normalized.includes("abandoned")
        ? ansi.red
        : ansi.yellow;
    return `${color}${ansi.bold} ${status} ${ansi.reset}`;
  }

  private renderTabs(width: number): string {
    const tabs = this.sections.map((section, index) => {
      const count = section.loading ? "…" : String(section.items.length);
      const label = ` ${section.config.title} ${count} `;
      return index === this.selectedSection ? `${ansi.inverse}${label}${ansi.reset}` : label;
    }).join(" ");
    return tabs.slice(0, width);
  }

  private renderHeader(width: number): string {
    const titleWidth = Math.max(20, width - 58);
    return `${ansi.dim}${truncate("ID", 8)} ${truncate("STATE", 12)} ${truncate("TITLE", titleWidth)} ${truncate("WHO", 18)} ${truncate("UPDATED", 14)}${ansi.reset}`;
  }

  private renderRow(item: DashboardItem, index: number, width: number): string {
    const titleWidth = Math.max(20, width - 58);
    const who = item.author ?? item.assignee ?? "";
    const updated = item.updatedAt ? relativeTime(item.updatedAt) : "";
    const row = `${truncate(item.id, 8)} ${truncate(item.state, 12)} ${truncate(item.title, titleWidth)} ${truncate(who, 18)} ${truncate(updated, 14)}`;
    return index === this.selectedRow ? `${ansi.inverse}${row}${ansi.reset}` : row;
  }

  private renderPreview(item: DashboardItem | undefined, width: number): string {
    if (!item) return ansi.dim + "No selection" + ansi.reset;
    const pieces = [
      `${ansi.bold}${item.id}${ansi.reset}`,
      item.title,
      item.subtitle ? `${ansi.dim}${item.subtitle}${ansi.reset}` : undefined,
      item.url ? `${ansi.blue}${item.url}${ansi.reset}` : undefined,
      item.reviewUrl ? `${ansi.dim}review: v${ansi.reset}` : undefined,
    ].filter(Boolean).join("  ");
    return pieces.slice(0, width);
  }

  private renderFooter(section: LoadedSection): string {
    const loaded = section.lastLoadedAt ? `last loaded ${section.lastLoadedAt.toLocaleTimeString()}` : "not loaded";
    return `${ansi.dim}tab/h/l switch • j/k move • r refresh • R refresh all • enter/o open • v review files • q quit | ${loaded} | ${this.status}${ansi.reset}`;
  }
}

function changeIcon(changeType: string): string {
  switch (changeType.toLowerCase()) {
    case "add":
    case "added":
      return `${ansi.green}${ansi.reset}`;
    case "delete":
    case "deleted":
      return `${ansi.red}${ansi.reset}`;
    case "rename":
    case "renamed":
      return `${ansi.yellow}${ansi.reset}`;
    case "edit":
    case "modified":
    case "changed":
      return `${ansi.yellow}${ansi.reset}`;
    default:
      return "";
  }
}

function oneLine(value: string): string {
  return value.replace(/\s+/g, " ").trim();
}

function relativeTime(value: string): string {
  const time = Date.parse(value);
  if (!Number.isFinite(time)) return value;
  const seconds = Math.floor((Date.now() - time) / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 48) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}
