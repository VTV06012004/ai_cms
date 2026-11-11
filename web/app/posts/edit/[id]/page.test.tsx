import { render, screen, waitFor } from '@testing-library/react';
import { AxiosError } from 'axios';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient, apiUrl } from '@/lib/api';
import { draftKey } from '@/lib/draft';
import EditPostPage from './page';

const push = vi.fn();

function apiError(message: string) {
  return new AxiosError('Request failed', 'ERR_BAD_REQUEST', undefined, undefined, {
    status: 400,
    data: { error: message },
  } as never);
}

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push }),
  useParams: () => ({ id: '7' }),
}));

vi.mock('@/lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/lib/api')>();
  return {
    ...actual,
    apiClient: { get: vi.fn(), post: vi.fn(), put: vi.fn() },
  };
});

type PostFixture = {
  title: string;
  slug: string;
  content: string;
  status: string;
  scheduled_at?: string | null;
  category_id: number;
  tags: { id: number }[];
};

const basePost: PostFixture = {
  title: 'Bài cũ',
  slug: 'bai-cu',
  content: 'Nội dung',
  status: 'draft',
  category_id: 1,
  tags: [],
};

function loadPage(post: PostFixture) {
  vi.mocked(apiClient.get).mockImplementation(async (url: string) => {
    if (url === apiUrl('categories')) return { data: { data: [{ id: 1, name: 'Tin tức' }] } };
    if (url === apiUrl('tags')) return { data: { data: [] } };
    if (url === apiUrl('posts/7')) return { data: { data: post } };
    if (url === apiUrl('posts-revisions/7')) return { data: { data: [] } };
    throw new Error(`unexpected GET ${url}`);
  });
}

async function openPage(role: string, post: PostFixture) {
  localStorage.setItem('userRole', role);
  loadPage(post);
  render(<EditPostPage />);
  await screen.findByDisplayValue(post.title);
}

