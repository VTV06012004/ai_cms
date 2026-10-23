'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import axios from 'axios';

interface User {
  id: number;
  email: string;
  role: string;
  created_at: string;
}

export default function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [updatingId, setUpdatingId] = useState<number | null>(null);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('token');
    const role = localStorage.getItem('userRole');
    const storedUserId = localStorage.getItem('userId'); // Lấy ID của chính admin đang đăng nhập

    setCurrentUserId(storedUserId);

    // Kiểm tra bảo mật phía client: Chỉ cho phép admin truy cập
    if (!token || (role || '').toLowerCase() !== 'admin') {
      alert('Bạn không có quyền truy cập trang này!');
      router.push('/posts');
      return;
    }

    const fetchUsers = async () => {
      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
        const res = await axios.get(`${apiUrl}/admin/users`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        setUsers(res.data.data || []);
      } catch (err: any) {
        console.error('Lỗi tải danh sách user:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, [router]);

  const handleRoleChange = async (userId: number, newRole: string) => {
    // Chặn tuyệt đối nếu tự đổi quyền chính mình
    if (currentUserId && userId.toString() === currentUserId) {
      alert('Bạn không thể tự thay đổi quyền của chính tài khoản mình đang đăng nhập!');
      return;
    }

    try {
      setUpdatingId(userId);
      const token = localStorage.getItem('token');
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

      await axios.put(
        `${apiUrl}/admin/users/${userId}/role`,
        { role: newRole },
        { headers: { Authorization: `Bearer ${token}` } }
      );

      setUsers((prev) =>
        prev.map((u) => (u.id === userId ? { ...u, role: newRole } : u))
      );
      alert('Cập nhật quyền thành công!');
    } catch (err: any) {
      alert(err.response?.data?.error || 'Cập nhật quyền thất bại!');
    } finally {
      setUpdatingId(null);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-6xl mx-auto">
        {/* Header trang */}
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6 bg-white p-6 rounded-xl shadow-sm border">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">👥 Quản lý Người dùng & Phân quyền</h1>
            <p className="text-sm text-gray-500 mt-1">Quản lý và phân quyền tài khoản khác trong hệ thống (Không được tự đổi quyền chính mình).</p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => router.push('/posts')}
              className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-semibold rounded-lg transition cursor-pointer"
            >
              ← Quay lại Bài viết
            </button>
            <button
              onClick={() => {
                localStorage.clear();
                router.push('/login');
              }}
              className="px-4 py-2 bg-red-50 hover:bg-red-100 text-red-600 text-sm font-semibold rounded-lg transition cursor-pointer"
            >
              Đăng xuất
            </button>
          </div>
        </div>

        {/* Nội dung danh sách */}
        {loading ? (
          <div className="text-center py-16 bg-white rounded-xl shadow-sm border text-gray-500">
            Đang tải danh sách người dùng...
          </div>
        ) : (
          <div className="bg-white rounded-xl shadow-sm border overflow-hidden">
            <table className="w-full text-left border-collapse text-sm">
              <thead>
                <tr className="bg-gray-50 border-b text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  <th className="py-3.5 px-6">ID</th>
                  <th className="py-3.5 px-6">Email / Tài khoản</th>
                  <th className="py-3.5 px-6">Vai trò hiện tại</th>
                  <th className="py-3.5 px-6">Ngày tạo</th>
                  <th className="py-3.5 px-6 text-center">Hành động (Phân quyền)</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {users.map((user) => {
                  const isMe = currentUserId && user.id.toString() === currentUserId;

                  return (
                    <tr key={user.id} className="hover:bg-gray-50/80 transition">
                      <td className="py-4 px-6 font-mono text-gray-500">
                        #{user.id} {isMe && <span className="text-xs text-blue-600 font-bold">(Bạn)</span>}
                      </td>
                      <td className="py-4 px-6 font-medium text-gray-900">{user.email}</td>
                      <td className="py-4 px-6">
                        <span className={`px-3 py-1 text-xs font-bold rounded-full border ${
                          user.role.toLowerCase() === 'admin' 
                            ? 'bg-purple-50 text-purple-700 border-purple-200' 
                            : user.role.toLowerCase() === 'editor'
                            ? 'bg-blue-50 text-blue-700 border-blue-200'
                            : 'bg-green-50 text-green-700 border-green-200'
                        }`}>
                          {user.role.toUpperCase()}
                        </span>
                      </td>
                      <td className="py-4 px-6 text-gray-500 text-xs">
                        {new Date(user.created_at).toLocaleDateString('vi-VN')}
                      </td>
                      <td className="py-4 px-6 text-center">
                        {isMe ? (
                          <span className="text-xs text-gray-400 italic">Không thể tự đổi quyền</span>
                        ) : (
                          <select
                            disabled={updatingId === user.id}
                            value={user.role.toLowerCase()}
                            onChange={(e) => handleRoleChange(user.id, e.target.value)}
                            className="px-3 py-1.5 text-xs border rounded-lg bg-white shadow-sm focus:ring-2 focus:ring-blue-500 focus:outline-none cursor-pointer disabled:opacity-50 disabled:bg-gray-100"
                          >
                            <option value="author">AUTHOR</option>
                            <option value="editor">EDITOR</option>
                            <option value="admin">ADMIN</option>
                          </select>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}