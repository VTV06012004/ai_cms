'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import axios from 'axios';
import { apiUrl, mediaUrl } from '@/lib/api';

interface Category {
  id: number;
  name: string;
  slug: string;
}

interface User {
  id: number;
  email: string;
  name?: string;
  role: string;
}

interface Post {
  id: number;
  title: string;
  slug: string;
  content: string;
  status: string;
  rejection_reason?: string;
  category_id: number;
  user_id?: number;
  author_id?: number;
  category?: Category;
  user?: User;
  author?: User;
  created_at: string;
}

const getCategoryStyle = (categoryName: string = '') => {
  const name = categoryName.toLowerCase().trim();
  if (name.includes('technology') || name.includes('tech') || name.includes('công nghệ')) {
    return { icon: '💻', bg: 'bg-blue-50 text-blue-700 border-blue-200' };
  }
  if (name.includes('lifestyle') || name.includes('đời sống') || name.includes('sống')) {
    return { icon: '🌿', bg: 'bg-emerald-50 text-emerald-700 border-emerald-200' };
  }
  if (name.includes('news') || name.includes('tin tức') || name.includes('báo')) {
    return { icon: '📰', bg: 'bg-purple-50 text-purple-700 border-purple-200' };
  }
  return { icon: '📌', bg: 'bg-gray-100 text-gray-700 border-gray-200' };
};

