import {
  Box,
  createCliRenderer,
  ScrollBox,
  Select,
  TabSelectRenderable,
  Text,
  TextAttributes,
  type CliRenderer,
  type KeyEvent,
} from "@opentui/core";
import type { AzureDevOpsClient } from "./azureDevOps.js";
import { openUrl } from "./terminal.js";
import type { AppConfig, DashboardItem, LoadedSection, PullRequestReview } from "./types.js";

interface ReviewState {
  item: DashboardItem;
  loading: boolean;
  selectedFile: number;
  selectedTab: number;
  data?: PullRequestReview;
  error?: string;
}

const COLORS = {
  bg: "#0d1117",
  panel: "#161b22",
  panel2: "#1f2630",
  border: "#30363d",
  selected: "#264f78",
  text: "#c9d1d9",
  muted: "#8b949e",
  cyan: "#7dd3fc",
  blue: "#58a6ff",
  green: "#3fb950",
  yellow: "#d29922",
  red: "#f85149",
  purple: "#a371f7",
};

const REVIEW_TABS = [" Overview", " Activity", " Files Changed"];

export class OpenTuiDashboard {
  private renderer?: CliRenderer;
  private sections: LoadedSection[];
  private selectedSection = 0;
  private selectedRow = 0;
  private status = "";
  private stopped = false;
  private refreshTimer?: NodeJS.Timeout;
  private resolveRun?: () => void;
  private review?: ReviewState;

  constructor(
    private readonly config: AppConfig,
    private readonly client: AzureDevOpsClient,
  ) {
    this.sections = config.sections.map((section) => ({ config: section, items: [], loading: false }));
  }

  async run(): Promise<void> {
    this.renderer = await createCliRenderer({
      exitOnCtrlC: false,
      clearOnShutdown: true,
      screenMode: "alternate-screen",
      targetFps: 30,
      backgroundColor: COLORS.bg,
      consoleMode: "disabled",
    });

    this.renderer.setBackgroundColor(COLORS.bg);
    this.renderer.keyInput.on("keypress", (key: KeyEvent) => void this.handleKey(key));
    this.renderer.on("resize", () => this.render());

    const refreshSeconds = this.config.defaults?.refreshSeconds ?? 120;
    this.refreshTimer = setInterval(() => void this.refreshCurrent(false), refreshSeconds * 1000);

    this.render();
    this.renderer.start();
    void this.refreshAll();

    return new Promise<void>((resolve) => {
      this.resolveRun = resolve;
    });
  }

  stop(): void {
    if (this.stopped) return;
    this.stopped = true;
    if (this.refreshTimer) clearInterval(this.refreshTimer);
    this.renderer?.destroy();
    this.resolveRun?.();
  }

  private async handleKey(key: KeyEvent): Promise<void> {
    const name = key.name;
    const seq = key.sequence;

    if ((key.ctrl && name === "c") || name === "q") {
      this.stop();
      return;
    }

    if (this.review) {
      await this.handleReviewKey(name, seq);
      return;
    }

    if (name === "tab" || name === "right" || seq === "l") {
      this.selectedSection = (this.selectedSection + 1) % this.sections.length;
      this.selectedRow = 0;
      this.render();
      if (!this.current.items.length && !this.current.loading) await this.refreshCurrent(false);
      return;
    }

    if (name === "left" || seq === "h") {
      this.selectedSection = (this.selectedSection - 1 + this.sections.length) % this.sections.length;
      this.selectedRow = 0;
      this.render();
      if (!this.current.items.length && !this.current.loading) await this.refreshCurrent(false);
      return;
    }

    if (name === "down" || seq === "j") {
      this.selectedRow = Math.min(this.selectedRow + 1, Math.max(0, this.current.items.length - 1));
      this.render();
      return;
    }

    if (name === "up" || seq === "k") {
      this.selectedRow = Math.max(this.selectedRow - 1, 0);
      this.render();
      return;
    }

    if (seq === "r") {
      await this.refreshCurrent(true);
      return;
    }

    if (seq === "R") {
      await this.refreshAll(true);
      return;
    }

    if (seq === "v") {
      await this.openReview();
      return;
    }

    if (name === "return" || seq === "o") {
      this.openSelected("default");
    }
  }

