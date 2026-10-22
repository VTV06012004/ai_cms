'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function AboutPage() {
  const [currentUser, setCurrentUser] = useState<any>(null);
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('userRole');
    const userId = localStorage.getItem('userId');
    const userName = localStorage.getItem('userName');

    if (token) {
      setCurrentUser({
        id: Number(userId || 0),
        email: userName || 'user@cms.com',
        role: role || 'author'
      });
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('userRole');
    localStorage.removeItem('userId');
    localStorage.removeItem('userName');
    router.push('/login');
  };

  const userRoleLower = (currentUser?.role || '').toLowerCase();

  return (
    <div className="min-h-screen bg-gray-50 text-gray-800">
      {/* HEADER & NAVIGATION */}
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <h1 className="text-xl font-bold text-blue-600 cursor-pointer" onClick={() => router.push('/posts')}>
              🚀 AI CMS
            </h1>
            <nav className="hidden md:flex space-x-4 text-sm font-medium items-center">
              <button
                onClick={() => router.push('/posts')}
                className="text-gray-600 hover:text-blue-600 transition cursor-pointer"
              >
                Tất cả bài viết
              </button>
              <button
                onClick={() => router.push('/about')}
                className="text-blue-600 border-b-2 border-blue-600 pb-1 cursor-pointer font-semibold"
              >
                Giới thiệu
              </button>
              {userRoleLower === 'admin' && (
                <button
                  onClick={() => router.push('/admin/logs')}
                    className="hover:text-blue-600 transition cursor-pointer flex items-center gap-1 font-semibold text-purple-600 bg-purple-50 px-3 py-1 rounded-md border border-purple-200"
                >
                  🛡️ Nhật ký hệ thống (Audit Logs)
                </button>
              )}
            </nav>
          </div>

          <div className="flex items-center gap-3">
            {currentUser && (
              <span className="text-xs text-gray-600 bg-gray-100 px-3 py-1.5 rounded-full border font-medium">
                👤 {currentUser.email} <span className="text-blue-600 font-bold">({currentUser.role})</span>
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

      {/* MAIN CONTENT */}
      <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-8 md:p-12 space-y-8">
          <div className="text-center space-y-3">
            <span className="inline-block px-3 py-1 text-xs font-bold rounded-full bg-blue-50 text-blue-600 border border-blue-200">
              Về hệ thống quản trị
            </span>
            <h2 className="text-3xl font-extrabold text-gray-900">Content Management System</h2>
            <p className="text-base text-gray-500 max-w-2xl mx-auto">
              Hệ thống quản trị nội dung hiện đại, hiệu năng cao được xây dựng để tối ưu hóa việc xuất bản, phân quyền và quản lý tài nguyên.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-4">
            <div className="bg-gray-50 p-6 rounded-xl border border-gray-100 space-y-2">
              <div className="text-2xl">⚡</div>
              <h3 className="font-bold text-gray-900">Frontend (Next.js)</h3>
              <p className="text-xs text-gray-500 leading-relaxed">
                Xây dựng với Next.js App Router, Tailwind CSS mang lại trải nghiệm giao diện mượt mà, tốc độ tải siêu nhanh và responsive đầy đủ.
              </p>
            </div>

            <div className="bg-gray-50 p-6 rounded-xl border border-gray-100 space-y-2">
              <div className="text-2xl">⚙️</div>
              <h3 className="font-bold text-gray-900">Backend (Go + Gin)</h3>
              <p className="text-xs text-gray-500 leading-relaxed">
                Sử dụng ngôn ngữ Go với framework Gin siêu nhẹ, xử lý API RESTful an toàn, quản lý static file tải lên hiệu quả.
              </p>
            </div>

            <div className="bg-gray-50 p-6 rounded-xl border border-gray-100 space-y-2">
              <div className="text-2xl">🛡️</div>
              <h3 className="font-bold text-gray-900">Phân quyền RBAC</h3>
              <p className="text-xs text-gray-500 leading-relaxed">
                Phân chia rõ ràng quyền hạn giữa Admin, Editor và Author giúp bảo mật dữ liệu và kiểm soát hành động người dùng tối ưu.
              </p>
            </div>
          </div>

          <div className="border-t pt-8 space-y-4">
            <h3 className="font-bold text-lg text-gray-900">Tính năng nổi bật</h3>
            <ul className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm text-gray-600">
              <li className="flex items-center gap-2">✅ Quản lý danh mục & bài viết trực quan</li>
              <li className="flex items-center gap-2">✅ Hỗ trợ xem trước (Preview) nội dung Markdown</li>
              <li className="flex items-center gap-2">✅ Tải lên & quản lý hình ảnh đính kèm an toàn</li>
              <li className="flex items-center gap-2">✅ Nhật ký hệ thống (Audit Logs) cho quản trị viên</li>
            </ul>
          </div>

          <div className="flex justify-center pt-4">
            <button
              onClick={() => router.push('/posts')}
              className="px-6 py-3 bg-blue-600 text-white font-semibold rounded-xl hover:bg-blue-700 transition shadow cursor-pointer text-sm"
            >
              Khám phá danh sách bài viết ngay ➔
            </button>
          </div>
        </div>
      </main>
    </div>
  );
}