'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import axios from 'axios';
import { apiUrl, mediaUrl } from '@/lib/api';
import { removeMarkdownImage } from '@/lib/markdown';

interface Category {
  id: number;
  name: string;
}

interface Tag {
  id: number;
  name: string;
}

export default function CreatePostPage() {
  const [title, setTitle] = useState('');
  const [slug, setSlug] = useState('');
  const [content, setContent] = useState('');
  const [categories, setCategories] = useState<Category[]>([]);
  const [categoryId, setCategoryId] = useState<number | ''>('');
  
  // State cho Tags
  const [tags, setTags] = useState<Tag[]>([]);
  const [selectedTags, setSelectedTags] = useState<number[]>([]);

  // State quản lý Trạng thái bài viết và Phân quyền vai trò
  const [status, setStatus] = useState('draft');
  const [userRole, setUserRole] = useState('author');

  // State cho Ảnh xem trước (Preview Image)
  const [previewImage, setPreviewImage] = useState<string>('');

  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  // Gọi API lấy danh mục, tags và kiểm tra vai trò user khi load trang
  useEffect(() => {
    const role = localStorage.getItem('userRole') || 'author';
    setUserRole(role);

    const fetchData = async () => {
      try {
        const token = localStorage.getItem('token');
        // 1. Lấy Categories
        const catRes = await axios.get(apiUrl('categories'), {
          headers: { Authorization: `Bearer ${token}` },
        });
        const categoryData = Array.isArray(catRes.data) ? catRes.data : (catRes.data.data || []);
        setCategories(categoryData);
        if (categoryData.length > 0) {
          setCategoryId(categoryData[0].id);
        }

        // 2. Lấy Tags
        const tagRes = await axios.get(apiUrl('tags'), {
          headers: { Authorization: `Bearer ${token}` },
        });
        const tagData = Array.isArray(tagRes.data) ? tagRes.data : (tagRes.data.data || []);
        setTags(tagData);

      } catch (err) {
        console.error('Không thể tải dữ liệu danh mục hoặc thẻ:', err);
      }
    };

    fetchData();
  }, []);

  // Tự động tạo slug đơn giản từ tiêu đề
  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setTitle(val);
    const generatedSlug = val
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[^\w\s-]/g, '')
      .replace(/\s+/g, '-');
    setSlug(generatedSlug);
  };

  // Xử lý upload hình ảnh và hiển thị preview
  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);

    try {
      const token = localStorage.getItem('token');
      const res = await axios.post(apiUrl('upload'), formData, {
        headers: { 
          Authorization: `Bearer ${token}` 
        }
      });

      const imageUrl = mediaUrl(res.data.url);
      setPreviewImage(imageUrl); 
      setContent((prev) => prev + `\n![Image](${imageUrl})\n`);
      alert('Upload ảnh thành công và đã chèn vào nội dung!');
    } catch (err) {
      console.error('Lỗi upload ảnh:', err);
      const message = axios.isAxiosError(err) ? err.response?.data?.error : undefined;
      alert(message || 'Upload ảnh thất bại!');
    }
  };

  const handleRemoveImage = () => {
    if (previewImage) {
      setContent((prev) => removeMarkdownImage(prev, previewImage));
    }
    setPreviewImage('');
    
    const fileInput = document.getElementById('image-upload-input') as HTMLInputElement;
    if (fileInput) fileInput.value = '';
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    
    if (categoryId === '') {
      setError('Vui lòng chọn danh mục bài viết.');
      return;
    }

    setLoading(true);

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        router.push('/login');
        return;
      }

      await axios.post(
        apiUrl('posts'),
        {
          title,
          slug,
          content,
          category_id: Number(categoryId),
          tag_ids: selectedTags,
          status: status, // 👈 Gửi trạng thái thực tế người dùng chọn
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      router.push('/posts');
    } catch (err: any) {
      console.error(err);
      setError(err.response?.data?.error || err.response?.data?.message || 'Có lỗi xảy ra khi tạo bài viết.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto mt-10 p-6 bg-white rounded-lg shadow-md">
      <h1 className="text-2xl font-bold mb-6 text-gray-800">Tạo Bài Viết Mới</h1>

      {error && (
        <div className="mb-4 p-3 bg-red-100 text-red-700 rounded-md text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Tiêu đề bài viết
          </label>
          <input
            type="text"
            value={title}
            onChange={handleTitleChange}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-black"
            placeholder="Nhập tiêu đề..."
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Slug (Đường dẫn tĩnh)
          </label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-black"
            placeholder="vd: bai-viet-dau-tien"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Danh mục bài viết
          </label>
          <select
            value={categoryId}
            onChange={(e) => setCategoryId(Number(e.target.value))}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-black bg-white"
          >
            <option value="">-- Chọn danh mục --</option>
            {categories.map((cat) => (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            ))}
          </select>
        </div>

        {/* 🔄 Phần chọn Trạng thái (Workflow Status & Phân quyền) */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Trạng thái bài viết
          </label>
          <select 
            value={status} 
            onChange={(e) => setStatus(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-black bg-white"
          >
            <option value="draft">Bản nháp (Draft)</option>
            <option value="review">Gửi kiểm duyệt (Review)</option>
            {/* Chỉ Admin hoặc Editor mới thấy option Published */}
            {(userRole === 'admin' || userRole === 'editor') && (
              <option value="published">Xuất bản (Published)</option>
            )}
          </select>
        </div>

        {/* 🏷 Phần chọn Thẻ (Tags) */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Thẻ (Tags)
          </label>
          <div className="flex flex-wrap gap-2 p-3 border border-gray-300 rounded-md bg-white">
            {tags.length === 0 ? (
              <span className="text-sm text-gray-400">Chưa có thẻ nào được tạo.</span>
            ) : (
              tags.map((tag) => (
                <label key={tag.id} className="flex items-center space-x-1 text-sm bg-gray-100 px-3 py-1 rounded cursor-pointer hover:bg-gray-200">
                  <input
                    type="checkbox"
                    value={tag.id}
                    checked={selectedTags.includes(tag.id)}
                    onChange={(e) => {
                      const tagId = Number(e.target.value);
                      if (e.target.checked) {
                        setSelectedTags([...selectedTags, tagId]);
                      } else {
                        setSelectedTags(selectedTags.filter((id) => id !== tagId));
                      }
                    }}
                  />
                  <span className="text-black">{tag.name}</span>
                </label>
              ))
            )}
          </div>
        </div>

        {/* 📷 Phần Upload Hình ảnh & Xem trước*/}
        <div>
          <label className="block text-sm font-medium mb-1 text-gray-700">Tải lên hình ảnh</label>
          <input
            id="image-upload-input"                                        
            type="file"
            accept="image/*"
            onChange={handleImageUpload}
            className="w-full border p-2 rounded bg-white text-sm text-black"
          />
          {previewImage && (
            <div className="mt-3 relative inline-block">
              <p className="text-xs text-gray-500 mb-1">Ảnh đã tải lên gần nhất:</p>
              <div className="relative inline-block group">
                <img src={previewImage} alt="Preview" className="h-40 object-cover rounded-md border shadow-sm" />
                <button
                  type="button"
                  onClick={handleRemoveImage}
                  className="absolute top-2 right-2 bg-red-600 text-white rounded-full w-7 h-7 flex items-center justify-center font-bold shadow-md hover:bg-red-700 cursor-pointer transition"
                  title="Xóa ảnh này"
                >
                  ✕
                </button>
              </div>
            </div>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Nội dung
          </label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            required
            rows={6}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-black"
            placeholder="Nhập nội dung bài viết..."
          />
        </div>

        <div className="flex justify-end space-x-3">
          <button
            type="button"
            onClick={() => router.push('/posts')}
            className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
          >
            Hủy
          </button>
          <button
            type="submit"
            disabled={loading}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? 'Đang tạo...' : 'Đăng bài'}
          </button>
        </div>
      </form>
    </div>
  );
}