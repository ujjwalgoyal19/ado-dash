import type { AuthProvider } from "./auth.js";
import type {
  AppConfig,
  DashboardItem,
  Identity,
  PullRequestReview,
  PullRequestSectionConfig,
  WorkItemSectionConfig,
} from "./types.js";

interface ConnectionDataResponse {
  authenticatedUser?: {
    id?: string;
    displayName?: string;
    providerDisplayName?: string;
    properties?: {
      Account?: { $value?: string };
    };
  };
}

interface PullRequestResponse {
  value?: AzurePullRequest[];
}

interface AzurePullRequest {
  pullRequestId: number;
  title: string;
  status: string;
  creationDate?: string;
  closedDate?: string;
  createdBy?: AzureIdentityRef;
  repository?: { id?: string; name?: string; webUrl?: string };
  sourceRefName?: string;
  targetRefName?: string;
  url?: string;
  _links?: { web?: { href?: string } };
}

interface AzureIdentityRef {
  id?: string;
  displayName?: string;
  uniqueName?: string;
}

interface WiqlResponse {
  workItems?: { id: number; url: string }[];
}

interface PullRequestIterationsResponse {
  value?: AzurePullRequestIteration[];
}

interface AzurePullRequestIteration {
  id: number;
  description?: string;
}

interface PullRequestIterationChangesResponse {
  changeEntries?: AzurePullRequestChange[];
}

interface AzurePullRequestChange {
  changeTrackingId?: number;
  changeType?: string;
  item?: { path?: string };
}

interface PullRequestThreadsResponse {
  value?: AzurePullRequestThread[];
}

interface AzurePullRequestThread {
  id: number;
  status?: string;
  threadContext?: {
    filePath?: string;
    rightFileStart?: { line?: number; offset?: number };
    rightFileEnd?: { line?: number; offset?: number };
  };
  comments?: AzurePullRequestComment[];
}

interface AzurePullRequestComment {
  author?: { displayName?: string };
  content?: string;
  commentType?: string;
  isDeleted?: boolean;
}

interface WorkItemsResponse {
  value?: AzureWorkItem[];
}

interface AzureWorkItem {
  id: number;
  rev?: number;
  fields?: Record<string, unknown>;
  url?: string;
  _links?: { html?: { href?: string } };
}

export class AzureDevOpsClient {
  private readonly baseUrl: string;
  private readonly project: string;
  private me?: Identity;

  constructor(
    config: AppConfig,
    private readonly authProvider: AuthProvider,
  ) {
    this.baseUrl = config.organizationUrl.replace(/\/+$/, "");
    this.project = config.project;
  }

  async getMe(): Promise<Identity> {
    if (this.me) return this.me;

    const data = await this.request<ConnectionDataResponse>("/_apis/connectionData?api-version=7.1-preview.1");
    const user = data.authenticatedUser;
    if (!user?.id) throw new Error("Azure DevOps did not return authenticated user identity");

    this.me = {
      id: user.id,
      displayName: user.displayName ?? user.providerDisplayName ?? user.id,
      uniqueName: user.properties?.Account?.$value,
    };
    return this.me;
  }

  async checkConnection(): Promise<{ user: Identity; projectName?: string }> {
    const user = await this.getMe();
    const project = await this.request<{ name?: string }>(
      `/_apis/projects/${encodeURIComponent(this.project)}?api-version=7.1`,
    );
    return { user, projectName: project.name };
  }

