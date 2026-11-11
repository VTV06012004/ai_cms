import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  clearAllDrafts,
  clearDraft,
  draftKey,
  loadDraft,
  sameDraft,
  saveDraft,
  type PostDraft,
} from './draft';

const form: PostDraft = {
  title: 'Tiêu đề',
  slug: 'tieu-de',
  content: 'Nội dung dài',
  status: 'draft',
  scheduledAt: '',
  categoryId: 1,
  selectedTags: [2, 3],
};

beforeEach(() => {
  localStorage.clear();
});

describe('draftKey', () => {
  it('separates users and posts', () => {
    expect(draftKey('1', '7')).not.toBe(draftKey('2', '7'));
    expect(draftKey('1', '7')).not.toBe(draftKey('1', '8'));
  });

  it('falls back to an anonymous user', () => {
    expect(draftKey(null, 'new')).toBe(draftKey('', 'new'));
  });
});

describe('sameDraft', () => {
  it('is true for equal forms and false when any field differs', () => {
    expect(sameDraft(form, { ...form, selectedTags: [2, 3] })).toBe(true);
    expect(sameDraft(form, { ...form, title: 'Khác' })).toBe(false);
    expect(sameDraft(form, { ...form, content: 'Khác' })).toBe(false);
    expect(sameDraft(form, { ...form, categoryId: '' })).toBe(false);
    expect(sameDraft(form, { ...form, selectedTags: [3, 2] })).toBe(false);
    expect(sameDraft(form, { ...form, scheduledAt: '2030-01-01T00:00' })).toBe(false);
  });
});

describe('saveDraft and loadDraft', () => {
  it('round-trips a draft with its save time', () => {
    saveDraft('k', form, 1000);

    expect(loadDraft('k', 2000)).toEqual({ savedAt: 1000, form });
  });

  it('returns null when nothing is stored', () => {
    expect(loadDraft('missing')).toBeNull();
  });

  it('ignores and removes data that is not valid JSON', () => {
    localStorage.setItem('k', '{not json');

    expect(loadDraft('k')).toBeNull();
  });

  it('ignores and removes data with the wrong shape', () => {
    localStorage.setItem('k', JSON.stringify({ savedAt: 1, form: { title: 5 } }));

    expect(loadDraft('k')).toBeNull();
    expect(localStorage.getItem('k')).toBeNull();
  });

  it('drops drafts older than 30 days', () => {
    const day = 24 * 60 * 60 * 1000;
    saveDraft('k', form, 0);

    expect(loadDraft('k', 31 * day)).toBeNull();
    expect(localStorage.getItem('k')).toBeNull();
  });

  it('keeps drafts that are 29 days old', () => {
    const day = 24 * 60 * 60 * 1000;
    saveDraft('k', form, 0);

    expect(loadDraft('k', 29 * day)).not.toBeNull();
  });
});

describe('clearDraft and clearAllDrafts', () => {
  it('clearDraft removes one draft only', () => {
    saveDraft('a', form);
    saveDraft('b', form);

    clearDraft('a');

    expect(loadDraft('a')).toBeNull();
    expect(loadDraft('b')).not.toBeNull();
  });

  it('clearAllDrafts removes drafts but keeps other keys', () => {
    saveDraft(draftKey('1', '7'), form);
    saveDraft(draftKey('2', 'new'), form);
    localStorage.setItem('token', 'abc');

    clearAllDrafts();

    expect(loadDraft(draftKey('1', '7'))).toBeNull();
    expect(loadDraft(draftKey('2', 'new'))).toBeNull();
    expect(localStorage.getItem('token')).toBe('abc');
  });
});

describe('when storage is unavailable', () => {
  it('never throws', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('quota exceeded');
    });
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('denied');
    });
    vi.spyOn(Storage.prototype, 'removeItem').mockImplementation(() => {
      throw new Error('denied');
    });

    expect(() => saveDraft('k', form)).not.toThrow();
    expect(loadDraft('k')).toBeNull();
    expect(() => clearDraft('k')).not.toThrow();
  });
});
