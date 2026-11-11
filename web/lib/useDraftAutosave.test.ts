import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { clearDraft, loadDraft, saveDraft, type PostDraft } from './draft';
import { AUTOSAVE_DELAY_MS, useDraftAutosave } from './useDraftAutosave';

const KEY = 'cms.draft.post.1.7';
const server: PostDraft = {
  title: 'Bài trên server',
  slug: 'bai',
  content: 'Nội dung',
  status: 'draft',
  scheduledAt: '',
  categoryId: 1,
  selectedTags: [],
};

type Props = { form: PostDraft; ready: boolean; draftKey: string | null };

function setup(initial: Partial<Props> = {}) {
  const apply = vi.fn();
  const view = renderHook(
    (props: Props) => useDraftAutosave({ ...props, apply }),
    { initialProps: { form: server, ready: true, draftKey: KEY, ...initial } },
  );
  return { ...view, apply };
}

beforeEach(() => {
  localStorage.clear();
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('useDraftAutosave', () => {
  it('saves a changed form after the delay, not before', () => {
    const { rerender } = setup();
    const edited = { ...server, content: 'Đã gõ thêm' };

    rerender({ form: edited, ready: true, draftKey: KEY });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS - 1));
    expect(loadDraft(KEY)).toBeNull();

    act(() => vi.advanceTimersByTime(1));
    expect(loadDraft(KEY)?.form).toEqual(edited);
  });

  it('waits for the user to stop typing', () => {
    const { rerender } = setup();

    rerender({ form: { ...server, content: 'a' }, ready: true, draftKey: KEY });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS - 100));
    rerender({ form: { ...server, content: 'ab' }, ready: true, draftKey: KEY });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS - 100));
    expect(loadDraft(KEY)).toBeNull();

    act(() => vi.advanceTimersByTime(100));
    expect(loadDraft(KEY)?.form.content).toBe('ab');
  });

  it('writes nothing for a form that matches the loaded post', () => {
    setup();

    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS * 3));

    expect(loadDraft(KEY)).toBeNull();
  });

  it('removes the draft when the user undoes every change', () => {
    const { rerender } = setup();
    rerender({ form: { ...server, content: 'x' }, ready: true, draftKey: KEY });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS));
    expect(loadDraft(KEY)).not.toBeNull();

    rerender({ form: server, ready: true, draftKey: KEY });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS));

    expect(loadDraft(KEY)).toBeNull();
  });

  it('does nothing before the post has loaded', () => {
    const { rerender } = setup({ ready: false });

    rerender({ form: { ...server, content: 'x' }, ready: false, draftKey: KEY });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS * 2));

    expect(loadDraft(KEY)).toBeNull();
  });

  it('does nothing until the storage key is known', () => {
    const { rerender } = setup({ draftKey: null });

    rerender({ form: { ...server, content: 'x' }, ready: true, draftKey: null });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS * 2));

    expect(localStorage.length).toBe(0);
  });

  it('offers a stored draft that differs from the loaded post', () => {
    const stored = { ...server, content: 'Bài viết dở dang' };
    saveDraft(KEY, stored, Date.now());

    const { result } = setup();

    expect(result.current.pending?.form).toEqual(stored);
  });

  it('does not offer a stored draft identical to the loaded post and removes it', () => {
    saveDraft(KEY, server, Date.now());

    const { result } = setup();

    expect(result.current.pending).toBeNull();
    expect(loadDraft(KEY)).toBeNull();
  });

  it('keeps an offered draft untouched while it waits for an answer', () => {
    const stored = { ...server, content: 'Bài viết dở dang' };
    saveDraft(KEY, stored, Date.now());

    setup();
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS * 3));

    expect(loadDraft(KEY)?.form).toEqual(stored);
  });

  it('restore applies the draft and clears the offer', () => {
    const stored = { ...server, content: 'Bài viết dở dang' };
    saveDraft(KEY, stored, Date.now());
    const { result, apply } = setup();

    act(() => result.current.restore());

    expect(apply).toHaveBeenCalledWith(stored);
    expect(result.current.pending).toBeNull();
  });

  it('discard removes the stored draft and clears the offer', () => {
    saveDraft(KEY, { ...server, content: 'Bài viết dở dang' }, Date.now());
    const { result, apply } = setup();

    act(() => result.current.discard());

    expect(apply).not.toHaveBeenCalled();
    expect(result.current.pending).toBeNull();
    expect(loadDraft(KEY)).toBeNull();
  });

  it('markSaved clears the draft and stops further saving', () => {
    const { result, rerender } = setup();
    rerender({ form: { ...server, content: 'x' }, ready: true, draftKey: KEY });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS));
    expect(loadDraft(KEY)).not.toBeNull();

    act(() => result.current.markSaved());
    rerender({ form: { ...server, content: 'xy' }, ready: true, draftKey: KEY });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS * 2));

    expect(loadDraft(KEY)).toBeNull();
  });

  it('keeps the draft when the save never succeeded', () => {
    const { rerender } = setup();
    const edited = { ...server, content: 'Rất nhiều chữ' };

    rerender({ form: edited, ready: true, draftKey: KEY });
    act(() => vi.advanceTimersByTime(AUTOSAVE_DELAY_MS));
    clearDraft('some-other-key');

    expect(loadDraft(KEY)?.form).toEqual(edited);
  });
});
