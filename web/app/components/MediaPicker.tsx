'use client';

import { useCallback, useEffect, useState } from 'react';
import axios from 'axios';
import { apiUrl, mediaUrl } from '@/lib/api';

export interface MediaItem {
  name: string;
  url: string;
  size: number;
  uploaded_at: string;
}

interface Props {
  onSelect: (url: string) => void;
  onClose: () => void;
}

export default function MediaPicker({ onSelect, onClose }: Props) {
  const [items, setItems] = useState<MediaItem[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = useCallback(async (nextPage: number) => {
    setLoading(true);
    setError('');
    try {
      const res = await axios.get(apiUrl('media'), { params: { page: nextPage, per_page: 24 } });
      setItems((prev: MediaItem[]) => (nextPage === 1 ? res.data.data : [...prev, ...res.data.data]));
      setTotalPages(res.data.pagination?.total_pages ?? 0);
      setPage(nextPage);
    } catch {
      setError('Không thể tải thư viện ảnh');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load(1);
  }, [load]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-3xl max-h-[85vh] flex flex-col">
        <div className="flex items-center justify-between p-4 border-b">
          <h3 className="font-bold text-gray-800">Chọn ảnh từ thư viện</h3>
          <button type="button" onClick={onClose} className="text-gray-500 hover:text-gray-800 cursor-pointer">✕</button>
        </div>
        <div className="p-4 overflow-y-auto">
          {error && <p className="text-sm text-red-600 mb-3">{error}</p>}
          {!loading && items.length === 0 && !error && (
            <p className="text-sm text-gray-500 italic">Thư viện chưa có ảnh nào.</p>
          )}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {items.map((item) => (
              <button
                key={item.name}
                type="button"
                onClick={() => onSelect(mediaUrl(item.url))}
                className="border rounded overflow-hidden hover:ring-2 hover:ring-blue-500 cursor-pointer bg-gray-50"
                title={item.name}
              >
                <img src={mediaUrl(item.url)} alt={item.name} className="h-28 w-full object-cover" />
              </button>
            ))}
          </div>
          {page < totalPages && (
            <div className="text-center mt-4">
              <button
                type="button"
                disabled={loading}
                onClick={() => load(page + 1)}
                className="px-4 py-2 text-sm bg-gray-100 border rounded hover:bg-gray-200 cursor-pointer"
              >
                {loading ? 'Đang tải...' : 'Tải thêm'}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}