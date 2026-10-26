'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter, useParams } from 'next/navigation';
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

interface Revision {
  id: number;
  post_id: number;
  title: string;
  slug: string;
  content: string;
  status: string;
  category_id: number;
  created_at: string;
  user?: {
    id: number;
    email: string;
    role: string;
  };
}

function EditPostForm() {
  const [title, setTitle] = useState('');
  const [slug, setSlug] = useState('');
  const [content, setContent] = useState('');
  const [status, setStatus] = useState('draft');
  const [categoryId, setCategoryId] = useState<number | ''>('');
  
  const [userRole, setUserRole] = useState('author');

  const [categories, setCategories] = useState<Category[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [selectedTags, setSelectedTags] = useState<number[]>([]);
  const [previewImage, setPreviewImage] = useState<string>('');

  const [revisions, setRevisions] = useState<Revision[]>([]);
  const [restoringId, setRestoringId] = useState<number | null>(null);
  const [previewRevision, setPreviewRevision] = useState<Revision | null>(null);

  const [loading, setLoading] = useState(true);
  const router = useRouter();
  const params = useParams();
  const { id } = params;

  const fetchRevisions = async () => {
    try {
      const token = localStorage.getItem('token');
      const res = await axios.get(apiUrl(`posts-revisions/${id}`), {
        headers: { Authorization: `Bearer ${token}` },
      });
      setRevisions(res.data.data || []);
    } catch (err) {
      console.error('Không thể tải lịch sử phiên bản:', err);
    }
  };

  useEffect(() => {
    const role = localStorage.getItem('userRole') || 'author';
    setUserRole(role);

    const fetchData = async () => {
      try {
        const token = localStorage.getItem('token');
        const catRes = await axios.get(apiUrl('categories'), {
          headers: { Authorization: `Bearer ${token}` },
        });
        const catData = Array.isArray(catRes.data) ? catRes.data : (catRes.data.data || []);
        setCategories(catData);

        const tagRes = await axios.get(apiUrl('tags'), {
          headers: { Authorization: `Bearer ${token}` },
        });
        const tagData = Array.isArray(tagRes.data) ? tagRes.data : (tagRes.data.data || []);
        setTags(tagData);

        const postRes = await axios.get(apiUrl(`posts/${id}`));
        const post = postRes.data.data || postRes.data; 

        if (post) {
          setTitle(post.title || '');
          setSlug(post.slug || '');
          const postContent = post.content || ''; 
          setContent(postContent);
          setStatus(post.status || 'draft');
          setCategoryId(post.category_id || post.CategoryID || '');

          const imgMatch = postContent.match(/!\[.*?\]\((.*?)\)/);
          if (imgMatch && imgMatch[1]) {
            setPreviewImage(mediaUrl(imgMatch[1]));
          }

          const postTags = post.tags || post.Tags || [];
          setSelectedTags(postTags.map((t: any) => t.id || t.ID));
        }

        fetchRevisions();
      } catch (err) {
        console.error('Lỗi tải dữ liệu:', err);
        alert('Không thể tải thông tin bài viết!');
      } finally {
        setLoading(false);
      }
    };

    if (id) {
      fetchData();
    }
  }, [id]);

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
    setLoading(true);

    try {
      const token = localStorage.getItem('token');
      await axios.put(
        apiUrl(`posts/${id}`),
        {
          title,
          slug,
          content,
          status,
          category_id: Number(categoryId),
          tag_ids: selectedTags,
        },
        {
          headers: { Authorization: `Bearer ${token}` }
        }
      );

      alert('Cập nhật bài viết thành công!');
      router.push('/posts'); 
    } catch (err: any) {
      console.error('Lỗi cập nhật:', err);
      alert(err.response?.data?.error || 'Cập nhật thất bại!');
    } finally {
      setLoading(false);
    }
  };

  const handleRestore = async (rev: Revision) => {
    if (!confirm(`Bạn có chắc chắn muốn khôi phục về phiên bản từ ngày ${new Date(rev.created_at).toLocaleString('vi-VN')} không?`)) {
      return;
    }

    setRestoringId(rev.id);
    try {
      const token = localStorage.getItem('token');
      await axios.post(
        apiUrl(`revisions/${rev.id}/restore`),
        {},
        {
          headers: { Authorization: `Bearer ${token}` }
        }
      );

      setTitle(rev.title);
      setSlug(rev.slug);
      setContent(rev.content);
      if (rev.category_id) setCategoryId(rev.category_id);

      // Tự động nhận diện lại ảnh nếu bản khôi phục có chứa ảnh
      const imgMatch = rev.content.match(/!\[.*?\]\((.*?)\)/);
      if (imgMatch && imgMatch[1]) {
        setPreviewImage(mediaUrl(imgMatch[1]));
      } else {
        setPreviewImage('');
      }

      alert('Khôi phục phiên bản thành công!');
      fetchRevisions();
    } catch (err: any) {
      console.error('Lỗi khôi phục:', err);
      alert(err.response?.data?.error || 'Khôi phục phiên bản thất bại!');
    } finally {
      setRestoringId(null);
    }
  };

  return (
    <div className="max-w-4xl mx-auto mt-10 p-6 bg-white rounded-lg shadow-md text-black relative">
      <h1 className="text-2xl font-bold mb-6">Chỉnh sửa Bài viết (ID: {id})</h1>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1">Tiêu đề</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
            className="w-full border p-2 rounded"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Slug</label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            required
            className="w-full border p-2 rounded"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Danh mục</label>
          <select
            value={categoryId}
            onChange={(e) => setCategoryId(Number(e.target.value))}
            required
            className="w-full border p-2 rounded bg-white"
          >
            <option value="">-- Chọn danh mục --</option>
            {categories.map((cat) => (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Thẻ (Tags)</label>
          <div className="flex flex-wrap gap-2 p-3 border rounded bg-white">
            {tags.length === 0 ? (
              <span className="text-sm text-gray-400">Chưa có thẻ nào.</span>
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
                        setSelectedTags(selectedTags.filter((tId) => tId !== tagId));
                      }
                    }}
                  />
                  <span>{tag.name}</span>
                </label>
              ))
            )}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Tải lên hình ảnh</label>
          <input
            id="image-upload-input"
            type="file"
            accept="image/*"
            onChange={handleImageUpload}
            className="w-full border p-2 rounded bg-white text-sm"
          />
          {previewImage && (
            <div className="mt-3 relative inline-block">
              <p className="text-xs text-gray-500 mb-1">Ảnh hiện tại:</p>
              <div className="relative inline-block group">
                <img 
                  src={mediaUrl(previewImage)}
                  alt="Preview" 
                  className="h-40 object-cover rounded-md border shadow-sm" 
                />
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
          <label className="block text-sm font-medium mb-1">Trạng thái</label>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="w-full border p-2 rounded bg-white"
          >
            <option value="draft">Bản nháp (Draft)</option>
            <option value="review">Gửi kiểm duyệt (Review)</option>
            {(userRole === 'admin' || userRole === 'editor') && (
              <option value="published">Xuất bản (Published)</option>
            )}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Nội dung</label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            required
            rows={8}
            className="w-full border p-2 rounded font-mono text-sm"
          />
        </div>

        <div className="flex justify-end space-x-3 pt-2">
          <button
            type="button"
            onClick={() => router.push('/posts')}
            className="px-4 py-2 border rounded bg-gray-100 hover:bg-gray-200 cursor-pointer"
          >
            Hủy
          </button>
          <button
            type="submit"
            disabled={loading}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400 cursor-pointer font-medium"
          >
            {loading ? 'Đang lưu...' : 'Lưu thay đổi'}
          </button>
        </div>
      </form>

      {/* LỊCH SỬ PHIÊN BẢN */}
      <div className="mt-12 pt-8 border-t border-gray-200">
        <h2 className="text-xl font-bold mb-4 text-gray-800 flex items-center gap-2">
          <span>📜 Lịch sử chỉnh sửa bài viết</span>
          <span className="text-xs font-normal text-gray-500 bg-gray-100 px-2 py-1 rounded">
            {revisions.length} phiên bản
          </span>
        </h2>

        {revisions.length === 0 ? (
          <p className="text-sm text-gray-500 italic bg-gray-50 p-4 rounded-md border">
            Chưa có lịch sử chỉnh sửa nào.
          </p>
        ) : (
          <div className="space-y-3">
            {revisions.map((rev) => (
              <div 
                key={rev.id} 
                className="p-4 border border-gray-200 rounded-lg bg-gray-50 hover:bg-white hover:shadow-sm transition flex flex-col md:flex-row md:items-center justify-between gap-4"
              >
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-gray-800">{rev.title}</span>
                    <span className={`text-xs px-2 py-0.5 rounded font-medium ${
                      rev.status === 'published' ? 'bg-green-100 text-green-700' : 'bg-gray-200 text-gray-700'
                    }`}>
                      {rev.status}
                    </span>
                  </div>
                  <p className="text-xs text-gray-500">
                    Sửa lúc: <span className="font-medium text-gray-700">{new Date(rev.created_at).toLocaleString('vi-VN')}</span>
                    {rev.user && <span> • bởi: <strong className="text-blue-600">{rev.user.email}</strong></span>}
                  </p>
                </div>

                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => setPreviewRevision(rev)}
                    className="px-3 py-1.5 bg-blue-50 text-blue-600 border border-blue-200 text-xs font-semibold rounded hover:bg-blue-100 cursor-pointer"
                  >
                    👁️ Xem thử
                  </button>
                  <button
                    type="button"
                    disabled={restoringId === rev.id}
                    onClick={() => handleRestore(rev)}
                    className="px-3 py-1.5 bg-amber-500 text-white text-xs font-semibold rounded hover:bg-amber-600 cursor-pointer"
                  >
                    {restoringId === rev.id ? 'Đang khôi phục...' : '↩️ Khôi phục'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* MODAL XEM TRƯỚC PHIÊN BẢN CŨ & RENDER ẢNH */}
      {previewRevision && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-2xl max-w-2xl w-full max-h-[85vh] flex flex-col overflow-hidden border text-black">
            <div className="p-4 border-b bg-gray-50 flex justify-between items-center">
              <div>
                <h3 className="text-lg font-bold text-gray-800">👁️ Xem trước phiên bản lịch sử</h3>
                <p className="text-xs text-gray-500">
                  Lưu lúc: {new Date(previewRevision.created_at).toLocaleString('vi-VN')}
                </p>
              </div>
              <button
                onClick={() => setPreviewRevision(null)}
                className="text-gray-400 hover:text-gray-600 text-xl font-bold p-1"
              >
                ✕
              </button>
            </div>

            <div className="p-6 overflow-y-auto space-y-4 text-sm text-gray-800">
              <div>
                <label className="text-xs font-bold text-gray-500 uppercase tracking-wider block mb-1">Tiêu đề</label>
                <div className="p-2.5 bg-gray-50 border rounded-md font-medium text-base">
                  {previewRevision.title}
                </div>
              </div>

              <div>
                <label className="text-xs font-bold text-gray-500 uppercase tracking-wider block mb-1">Nội dung chi tiết & Hình ảnh</label>
                <div className="p-3 bg-gray-50 border rounded-md text-xs space-y-3 leading-relaxed">
                  {previewRevision.content.split('\n').map((line, idx) => {
                    const imgMatch = line.match(/!\[.*?\]\((.*?)\)/);
                    if (imgMatch && imgMatch[1]) {
                      const imgUrl = mediaUrl(imgMatch[1]);
                      return (
                        <div key={idx} className="my-2">
                          <img src={imgUrl} alt="Revision content" className="h-40 object-cover rounded border shadow-sm" />
                        </div>
                      );
                    }
                    return line.trim() ? <p key={idx}>{line}</p> : null;
                  })}
                </div>
              </div>
            </div>

            <div className="p-4 border-t bg-gray-50 flex justify-end gap-3">
              <button
                type="button"
                onClick={() => setPreviewRevision(null)}
                className="px-4 py-2 border rounded-md text-xs font-medium cursor-pointer"
              >
                Đóng
              </button>
              <button
                type="button"
                onClick={() => {
                  const rev = previewRevision;
                  setPreviewRevision(null);
                  handleRestore(rev);
                }}
                className="px-4 py-2 bg-amber-500 text-white rounded-md text-xs font-medium cursor-pointer"
              >
                ↩️ Khôi phục phiên bản này
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default function EditPostPage() {
  return (
    <Suspense fallback={<div className="text-center mt-10">Đang tải trang chỉnh sửa...</div>}>
      <EditPostForm />
    </Suspense>
  );
}