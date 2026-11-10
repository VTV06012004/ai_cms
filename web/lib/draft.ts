// Unsaved post drafts kept in the browser so a session that expires (the JWT
// lives 15 minutes) or a closed tab does not lose a long article.

export type PostDraft = {
  title: string;
  slug: string;
  content: string;
  status: string;
  scheduledAt: string;
  categoryId: number | "";
  selectedTags: number[];
};

export type StoredDraft = {
  savedAt: number;
  form: PostDraft;
};

const PREFIX = "cms.draft.";
const MAX_AGE_MS = 30 * 24 * 60 * 60 * 1000;

// Drafts are scoped by user so another account on the same browser never sees them.
export function draftKey(userId: string | null, postId: string): string {
  return `${PREFIX}post.${userId || "anonymous"}.${postId}`;
}

export function sameDraft(a: PostDraft, b: PostDraft): boolean {
  return (
    a.title === b.title &&
    a.slug === b.slug &&
    a.content === b.content &&
    a.status === b.status &&
    a.scheduledAt === b.scheduledAt &&
    a.categoryId === b.categoryId &&
    a.selectedTags.length === b.selectedTags.length &&
    a.selectedTags.every((id, index) => id === b.selectedTags[index])
  );
}

function isPostDraft(value: unknown): value is PostDraft {
  if (typeof value !== "object" || value === null) return false;
  const draft = value as Record<string, unknown>;
  return (
    typeof draft.title === "string" &&
    typeof draft.slug === "string" &&
    typeof draft.content === "string" &&
    typeof draft.status === "string" &&
    typeof draft.scheduledAt === "string" &&
    (typeof draft.categoryId === "number" || draft.categoryId === "") &&
    Array.isArray(draft.selectedTags) &&
    draft.selectedTags.every((id) => typeof id === "number")
  );
}

export function saveDraft(key: string, form: PostDraft, now: number = Date.now()): void {
  try {
    const stored: StoredDraft = { savedAt: now, form };
    localStorage.setItem(key, JSON.stringify(stored));
  } catch {
    // Storage can be full or unavailable (private mode); autosave is best effort.
  }
}

export function loadDraft(key: string, now: number = Date.now()): StoredDraft | null {
  try {
    const raw = localStorage.getItem(key);
    if (raw === null) return null;
    const parsed: unknown = JSON.parse(raw);
    if (
      typeof parsed !== "object" ||
      parsed === null ||
      typeof (parsed as StoredDraft).savedAt !== "number" ||
      !isPostDraft((parsed as StoredDraft).form)
    ) {
      localStorage.removeItem(key);
      return null;
    }
    const stored = parsed as StoredDraft;
    if (now - stored.savedAt > MAX_AGE_MS) {
      localStorage.removeItem(key);
      return null;
    }
    return stored;
  } catch {
    return null;
  }
}

export function clearDraft(key: string): void {
  try {
    localStorage.removeItem(key);
  } catch {
    // Nothing to clear when storage is unavailable.
  }
}

// Called on an explicit logout so drafts do not stay on a shared computer.
// A session that merely expired keeps its drafts on purpose.
export function clearAllDrafts(): void {
  try {
    const keys: string[] = [];
    for (let index = 0; index < localStorage.length; index++) {
      const key = localStorage.key(index);
      if (key !== null && key.startsWith(PREFIX)) keys.push(key);
    }
    keys.forEach((key) => localStorage.removeItem(key));
  } catch {
    // Nothing to clear when storage is unavailable.
  }
}