  async listPullRequests(section: PullRequestSectionConfig, defaultLimit: number): Promise<DashboardItem[]> {
    const limit = section.limit ?? defaultLimit;
    const params = new URLSearchParams({
      "api-version": "7.1",
      "$top": String(limit),
      "searchCriteria.includeLinks": "true",
    });

    if (section.status && section.status !== "all") params.set("searchCriteria.status", section.status);
    if (section.repositoryId) params.set("searchCriteria.repositoryId", section.repositoryId);
    if (section.targetBranch) params.set("searchCriteria.targetRefName", normalizeRef(section.targetBranch));
    if (section.sourceBranch) params.set("searchCriteria.sourceRefName", normalizeRef(section.sourceBranch));

    if (section.role === "author" || section.role === "reviewer") {
      const me = await this.getMe();
      params.set(section.role === "author" ? "searchCriteria.creatorId" : "searchCriteria.reviewerId", me.id);
    }

    const data = await this.request<PullRequestResponse>(`/${encodeURIComponent(this.project)}/_apis/git/pullrequests?${params}`);
    return (data.value ?? []).map((pr) => {
      const url = pullRequestWebUrl(pr, this.baseUrl, this.project);
      return {
        id: `PR ${pr.pullRequestId}`,
        title: pr.title,
        state: pr.status,
        author: pr.createdBy?.displayName,
        subtitle: [pr.repository?.name, shortRef(pr.sourceRefName), "→", shortRef(pr.targetRefName)]
          .filter(Boolean)
          .join(" "),
        updatedAt: pr.closedDate ?? pr.creationDate,
        url,
        reviewUrl: url ? reviewFilesUrl(url) : undefined,
        kind: "pullRequest",
        pullRequest: {
          pullRequestId: pr.pullRequestId,
          repositoryId: pr.repository?.id,
          repositoryName: pr.repository?.name,
        },
        raw: pr,
      };
    });
  }

  async getPullRequestReview(item: DashboardItem): Promise<PullRequestReview> {
    if (!item.pullRequest) throw new Error("Selected item is not a pull request");
    const repo = item.pullRequest.repositoryId ?? item.pullRequest.repositoryName;
    if (!repo) throw new Error("Selected pull request has no repository id/name");

    const repoPath = `/${encodeURIComponent(this.project)}/_apis/git/repositories/${encodeURIComponent(repo)}`;
    const prPath = `${repoPath}/pullRequests/${item.pullRequest.pullRequestId}`;

    const iterations = await this.request<PullRequestIterationsResponse>(`${prPath}/iterations?api-version=7.1`);
    const latestIteration = [...(iterations.value ?? [])].sort((a, b) => b.id - a.id)[0];
    if (!latestIteration) {
      return { title: item.title, files: [], threads: [], iterationId: undefined };
    }

    const changes = await this.request<PullRequestIterationChangesResponse>(
      `${prPath}/iterations/${latestIteration.id}/changes?$top=2000&api-version=7.1`,
    );
    const threads = await this.request<PullRequestThreadsResponse>(`${prPath}/threads?api-version=7.1`);

    return {
      title: item.title,
      iterationId: latestIteration.id,
      files: (changes.changeEntries ?? [])
        .filter((change) => change.item?.path)
        .map((change) => ({
          path: change.item?.path ?? "",
          changeType: change.changeType ?? "unknown",
          changeTrackingId: change.changeTrackingId,
        })),
      threads: (threads.value ?? [])
        .filter((thread) => !isSystemOnlyThread(thread))
        .map((thread) => ({
          id: thread.id,
          status: thread.status,
          filePath: thread.threadContext?.filePath,
          line: thread.threadContext?.rightFileStart?.line ?? thread.threadContext?.rightFileEnd?.line,
          comments: (thread.comments ?? [])
            .filter((comment) => !comment.isDeleted && comment.content)
            .map((comment) => ({
              author: comment.author?.displayName,
              content: comment.content ?? "",
              commentType: comment.commentType,
            })),
        }))
        .filter((thread) => thread.comments.length > 0),
    };
  }

