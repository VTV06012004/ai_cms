import { useCallback, useEffect, useRef, useState } from "react";
import {
  clearDraft,
  loadDraft,
  sameDraft,
  saveDraft,
  type PostDraft,
  type StoredDraft,
} from "./draft";

export const AUTOSAVE_DELAY_MS = 1000;

type Options = {
  // Storage key from draftKey(); null until the user and post are known.
  draftKey: string | null;
  // The current form values.
  form: PostDraft;
  // True once the form shows its starting values (the post has loaded). The
  // values at that moment are the baseline that counts as "no unsaved change".
  ready: boolean;
  // Puts a restored draft back into the form.
  apply: (form: PostDraft) => void;
};

// Keeps unsaved edits in localStorage and offers them back after a reload or a
// forced sign-in. Call markSaved() after the server accepted the form.
export function useDraftAutosave({ draftKey, form, ready, apply }: Options) {
  const [pending, setPending] = useState<StoredDraft | null>(null);
  const baseline = useRef<PostDraft | null>(null);
  const checked = useRef(false);
  const saved = useRef(false);

  // On the first ready render, remember the starting values and look for a
  // draft that differs from them.
  useEffect(() => {
    if (!ready || draftKey === null || checked.current) return;
    checked.current = true;
    baseline.current = form;
    const stored = loadDraft(draftKey);
    if (stored && !sameDraft(stored.form, form)) {
      // One-time read of an external store (localStorage) after mount.
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setPending(stored);
    } else {
      clearDraft(draftKey);
    }
  }, [ready, draftKey, form]);

  // Save a changed form shortly after the last keystroke. While an offered
  // draft waits for an answer nothing is written, so it is not overwritten.
  useEffect(() => {
    if (!ready || draftKey === null || !checked.current || pending || saved.current) return;
    const timer = setTimeout(() => {
      if (saved.current || baseline.current === null) return;
      if (sameDraft(form, baseline.current)) {
        clearDraft(draftKey);
      } else {
        saveDraft(draftKey, form);
      }
    }, AUTOSAVE_DELAY_MS);
    return () => clearTimeout(timer);
  }, [ready, draftKey, form, pending]);

  const restore = useCallback(() => {
    if (pending) apply(pending.form);
    setPending(null);
  }, [pending, apply]);

  const discard = useCallback(() => {
    if (draftKey !== null) clearDraft(draftKey);
    setPending(null);
  }, [draftKey]);

  const markSaved = useCallback(() => {
    saved.current = true;
    if (draftKey !== null) clearDraft(draftKey);
  }, [draftKey]);

  return { pending, restore, discard, markSaved };
}
