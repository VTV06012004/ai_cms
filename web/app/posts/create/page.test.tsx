import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient, apiUrl } from '@/lib/api';
import { draftKey } from '@/lib/draft';
import CreatePostPage from './page';

const push = vi.fn();

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push }),
}));

vi.mock('@/lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/api')>();
  return {
    ...actual,
    apiClient: { get: vi.fn(), post: vi.fn(), put: vi.fn() },
  };
});

const key = draftKey('3', 'new');
const storedForm = {
  title: 'Bài viết dở dang',
  slug: 'bai-viet-do-dang',
  content: 'Đoạn đã gõ',
  status: 'draft',
  scheduledAt: '',
  categoryId: 1,
  selectedTags: [],
};

beforeEach(() => {
  localStorage.clear();
  localStorage.setItem('userId', '3');
  localStorage.setItem('token', 'abc');
  push.mockReset();
  vi.mocked(apiClient.get).mockReset().mockImplementation(async (url: string) => {
    if (url === apiUrl('categories')) return { data: { data: [{ id: 1, name: 'Tin tức' }] } };
    if (url === apiUrl('tags')) return { data: { data: [] } };
    throw new Error(`unexpected GET ${url}`);
  });
  vi.mocked(apiClient.post).mockReset().mockResolvedValue({ data: {} });
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

describe('create post draft recovery (M6)', () => {
  it('offers an unsaved draft and restores it', async () => {
    localStorage.setItem(key, JSON.stringify({ savedAt: Date.now(), form: storedForm }));
    render(<CreatePostPage />);

    expect(await screen.findByText(/Có bản nháp chưa lưu/)).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Khôi phục bản nháp' }));

    expect(screen.getByDisplayValue('Bài viết dở dang')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Đoạn đã gõ')).toBeInTheDocument();
  });

  it('shows no banner when there is no draft', async () => {
    render(<CreatePostPage />);
    await waitFor(() => expect(apiClient.get).toHaveBeenCalledWith(apiUrl('tags')));

    expect(screen.queryByText(/Có bản nháp chưa lưu/)).not.toBeInTheDocument();
  });

  it('removes the draft once the post is created', async () => {
    localStorage.setItem(key, JSON.stringify({ savedAt: Date.now(), form: storedForm }));
    render(<CreatePostPage />);
    await userEvent.click(await screen.findByRole('button', { name: 'Khôi phục bản nháp' }));

    await userEvent.click(screen.getByRole('button', { name: /Lưu|Đăng|Tạo/ }));

    await waitFor(() => expect(push).toHaveBeenCalledWith('/posts'));
    expect(localStorage.getItem(key)).toBeNull();
  });
});