  async listWorkItems(section: WorkItemSectionConfig, defaultLimit: number): Promise<DashboardItem[]> {
    const limit = section.limit ?? defaultLimit;
    const wiqlData = await this.request<WiqlResponse>(
      `/${encodeURIComponent(this.project)}/_apis/wit/wiql?api-version=7.1`,
      {
        method: "POST",
        body: JSON.stringify({ query: section.wiql }),
      },
    );

    const ids = (wiqlData.workItems ?? []).slice(0, limit).map((item) => item.id);
    if (ids.length === 0) return [];

    const fields = section.fields ?? [
      "System.Id",
      "System.Title",
      "System.State",
      "System.WorkItemType",
      "System.AssignedTo",
      "System.ChangedDate",
    ];

    const params = new URLSearchParams({
      ids: ids.join(","),
      fields: fields.join(","),
      "api-version": "7.1",
    });

    const workItems = await this.request<WorkItemsResponse>(`/_apis/wit/workitems?${params}`);
    return (workItems.value ?? []).map((item) => {
      const f = item.fields ?? {};
      const assignedTo = asIdentityName(f["System.AssignedTo"]);
      return {
        id: String(item.id),
        title: String(f["System.Title"] ?? "Untitled"),
        state: stringField(f["System.State"]),
        subtitle: stringField(f["System.WorkItemType"]),
        assignee: assignedTo,
        updatedAt: stringField(f["System.ChangedDate"]),
        url: item._links?.html?.href ?? workItemWebUrl(this.baseUrl, this.project, item.id),
        kind: "workItem",
        raw: item,
      };
    });
  }

  private async request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const authorization = await this.authProvider.getAuthorizationHeader();
    const url = `${this.baseUrl}${path}`;
    const response = await fetch(url, {
      ...init,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        Authorization: authorization,
        ...(init.headers ?? {}),
      },
    });

    if (!response.ok) {
      const body = await response.text().catch(() => "");
      throw new Error(
        `Azure DevOps API ${response.status} ${response.statusText}\n` +
          `URL: ${url}\n` +
          `${explainHttpFailure(response.status)}\n` +
          `Response: ${body.slice(0, 500)}`,
      );
    }

    return response.json() as Promise<T>;
  }
}

function isSystemOnlyThread(thread: AzurePullRequestThread): boolean {
  const comments = thread.comments ?? [];
  return comments.length > 0 && comments.every((comment) => comment.commentType === "system");
}

function explainHttpFailure(status: number): string {
  if (status === 404) {
    return "Hint: check organizationUrl and project. organizationUrl must be the org root, e.g. https://dev.azure.com/my-org, not a project/repo URL. Azure DevOps can also return 404 when your signed-in identity cannot see the project.";
  }
  if (status === 401 || status === 403) {
    return "Hint: check auth. For azure-cli auth, run az login and ensure the account/tenant can access this Azure DevOps org.";
  }
  return "";
}

function normalizeRef(value: string): string {
  if (value.startsWith("refs/")) return value;
  return `refs/heads/${value}`;
}

function shortRef(value?: string): string | undefined {
  return value?.replace(/^refs\/heads\//, "");
}

function stringField(value: unknown): string | undefined {
  return typeof value === "string" ? value : undefined;
}

function asIdentityName(value: unknown): string | undefined {
  if (typeof value === "string") return value;
  if (typeof value === "object" && value !== null && "displayName" in value) {
    const displayName = (value as { displayName?: unknown }).displayName;
    return typeof displayName === "string" ? displayName : undefined;
  }
  return undefined;
}

function pullRequestWebUrl(pr: AzurePullRequest, baseUrl: string, project: string): string | undefined {
  if (pr._links?.web?.href) return pr._links.web.href;
  if (pr.repository?.webUrl) return `${pr.repository.webUrl.replace(/\/+$/, "")}/pullrequest/${pr.pullRequestId}`;

  const repository = pr.repository?.name ?? pr.repository?.id;
  if (!repository) return undefined;

  return `${baseUrl}/${encodeURIComponent(project)}/_git/${encodeURIComponent(repository)}/pullrequest/${pr.pullRequestId}`;
}

function reviewFilesUrl(url: string): string {
  try {
    const parsed = new URL(url);
    parsed.searchParams.set("_a", "files");
    return parsed.toString();
  } catch {
    const separator = url.includes("?") ? "&" : "?";
    return `${url}${separator}_a=files`;
  }
}

function workItemWebUrl(baseUrl: string, project: string, id: number): string {
  return `${baseUrl}/${encodeURIComponent(project)}/_workitems/edit/${id}`;
}