  private async handleReviewKey(name: string, seq: string): Promise<void> {
    if (!this.review) return;

    if (name === "escape" || seq === "b" || seq === "v") {
      this.review = undefined;
      this.render();
      return;
    }

    if (name === "right" || seq === "l" || seq === "]") {
      this.review.selectedTab = (this.review.selectedTab + 1) % REVIEW_TABS.length;
      this.render();
      return;
    }

    if (name === "left" || seq === "h" || seq === "[") {
      this.review.selectedTab = (this.review.selectedTab - 1 + REVIEW_TABS.length) % REVIEW_TABS.length;
      this.render();
      return;
    }

    const files = this.review.data?.files ?? [];
    if (this.review.selectedTab === 2 && (name === "down" || seq === "j")) {
      this.review.selectedFile = Math.min(this.review.selectedFile + 1, Math.max(0, files.length - 1));
      this.render();
      return;
    }

    if (this.review.selectedTab === 2 && (name === "up" || seq === "k")) {
      this.review.selectedFile = Math.max(this.review.selectedFile - 1, 0);
      this.render();
      return;
    }

    if (name === "return" || seq === "o") this.openSelected("review");
  }

  private get current(): LoadedSection {
    return this.sections[this.selectedSection]!;
  }

  private async refreshAll(force = false): Promise<void> {
    await Promise.all(this.sections.map((_section, index) => this.refreshSection(index, force)));
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
      section.items = section.config.type === "pullRequests"
        ? await this.client.listPullRequests(section.config, limit)
        : await this.client.listWorkItems(section.config, limit);
      section.lastLoadedAt = new Date();
      this.selectedRow = Math.min(this.selectedRow, Math.max(0, section.items.length - 1));
      this.status = `Loaded ${section.items.length} item(s) for ${section.config.title}`;
    } catch (error) {
      section.error = error instanceof Error ? error.message : String(error);
      this.status = `Failed loading ${section.config.title}`;
    } finally {
      section.loading = false;
      this.render();
    }
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

  private openSelected(mode: "default" | "review"): void {
    const item = this.current.items[this.selectedRow] ?? this.review?.item;
    const url = mode === "review" ? item?.reviewUrl : item?.url;
    if (url) {
      openUrl(url);
      this.status = mode === "review" ? `Opened browser review` : `Opened ${item?.id ?? "item"}`;
    } else {
      this.status = mode === "review" ? "No review URL for selected item" : "No URL for selected item";
    }
    this.render();
  }

  private render(): void {
    if (this.stopped || !this.renderer) return;
    const previous = this.renderer.root.getRenderable("app");
    if (previous) this.renderer.root.remove("app");

    this.renderer.root.add(
      Box(
        {
          id: "app",
          width: "100%",
          height: "100%",
          flexDirection: "column",
          backgroundColor: COLORS.bg,
        },
        this.review ? this.reviewScreen() : this.dashboardScreen(),
        this.footer(),
      ),
    );
    this.renderer.requestRender();
  }

  private dashboardScreen() {
    const selected = this.current.items[this.selectedRow];
    return Box(
      { width: "100%", flexGrow: 1, flexDirection: "column", gap: 1, padding: 1 },
      this.header("ado-dash", this.config.project),
      this.sectionTabs(),
      Box(
        { width: "100%", flexGrow: 1, flexDirection: "row", gap: 1 },
        this.itemsPane(),
        this.previewPane(selected),
      ),
    );
  }

  private reviewScreen() {
    const review = this.review!;
    const data = review.data;
    return Box(
      { width: "100%", flexGrow: 1, flexDirection: "column", padding: 1, gap: 1 },
      this.previewHeader(review.item),
      this.reviewTabs(),
      Box(
        { width: "100%", flexGrow: 1, border: true, borderStyle: "rounded", borderColor: COLORS.border, padding: 1 },
        review.loading
          ? Text({ content: "Loading PR details…", fg: COLORS.yellow })
          : review.error
            ? Text({ content: review.error, fg: COLORS.red, wrapMode: "word" })
            : this.reviewBody(data!, review),
      ),
    );
  }

  private header(title: string, subtitle: string) {
    return Box(
      {
        width: "100%",
        height: 3,
        border: true,
        borderStyle: "rounded",
        borderColor: COLORS.border,
        backgroundColor: COLORS.panel,
        paddingX: 1,
        alignItems: "center",
      },
      Text({ content: `${title}  ${subtitle}`, fg: COLORS.cyan, attributes: TextAttributes.BOLD }),
    );
  }

