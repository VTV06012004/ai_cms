import { useMemo, useSyncExternalStore } from "react";

export type Session = {
  signedIn: boolean;
  token: string;
  userId: string;
  role: string;
  name: string;
  email: string;
};

const FIELDS = {
  token: "token",
  userId: "userId",
  role: "userRole",
  name: "userName",
  email: "userEmail",
} as const;

const SESSION_CHANGED = "cms:session-changed";

function readItem(key: string): string {
  try {
    return localStorage.getItem(key) ?? "";
  } catch {
    return "";
  }
}

// One string per snapshot so React can compare it by value.
function snapshot(): string {
  return JSON.stringify(Object.values(FIELDS).map(readItem));
}

const SIGNED_OUT: Session = {
  signedIn: false,
  token: "",
  userId: "",
  role: "author",
  name: "",
  email: "",
};

function parse(raw: string): Session {
  if (!raw) return SIGNED_OUT;
  const [token, userId, role, name, email] = JSON.parse(raw) as string[];
  return {
    signedIn: token !== "",
    token,
    userId,
    role: (role || "author").toLowerCase(),
    name,
    email: email || "user@cms.com",
  };
}

export function readSession(): Session {
  return parse(snapshot());
}

export function saveSession(session: Omit<Session, "signedIn">): void {
  try {
    localStorage.setItem(FIELDS.token, session.token);
    localStorage.setItem(FIELDS.role, session.role);
    localStorage.setItem(FIELDS.name, session.name);
    localStorage.setItem(FIELDS.email, session.email);
    localStorage.setItem(FIELDS.userId, session.userId);
  } catch {
    // Without storage the user simply has to sign in again.
  }
  notifySessionChanged();
}

// Call after code outside this module changed the stored session (sign-out,
// token refresh) so mounted components read it again.
export function notifySessionChanged(): void {
  window.dispatchEvent(new Event(SESSION_CHANGED));
}

function subscribe(onChange: () => void): () => void {
  window.addEventListener("storage", onChange);
  window.addEventListener(SESSION_CHANGED, onChange);
  return () => {
    window.removeEventListener("storage", onChange);
    window.removeEventListener(SESSION_CHANGED, onChange);
  };
}

// Signed-out values while nobody is signed in and during server rendering.
export function useSession(): Session {
  const raw = useSyncExternalStore(subscribe, snapshot, () => "");
  return useMemo(() => parse(raw), [raw]);
}
