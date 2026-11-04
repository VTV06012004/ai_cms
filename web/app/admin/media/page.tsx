'use client';

import { useCallback, useEffect, useState } from 'react';
import Link from 'next/link';
import axios from 'axios';
import { apiUrl } from '@/lib/api';
import type { MediaItem } from '@/app/components/MediaPicker';

export default function MediaLibraryPage() {
  const [items, setItems] = useState<MediaItem[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');
  const [userRole, setUserRole] = useState('author');

  const canDelete = userRole === 'admin' || userRole === 'editor';

  const load = useCallback(async (nextPage: number) => {
    setLoading(true);
    setError('');
    try {
      const res = await axios.get(apiUrl('media'), { params: { page: nextPage, per_page: 24 } });
      setItems((prev: MediaItem[]) => (nextPage === 1 ? res.data.data : [...prev, ...res.data.data]));
      setTotalPages(res.data.pagination?.total_pages ?? 0);
      setTotal(res.data.pagination?.total ?? 0);
      setPage(nextPage);
    } catch {
      setError('Không thể tải thư viện ảnh');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    setUserRole((localStorage.getItem('userRole') || 'author').toLowerCase());
    load(1);
  }, [load]);

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const formData = new FormData();
    formData.append('file', file);
    setUploading(true);
    try {
      await axios.post(apiUrl('upload'), formData);
      await load(1);
    } catch (err) {
      const message = axios.isAxiosError(err) ? err.response?.data?.error : undefined;
      alert(message || 'Upload ảnh thất bại!');
    } finally {
      setUploading(false);
      e.target.value = '';
    }
  };

  const handleDelete = async (item: MediaItem) => {
    if (!confirm(`Xóa ảnh ${item.name}? Các bài viết đang dùng ảnh này sẽ bị mất ảnh.`)) return;
    try {
      await axios.delete(apiUrl(`media/${encodeURIComponent(item.name)}`));
      setItems((prev: MediaItem[]) => prev.filter((i) => i.name !== item.name));
      setTotal((n) => Math.max(0, n - 1));
    } catch (err) {
      const message = axios.isAxiosError(err) ? err.response?.data?.error : undefined;
      alert(message || 'Xóa ảnh thất bại!');
    }
  };

  const copyUrl = async (url: string) => {
    try {
      await navigator.clipboard.writeText(url);
      alert('Đã sao chép liên kết ảnh');
    } catch {
      prompt('Sao chép liên kết:', url);
    }
  };

  return (
    <div className="max-w-5xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6 flex-wrap gap-3">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">🖼️ Thư viện ảnh</h1>
          <p className="text-sm text-gray-500">{total} ảnh</p>
        </div>
        <div className="flex items-center gap-3">
          <label className="px-4 py-2 bg-blue-600 text-white text-sm font-semibold rounded cursor-pointer hover:bg-blue-700">
            {uploading ? 'Đang tải lên...' : '＋ Tải ảnh lên'}
            <input type="file" accept="image/*" onChange={handleUpload} disabled={uploading} className="hidden" />
          </label>
          <Link href="/posts" className="text-sm text-blue-600 hover:underline">← Về bài viết</Link>
        </div>
      </div>

      {error && <p className="text-sm text-red-600 mb-4">{error}</p>}
      {!loading && items.length === 0 && !error && (
        <p className="text-sm text-gray-500 italic bg-gray-50 p-4 rounded border">Chưa có ảnh nào.</p>
      )}

      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
        {items.map((item) => (
          <div key={item.name} className="border rounded-lg overflow-hidden bg-white shadow-sm">
            <img src={item.url} alt={item.name} className="h-36 w-full object-cover bg-gray-100" />
            <div className="p-2 space-y-1">
              <p className="text-xs text-gray-700 truncate" title={item.name}>{item.name}</p>
              <p className="text-[11px] text-gray-400">
                {(item.size / 1024).toFixed(0)} KB • {new Date(item.uploaded_at).toLocaleDateString('vi-VN')}
              </p>
              <div className="flex gap-2 pt-1">
                <button
                  type="button"
                  onClick={() => copyUrl(item.url)}
                  className="flex-1 text-xs px-2 py-1 bg-blue-50 text-blue-600 border border-blue-200 rounded hover:bg-blue-100 cursor-pointer"
                >
                  Sao chép link
                </button>
                {canDelete && (
                  <button
                    type="button"
                    onClick={() => handleDelete(item)}
                    className="text-xs px-2 py-1 bg-red-50 text-red-600 border border-red-200 rounded hover:bg-red-100 cursor-pointer"
                  >
                    Xóa
                  </button>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {page < totalPages && (
        <div className="text-center mt-6">
          <button
            type="button"
            disabled={loading}
            onClick={() => load(page + 1)}
            className="px-5 py-2 text-sm bg-gray-100 border rounded hover:bg-gray-200 cursor-pointer"
          >
            {loading ? 'Đang tải...' : 'Tải thêm'}
          </button>
        </div>
      )}
    </div>
  );
}