  private sectionTabs() {
    const tabs = new TabSelectRenderable(this.renderer!, {
      id: "section-tabs",
      width: "100%",
      height: 3,
      options: this.sections.map((section) => ({
        name: `${section.config.title} ${section.loading ? "…" : section.items.length}`,
        description: section.config.type,
      })),
      tabWidth: 24,
      showDescription: false,
      showUnderline: true,
      wrapSelection: true,
      selectedBackgroundColor: COLORS.selected,
      selectedTextColor: "#ffffff",
      backgroundColor: COLORS.bg,
      textColor: COLORS.muted,
      focusedBackgroundColor: COLORS.selected,
      focusedTextColor: "#ffffff",
    });
    tabs.setSelectedIndex(this.selectedSection);
    return tabs;
  }

  private itemsPane() {
    const current = this.current;
    const content = current.error
      ? Text({ content: current.error, fg: COLORS.red, wrapMode: "word" })
      : current.loading && current.items.length === 0
        ? Text({ content: "Loading…", fg: COLORS.yellow })
        : current.items.length === 0
          ? Text({ content: "No items.", fg: COLORS.muted })
          : Select({
              width: "100%",
              height: "100%",
              options: current.items.map((item) => ({
                name: `${item.id}  ${item.title}`,
                description: this.itemDescription(item),
                value: item,
              })),
              selectedIndex: this.selectedRow,
              showDescription: true,
              showScrollIndicator: true,
              wrapSelection: false,
              backgroundColor: COLORS.panel,
              textColor: COLORS.text,
              descriptionColor: COLORS.muted,
              selectedBackgroundColor: COLORS.selected,
              selectedTextColor: "#ffffff",
              selectedDescriptionColor: "#dbeafe",
              focusedBackgroundColor: COLORS.panel,
              focusedTextColor: COLORS.text,
            });

    return Box(
      {
        flexGrow: 2,
        height: "100%",
        border: true,
        borderStyle: "rounded",
        borderColor: COLORS.border,
        backgroundColor: COLORS.panel,
        padding: 1,
        title: current.config.title,
      },
      content,
    );
  }

  private itemDescription(item: DashboardItem): string {
    const who = item.author ?? item.assignee ?? "unassigned";
    const updated = item.updatedAt ? relativeTime(item.updatedAt) : "";
    return [item.state, item.subtitle, who, updated].filter(Boolean).join(" • ");
  }

  private previewPane(item: DashboardItem | undefined) {
    const content = item
      ? [
          item.id,
          item.title,
          "",
          item.subtitle ?? "",
          item.author ? `Author: ${item.author}` : item.assignee ? `Assignee: ${item.assignee}` : "",
          item.state ? `State: ${item.state}` : "",
          item.url ?? "",
          "",
          item.kind === "pullRequest" ? "Press v for local review" : "Press enter/o to open",
        ].filter(Boolean).join("\n")
      : "No selection";

    return Box(
      {
        flexGrow: 1,
        height: "100%",
        border: true,
        borderStyle: "rounded",
        borderColor: COLORS.border,
        backgroundColor: COLORS.panel,
        padding: 1,
        title: "Preview",
      },
      ScrollBox(
        {
          width: "100%",
          height: "100%",
          scrollY: true,
          scrollbarOptions: { trackOptions: { foregroundColor: COLORS.blue, backgroundColor: COLORS.panel2 } },
        },
        Text({ content, fg: COLORS.text, wrapMode: "word" }),
      ),
    );
  }

  private previewHeader(item: DashboardItem) {
    return Box(
      { width: "100%", flexDirection: "column" },
      Box(
        { width: "100%", height: 2, backgroundColor: COLORS.selected, paddingX: 1 },
        Text({ content: `${item.subtitle ?? this.config.project} · ${item.id}`, fg: COLORS.text }),
      ),
      Box(
        { width: "100%", height: 5, backgroundColor: COLORS.panel2, paddingX: 1, justifyContent: "center" },
        Text({ content: item.title, fg: COLORS.text, attributes: TextAttributes.BOLD, wrapMode: "word" }),
      ),
      Box(
        { width: "100%", height: 2, backgroundColor: COLORS.panel, paddingX: 1 },
        Text({ content: `${item.state ?? "active"}  ${item.author ? `by ${item.author}` : ""}`, fg: statusColor(item.state) }),
      ),
    );
  }

  private reviewTabs() {
    const tabs = new TabSelectRenderable(this.renderer!, {
      id: "review-tabs",
      width: "100%",
      height: 3,
      options: REVIEW_TABS.map((name) => ({ name, description: "" })),
      tabWidth: 22,
      showDescription: false,
      showUnderline: true,
      wrapSelection: true,
      selectedBackgroundColor: COLORS.selected,
      selectedTextColor: "#ffffff",
      backgroundColor: COLORS.bg,
      textColor: COLORS.muted,
      focusedBackgroundColor: COLORS.selected,
      focusedTextColor: "#ffffff",
    });
    tabs.setSelectedIndex(this.review?.selectedTab ?? 0);
    return tabs;
  }

