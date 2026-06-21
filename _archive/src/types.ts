export type SectionType = "pullRequests" | "workItems";

export interface AppConfig {
  organizationUrl: string;
  project: string;
  auth: AuthConfig;
  /** @deprecated use auth.env with auth.type: pat */
  patEnv?: string;
  defaults?: {
    refreshSeconds?: number;
    limit?: number;
  };
  sections: SectionConfig[];
}

export type AuthConfig = PatAuthConfig | AzureCliAuthConfig;

export interface PatAuthConfig {
  type: "pat";
  env?: string;
}

export interface AzureCliAuthConfig {
  type: "azure-cli";
  tenant?: string;
}

export type SectionConfig = PullRequestSectionConfig | WorkItemSectionConfig;

export interface BaseSectionConfig {
  title: string;
  type: SectionType;
  limit?: number;
}

export interface PullRequestSectionConfig extends BaseSectionConfig {
  type: "pullRequests";
  status?: "active" | "abandoned" | "completed" | "all";
  role?: "author" | "reviewer" | "all";
  repositoryId?: string;
  targetBranch?: string;
  sourceBranch?: string;
}

export interface WorkItemSectionConfig extends BaseSectionConfig {
  type: "workItems";
  wiql: string;
  fields?: string[];
}

export interface DashboardItem {
  id: string;
  title: string;
  state?: string;
  subtitle?: string;
  author?: string;
  assignee?: string;
  updatedAt?: string;
  url?: string;
  reviewUrl?: string;
  kind?: "pullRequest" | "workItem";
  pullRequest?: PullRequestRef;
  raw?: unknown;
}

export interface PullRequestRef {
  pullRequestId: number;
  repositoryId?: string;
  repositoryName?: string;
}

export interface PullRequestReview {
  title: string;
  files: PullRequestReviewFile[];
  threads: PullRequestReviewThread[];
  iterationId?: number;
}

export interface PullRequestReviewFile {
  path: string;
  changeType: string;
  changeTrackingId?: number;
}

export interface PullRequestReviewThread {
  id: number;
  status?: string;
  filePath?: string;
  line?: number;
  comments: PullRequestReviewComment[];
}

export interface PullRequestReviewComment {
  author?: string;
  content: string;
  commentType?: string;
}

export interface LoadedSection {
  config: SectionConfig;
  items: DashboardItem[];
  loading: boolean;
  error?: string;
  lastLoadedAt?: Date;
}

export interface Identity {
  id: string;
  displayName: string;
  uniqueName?: string;
}