export default function PostsPage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  // 👈 Cho phép chọn 'all', id chuyên mục, hoặc 'my-posts'
  const [selectedCategory, setSelectedCategory] = useState<number | 'all' | 'my-posts'>('all');
  const [loading, setLoading] = useState<boolean>(true);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [previewPost, setPreviewPost] = useState<Post | null>(null);

  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('userRole');
    const userId = localStorage.getItem('userId');
    const userName = localStorage.getItem('userName');
    const userEmail = localStorage.getItem('userEmail');

    if (token) {
      setCurrentUser({
        id: Number(userId || 0),
        email: userEmail || 'user@cms.com',
        name: userName || undefined,
        role: role || 'author'
      });
    }

    const fetchData = async () => {
      try {
        const [postsRes, catRes] = await Promise.all([
          axios.get(apiUrl('posts')),
          axios.get(apiUrl('categories'))
        ]);

        setPosts(postsRes.data.data || postsRes.data || []);
        setCategories(catRes.data.data || catRes.data || []);
      } catch (err) {
        console.error('Lỗi khi tải dữ liệu bài viết:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('userRole');
    localStorage.removeItem('userId');
    localStorage.removeItem('userName');
    localStorage.removeItem('userEmail');
    router.push('/login');
  };

  const handleDeletePost = async (id: number, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm('Bạn có chắc chắn muốn xóa bài viết này không?')) return;

    try {
      const token = localStorage.getItem('token');
      await axios.delete(apiUrl(`posts/${id}`), {
        headers: { Authorization: `Bearer ${token}` }
      });

      setPosts(posts.filter((p) => p.id !== id));
    } catch (err: any) {
      alert(err.response?.data?.error || 'Xóa bài viết thất bại!');
    }
  };

  const handleModerate = async (postId: number, action: 'approve' | 'reject') => {
    const newStatus = action === 'approve' ? 'published' : 'rejected';
    let rejection_reason = '';

    if (newStatus === 'rejected') {
      rejection_reason = prompt('Nhập lý do từ chối bài viết này:') || '';
      if (!rejection_reason.trim()) {
        alert('Bạn bắt buộc phải nhập lý do từ chối!');
        return;
      }
    }

    try {
      const token = localStorage.getItem('token');
      await axios.put(
        apiUrl(`posts/${postId}/moderate`),
        { status: newStatus, rejection_reason },
        { headers: { Authorization: `Bearer ${token}` } }
      );

      alert(`Đã ${action === 'approve' ? 'duyệt bài viết (Published)' : 'từ chối bài viết'} thành công!`);
      
      setPosts((prev) =>
        prev.map((p) => (p.id === postId ? { ...p, status: newStatus, rejection_reason } : p))
      );
    } catch (err: any) {
      alert(err.response?.data?.error || 'Thao tác kiểm duyệt thất bại!');
    }
  };

  const userRoleLower = (currentUser?.role || '').toLowerCase();
  const currentUserId = Number(currentUser?.id || 0);

  // 🛡️ LỌC BÀI VIẾT THEO QUY TẮC CHUNG
  const visiblePosts = posts.filter((post) => {
    if (userRoleLower === 'admin' || userRoleLower === 'editor') {
      return true;
    }

    const postUserId = Number(post.user_id || post.author_id || post.user?.id || post.author?.id || 0);
    const isOwner = currentUserId > 0 && postUserId > 0 && currentUserId === postUserId;
    const isPublished = post.status === 'published';

    return isOwner || isPublished;
  });

  // 🏷️ LỌC THEO CHUYÊN MỤC HOẶC "BÀI VIẾT CỦA TÔI"
  const filteredPosts = visiblePosts.filter((post) => {
    const postUserId = Number(post.user_id || post.author_id || post.user?.id || post.author?.id || 0);
    const isOwner = currentUserId > 0 && postUserId > 0 && currentUserId === postUserId;

    // Nếu chọn xem "Bài viết của tôi"
    if (selectedCategory === 'my-posts') {
      return isOwner;
    }

    // Nếu chọn "Tất cả" hoặc theo chuyên mục
    const matchesCategory =
      selectedCategory === 'all' ||
      Number(post.category_id || post.category?.id) === selectedCategory;
    return matchesCategory;
  });

  // Đếm số lượng bài viết của riêng user hiện tại
  const myPostsCount = visiblePosts.filter((post) => {
    const postUserId = Number(post.user_id || post.author_id || post.user?.id || post.author?.id || 0);
    return currentUserId > 0 && postUserId > 0 && currentUserId === postUserId;
  }).length;

  return (
    <div className="min-h-screen bg-gray-50 text-gray-800">
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <h1 className="text-xl font-bold text-blue-600 cursor-pointer" onClick={() => router.push('/posts')}>
              🚀 AI CMS
            </h1>
            <nav className="hidden md:flex space-x-3 text-sm font-medium items-center">
              <button
                onClick={() => router.push('/posts')}
                className="text-blue-600 border-b-2 border-blue-600 pb-1 cursor-pointer"
              >
                Tất cả bài viết
              </button>
              <button
                onClick={() => router.push('/about')}
                className="text-gray-600 hover:text-blue-600 transition cursor-pointer"
              >
                Giới thiệu
              </button>

              {userRoleLower === 'admin' && (
                <>
                  <button
                    onClick={() => router.push('/admin/users')}
                    className="hover:text-indigo-600 transition cursor-pointer flex items-center gap-1 font-semibold text-indigo-600 bg-indigo-50 px-3 py-1 rounded-md border border-indigo-200"
                  >
                    👥 Quản lý Người dùng
                  </button>
                  <button
                    onClick={() => router.push('/admin/logs')}
                    className="hover:text-purple-600 transition cursor-pointer flex items-center gap-1 font-semibold text-purple-600 bg-purple-50 px-3 py-1 rounded-md border border-purple-200"
                  >
                    🛡️ Nhật ký hệ thống
                  </button>
                </>
              )}
            </nav>
          </div>

          <div className="flex items-center gap-3">
            {currentUser && (
              <span className="text-xs text-gray-600 bg-gray-100 px-3 py-1.5 rounded-full border font-medium">
                👤 {currentUser.name || currentUser.email} <span className="text-blue-600 font-bold">({currentUser.role})</span>
              </span>
            )}
            <button
              onClick={() => router.push('/posts/create')}
              className="px-3.5 py-2 bg-blue-600 text-white text-xs font-semibold rounded-lg hover:bg-blue-700 transition cursor-pointer shadow-sm"
            >
              + Tạo bài viết mới
            </button>
            <button
              onClick={handleLogout}
              className="px-3.5 py-2 bg-gray-200 text-gray-700 text-xs font-semibold rounded-lg hover:bg-gray-300 transition cursor-pointer"
            >
              Đăng xuất
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">Danh sách Bài viết</h2>
            <p className="text-sm text-gray-500 mt-1">
              Khám phá và lọc nội dung theo chuyên mục chủ đề hoặc bài viết cá nhân.
            </p>
          </div>

          <div className="flex items-center gap-2 overflow-x-auto pb-2 md:pb-0">
            {/* NÚT TẤT CẢ */}
            <button
              onClick={() => setSelectedCategory('all')}
              className={`px-3.5 py-1.5 text-xs font-semibold rounded-full border transition whitespace-nowrap cursor-pointer ${
                selectedCategory === 'all'
                  ? 'bg-blue-600 text-white border-blue-600 shadow-sm'
                  : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-100'
              }`}
            >
              🌐 Tất cả ({visiblePosts.length})
            </button>

            {/* ✍️ NÚT LỌC RIÊNG BÀI CỦA TÔI */}
            <button
              onClick={() => setSelectedCategory('my-posts')}
              className={`px-3.5 py-1.5 text-xs font-semibold rounded-full border transition whitespace-nowrap cursor-pointer flex items-center gap-1 ${
                selectedCategory === 'my-posts'
                  ? 'bg-indigo-600 text-white border-indigo-600 shadow-sm'
                  : 'bg-white text-indigo-600 border-indigo-200 hover:bg-indigo-50'
              }`}
            >
              <span>✍️</span>
              <span>Bài viết của tôi ({myPostsCount})</span>
            </button>

            {/* CÁC CHUYÊN MỤC */}
            {categories.map((cat) => {
              const count = visiblePosts.filter(
                (p) => (p.category_id || p.category?.id) === cat.id
              ).length;
              const style = getCategoryStyle(cat.name);
              return (
                <button
                  key={cat.id}
                  onClick={() => setSelectedCategory(cat.id)}
                  className={`px-3.5 py-1.5 text-xs font-semibold rounded-full border transition whitespace-nowrap cursor-pointer flex items-center gap-1.5 ${
                    selectedCategory === cat.id
                      ? 'bg-blue-600 text-white border-blue-600 shadow-sm'
                      : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-100'
                  }`}
                >
                  <span>{style.icon}</span>
                  <span>{cat.name} ({count})</span>
                </button>
              );
            })}
          </div>
        </div>

        {loading ? (
          <div className="bg-white rounded-lg p-12 text-center text-gray-500 shadow-sm border">
            Đang tải dữ liệu bài viết...
          </div>
        ) : filteredPosts.length === 0 ? (
          <div className="bg-white rounded-lg p-12 text-center text-gray-400 border italic">
            Không có bài viết nào phù hợp.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredPosts.map((post) => {
              const postCat =
                post.category ||
                categories.find((c) => c.id === post.category_id);

              const catStyle = getCategoryStyle(postCat?.name);

              const imgMatch = post.content.match(/!\[.*?\]\((.*?)\)/);
              const thumbnailUrl = imgMatch
                ? imgMatch[1].startsWith('http')
                  ? imgMatch[1]
                  : mediaUrl(imgMatch[1])
                : null;

              const postUserId = Number(
                post.user_id || 
                post.author_id || 
                post.user?.id || 
                post.author?.id || 
                0
              );

              const isOwner = Boolean(
                currentUser && (
                  userRoleLower === 'admin' ||
                  userRoleLower === 'editor' ||
                  (currentUserId > 0 && postUserId > 0 && currentUserId === postUserId)
                )
              );

              return (
                <div
                  key={post.id}
                  className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition flex flex-col justify-between"
                >
                  <div>
                    {thumbnailUrl && (
                      <div className="h-44 w-full bg-gray-100 overflow-hidden border-b cursor-pointer" onClick={() => setPreviewPost(post)}>
                        <img
                          src={thumbnailUrl}
                          alt={post.title}
                          className="w-full h-full object-cover hover:scale-105 transition duration-300"
                        />
                      </div>
                    )}

                    <div className="p-5 space-y-3">
                      <div className="flex items-center justify-between gap-2">
                        <span className={`inline-flex items-center gap-1 px-2.5 py-1 text-xs font-bold rounded-md border ${catStyle.bg}`}>
                          <span>{catStyle.icon}</span> {postCat?.name || 'Chưa phân loại'}
                        </span>
                        <span className="text-xs text-gray-400">
                          Tác giả ID: <span className="font-medium text-gray-600">{postUserId || 'N/A'}</span>
                        </span>
                      </div>

                      <div className="flex items-center gap-2 flex-wrap">
                        <span className={`px-2.5 py-0.5 text-[10px] font-bold rounded-full uppercase border ${
                          post.status === 'published'
                            ? 'bg-green-50 text-green-700 border-green-200'
                            : post.status === 'review'
                            ? 'bg-amber-50 text-amber-700 border-amber-200'
                            : post.status === 'rejected'
                            ? 'bg-red-50 text-red-700 border-red-200'
                            : 'bg-gray-100 text-gray-600 border-gray-200'
                        }`}>
                          ● {post.status}
                        </span>
                      </div>

                      {post.status === 'rejected' && post.rejection_reason && (
                        <div className="p-2.5 bg-red-50 border border-red-200 rounded-lg text-xs text-red-600 leading-relaxed">
                          <span className="font-bold">Lý do từ chối:</span> {post.rejection_reason}
                        </div>
                      )}

                      <h3
                        onClick={() => setPreviewPost(post)}
                        className="font-bold text-lg text-gray-900 hover:text-blue-600 transition cursor-pointer line-clamp-2"
                      >
                        {post.title}
                      </h3>

                      <p className="text-xs text-gray-500 line-clamp-3 leading-relaxed">
                        {post.content.replace(/!\[.*?\]\((.*?)\)/g, '').trim()}
                      </p>
                    </div>
                  </div>

                  <div className="px-5 py-3 bg-gray-50 border-t border-gray-100 flex items-center justify-between text-xs">
                    <span className="text-gray-400">{new Date(post.created_at).toLocaleDateString('vi-VN')}</span>
                    
                    <div className="flex items-center gap-2">
                      {(userRoleLower === 'admin' || userRoleLower === 'editor') && post.status === 'review' && (
                        <div className="flex items-center gap-1 bg-amber-50 border border-amber-200 p-1 rounded-lg">
                          <button
                            onClick={() => handleModerate(post.id, 'approve')}
                            className="px-2 py-0.5 bg-green-600 text-white rounded font-medium hover:bg-green-700 transition cursor-pointer"
                            title="Duyệt bài"
                          >
                            Duyệt
                          </button>
                          <button
                            onClick={() => handleModerate(post.id, 'reject')}
                            className="px-2 py-0.5 bg-red-600 text-white rounded font-medium hover:bg-red-700 transition cursor-pointer"
                            title="Từ chối bài viết"
                          >
                            Từ chối
                          </button>
                        </div>
                      )}

                      <button
                        onClick={() => setPreviewPost(post)}
                        className="text-gray-600 font-semibold hover:text-blue-600 transition cursor-pointer px-2 py-1"
                      >
                        Xem
                      </button>

                      {isOwner && (
                        <>
                          <button
                            onClick={() => router.push(`/posts/edit/${post.id}`)}
                            className="text-blue-600 font-semibold hover:underline cursor-pointer"
                          >
                            Sửa
                          </button>
                          <button
                            onClick={(e) => handleDeletePost(post.id, e)}
                            className="text-red-600 font-semibold hover:underline cursor-pointer"
                          >
                            Xóa
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </main>

      {previewPost && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl max-w-2xl w-full max-h-[85vh] overflow-y-auto shadow-2xl p-6 space-y-4">
            <div className="flex items-center justify-between border-b pb-3">
              {(() => {
                const previewCat = previewPost.category || categories.find((c) => c.id === previewPost.category_id);
                const previewStyle = getCategoryStyle(previewCat?.name);
                return (
                  <span className={`inline-flex items-center gap-1 px-2.5 py-1 text-xs font-bold rounded-md border ${previewStyle.bg}`}>
                    <span>{previewStyle.icon}</span> {previewCat?.name || 'Chuyên mục'}
                  </span>
                );
              })()}
              <button
                onClick={() => setPreviewPost(null)}
                className="text-gray-400 hover:text-gray-700 font-bold text-lg cursor-pointer"
              >
                ✕
              </button>
            </div>

            <h2 className="text-2xl font-bold text-gray-900">{previewPost.title}</h2>

            <div className="text-xs text-gray-500 flex items-center justify-between border-b pb-3">
              <span>Mã tác giả ID: {previewPost.user_id || previewPost.author_id}</span>
              <span>Ngày đăng: {new Date(previewPost.created_at).toLocaleDateString('vi-VN')}</span>
            </div>

            <div className="text-sm text-gray-700 space-y-4 leading-relaxed py-2">
              {previewPost.content.split('\n').map((paragraph, index) => {
                const imgMatch = paragraph.match(/!\[(.*?)\]\((.*?)\)/);
                if (imgMatch) {
                  const altText = imgMatch[1];
                  const rawUrl = imgMatch[2];
                  const imgUrl = mediaUrl(rawUrl);
                  return (
                    <div key={index} className="my-3 rounded-lg overflow-hidden border bg-gray-50">
                      <img src={imgUrl} alt={altText} className="max-h-80 mx-auto object-contain" />
                    </div>
                  );
                }
                return paragraph.trim() ? <p key={index}>{paragraph}</p> : null;
              })}
            </div>

            <div className="flex justify-end pt-4 border-t">
              <button
                onClick={() => setPreviewPost(null)}
                className="px-4 py-2 bg-gray-200 text-gray-700 text-xs font-semibold rounded-lg hover:bg-gray-300 transition cursor-pointer"
              >
                Đóng
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}