  private reviewBody(data: PullRequestReview, review: ReviewState) {
    if (review.selectedTab === 1) return this.activityView(data);
    if (review.selectedTab === 2) return this.filesView(data, review.selectedFile);
    return this.overviewView(data);
  }

  private overviewView(data: PullRequestReview) {
    const files = data.files.map((file) => `${changeIcon(file.changeType)} ${pad(file.changeType, 8)} ${file.path}`);
    return Box(
      { width: "100%", height: "100%", flexDirection: "column", gap: 1 },
      Box(
        { width: "100%", border: true, borderStyle: "rounded", borderColor: COLORS.border, padding: 1 },
        Text({
          content: ` ${data.files.length} files changed\n ${data.threads.length} discussion thread(s)\n iteration ${data.iterationId ?? "?"}`,
          fg: COLORS.text,
        }),
      ),
      Text({ content: " Files", fg: COLORS.cyan, attributes: TextAttributes.BOLD }),
      ScrollBox(
        {
          width: "100%",
          flexGrow: 1,
          scrollY: true,
          scrollbarOptions: { trackOptions: { foregroundColor: COLORS.blue, backgroundColor: COLORS.panel2 } },
        },
        Text({ content: files.join("\n") || "No changed files", fg: COLORS.text, wrapMode: "none", truncate: true }),
      ),
    );
  }

  private activityView(data: PullRequestReview) {
    const content = data.threads.length === 0
      ? "No PR discussion threads."
      : data.threads.map((thread) => {
          const location = thread.filePath ? `${thread.filePath}${thread.line ? `:${thread.line}` : ""}` : "Conversation";
          const comments = thread.comments.map((comment) => `  ${comment.author ?? "unknown"}: ${oneLine(comment.content)}`).join("\n");
          return `#${thread.id} ${thread.status ?? "active"} ${location}\n${comments}`;
        }).join("\n\n");

    return ScrollBox(
      {
        width: "100%",
        height: "100%",
        scrollY: true,
        scrollbarOptions: { trackOptions: { foregroundColor: COLORS.blue, backgroundColor: COLORS.panel2 } },
      },
      Text({ content, fg: COLORS.text, wrapMode: "word" }),
    );
  }

  private filesView(data: PullRequestReview, selectedFile: number) {
    if (data.files.length === 0) return Text({ content: "No changed files.", fg: COLORS.muted });

    return Select({
      width: "100%",
      height: "100%",
      options: data.files.map((file) => ({
        name: `${changeIcon(file.changeType)} ${file.path}`,
        description: `${file.changeType}${file.changeTrackingId ? ` • tracking ${file.changeTrackingId}` : ""}`,
        value: file,
      })),
      selectedIndex: selectedFile,
      showDescription: true,
      showScrollIndicator: true,
      backgroundColor: COLORS.panel,
      textColor: COLORS.text,
      descriptionColor: COLORS.muted,
      selectedBackgroundColor: COLORS.selected,
      selectedTextColor: "#ffffff",
      selectedDescriptionColor: "#dbeafe",
      focusedBackgroundColor: COLORS.panel,
      focusedTextColor: COLORS.text,
    });
  }

  private footer() {
    const help = this.review
      ? "[/]/h/l tabs • j/k files • b/esc/v back • enter/o browser • q quit"
      : "tab/h/l sections • j/k move • r refresh • R refresh all • enter/o open • v review • q quit";
    return Box(
      { width: "100%", height: 2, backgroundColor: COLORS.panel2, paddingX: 1 },
      Text({ content: `${help} │ ${this.status}`, fg: COLORS.muted, truncate: true }),
    );
  }
}

function pad(value: string, width: number): string {
  if (value.length > width) return `${value.slice(0, Math.max(0, width - 1))}…`;
  return value.padEnd(width);
}

function changeIcon(changeType: string): string {
  switch (changeType.toLowerCase()) {
    case "add":
    case "added":
      return "";
    case "delete":
    case "deleted":
      return "";
    case "rename":
    case "renamed":
      return "";
    default:
      return "";
  }
}

function statusColor(status?: string): string {
  const normalized = status?.toLowerCase() ?? "";
  if (normalized.includes("active") || normalized.includes("open")) return COLORS.green;
  if (normalized.includes("closed") || normalized.includes("abandoned")) return COLORS.red;
  return COLORS.yellow;
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