beforeEach(() => {
  localStorage.clear();
  push.mockReset();
  vi.mocked(apiClient.get).mockReset();
  vi.mocked(apiClient.put).mockReset().mockResolvedValue({ data: {} });
  vi.spyOn(window, 'alert').mockImplementation(() => {});
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

describe('edit post form', () => {
  it('lets an author save a published post without changing its status (H1)', async () => {
    await openPage('author', { ...basePost, status: 'published' });

    expect(screen.getByDisplayValue('Đã đăng (Published)')).toBeInTheDocument();
    expect(
      screen.getByText(/Sửa bài đã đăng sẽ chuyển bài về trạng thái chờ duyệt/),
    ).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Lưu thay đổi' }));

    await waitFor(() => expect(apiClient.put).toHaveBeenCalledTimes(1));
    expect(apiClient.put).toHaveBeenCalledWith(apiUrl('posts/7'), {
      title: 'Bài cũ',
      slug: 'bai-cu',
      content: 'Nội dung',
      status: 'published',
      scheduled_at: null,
      category_id: 1,
      tag_ids: [],
    });
    await waitFor(() => expect(push).toHaveBeenCalledWith('/posts'));
  });

  it('keeps the schedule of a scheduled post when it is saved again', async () => {
    await openPage('editor', {
      ...basePost,
      status: 'scheduled',
      scheduled_at: '2030-01-02T03:04:00.000Z',
    });

    await userEvent.click(screen.getByRole('button', { name: 'Lưu thay đổi' }));

    await waitFor(() => expect(apiClient.put).toHaveBeenCalledTimes(1));
    expect(apiClient.put).toHaveBeenCalledWith(
      apiUrl('posts/7'),
      expect.objectContaining({
        status: 'scheduled',
        scheduled_at: '2030-01-02T03:04:00.000Z',
      }),
    );
  });

  it('does not show the published warning or option to an author editing a draft', async () => {
    await openPage('author', basePost);

    expect(screen.queryByText(/Sửa bài đã đăng/)).not.toBeInTheDocument();
    expect(screen.queryByRole('option', { name: /Published/ })).not.toBeInTheDocument();
  });

  it('offers publish and schedule to an editor', async () => {
    await openPage('editor', basePost);

    expect(screen.getByRole('option', { name: 'Xuất bản (Published)' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'Hẹn giờ đăng (Scheduled)' })).toBeInTheDocument();
  });

  it('shows the API error and stays on the page when saving fails', async () => {
    await openPage('author', basePost);
    vi.mocked(apiClient.put).mockRejectedValue(
      apiError('vai trò hiện tại không có quyền xuất bản bài viết'),
    );

    await userEvent.click(screen.getByRole('button', { name: 'Lưu thay đổi' }));

    await waitFor(() =>
      expect(window.alert).toHaveBeenCalledWith('vai trò hiện tại không có quyền xuất bản bài viết'),
    );
    expect(push).not.toHaveBeenCalled();
  });

  it('never sets the Authorization header itself', async () => {
    await openPage('author', basePost);
    await userEvent.click(screen.getByRole('button', { name: 'Lưu thay đổi' }));
    await waitFor(() => expect(apiClient.put).toHaveBeenCalled());

    const calls = [
      ...vi.mocked(apiClient.get).mock.calls,
      ...vi.mocked(apiClient.put).mock.calls,
    ];
    for (const call of calls) {
      expect(JSON.stringify(call.slice(1))).not.toMatch(/Authorization/i);
    }
  });
});

describe('edit post draft recovery (M6)', () => {
  const key = draftKey('3', '7');
  const storedForm = {
    title: 'Tiêu đề đang gõ dở',
    slug: 'bai-cu',
    content: 'Nội dung',
    status: 'draft',
    scheduledAt: '',
    categoryId: 1,
    selectedTags: [],
  };

  beforeEach(() => {
    localStorage.setItem('userId', '3');
  });

  function storeDraft() {
    localStorage.setItem(key, JSON.stringify({ savedAt: Date.now(), form: storedForm }));
  }

  it('offers an unsaved draft and restores it on request', async () => {
    storeDraft();
    await openPage('author', basePost);

    expect(await screen.findByText(/Có bản nháp chưa lưu/)).toBeInTheDocument();
    expect(screen.getByDisplayValue('Bài cũ')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Khôi phục bản nháp' }));

    expect(screen.getByDisplayValue('Tiêu đề đang gõ dở')).toBeInTheDocument();
    expect(screen.queryByText(/Có bản nháp chưa lưu/)).not.toBeInTheDocument();
  });

  it('drops the draft when the author discards it', async () => {
    storeDraft();
    await openPage('author', basePost);

    await userEvent.click(await screen.findByRole('button', { name: 'Bỏ bản nháp' }));

    expect(screen.getByDisplayValue('Bài cũ')).toBeInTheDocument();
    expect(localStorage.getItem(key)).toBeNull();
  });

  it('removes the draft after the server accepted the save', async () => {
    storeDraft();
    await openPage('author', basePost);
    await userEvent.click(await screen.findByRole('button', { name: 'Khôi phục bản nháp' }));

    await userEvent.click(screen.getByRole('button', { name: 'Lưu thay đổi' }));

    await waitFor(() => expect(push).toHaveBeenCalledWith('/posts'));
    expect(localStorage.getItem(key)).toBeNull();
  });

  it('keeps the draft when saving fails', async () => {
    storeDraft();
    await openPage('author', basePost);
    await userEvent.click(await screen.findByRole('button', { name: 'Khôi phục bản nháp' }));
    vi.mocked(apiClient.put).mockRejectedValue(apiError('lỗi'));

    await userEvent.click(screen.getByRole('button', { name: 'Lưu thay đổi' }));

    await waitFor(() => expect(window.alert).toHaveBeenCalledWith('lỗi'));
    expect(localStorage.getItem(key)).not.toBeNull();
  });
});
