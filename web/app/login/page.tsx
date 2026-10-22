'use client';

import { useState } from 'react';
import axios from 'axios';
import { useRouter } from 'next/navigation';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
      const response = await axios.post(`${apiUrl}/auth/login`, {
        email,
        password,
      });

      const { token, user } = response.data;

      localStorage.setItem('token', token);
      localStorage.setItem('userRole', user?.role || 'author');
      localStorage.setItem('userName', user?.name || 'User');

      const currentUserId = user?.id || user?.ID || response.data.user_id || 0;
      localStorage.setItem('userId', user?.id || user?.ID || '');
      alert(`Đăng nhập thành công với vai trò: ${user?.role || 'author'}`);

      // Chuyển hướng về trang quản lý bài viết
      router.push('/posts');
    } catch (err: any) {
      console.error('Lỗi đăng nhập:', err);
      setError(err.response?.data?.error || 'Email hoặc mật khẩu không chính xác!');
    } finally {
      setLoading(false);
    }
  };

  const handleQuickSeedLogin = (selectedRole: string) => {
    // Dùng để test nhanh các tài khoản mẫu nếu cần
    setEmail(`${selectedRole}@cms.com`);
    setPassword('123456');
  };

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-xl shadow-md overflow-hidden p-8">
        <div className="text-center mb-8">
          <h2 className="text-2xl font-bold text-gray-900">Đăng nhập hệ thống CMS</h2>
          <p className="text-sm text-gray-500 mt-1">Go Backend + Next.js 16 (RBAC)</p>
        </div>

        {error && (
          <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none text-gray-900"
              placeholder="admin@cms.com"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Mật khẩu</label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:outline-none text-gray-900"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white py-2.5 rounded-lg font-semibold shadow transition cursor-pointer disabled:opacity-50"
          >
            {loading ? 'Đang xác thực...' : 'Đăng nhập'}
          </button>
        </form>

        {/* Nút gợi ý test nhanh 3 vai trò */}
        <div className="mt-8 border-t border-gray-200 pt-4">
          <p className="text-xs text-gray-400 text-center mb-3">Chọn nhanh tài khoản test mẫu:</p>
          <div className="flex justify-center gap-2">
            <button 
              onClick={() => handleQuickSeedLogin('admin')}
              className="px-3 py-1 bg-purple-100 text-purple-700 text-xs font-semibold rounded hover:bg-purple-200 cursor-pointer"
            >
              Admin
            </button>
            <button 
              onClick={() => handleQuickSeedLogin('editor')}
              className="px-3 py-1 bg-blue-100 text-blue-700 text-xs font-semibold rounded hover:bg-blue-200 cursor-pointer"
            >
              Editor
            </button>
            <button 
              onClick={() => handleQuickSeedLogin('author')}
              className="px-3 py-1 bg-green-100 text-green-700 text-xs font-semibold rounded hover:bg-green-200 cursor-pointer"
            >
              Author
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}