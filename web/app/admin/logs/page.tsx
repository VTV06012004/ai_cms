'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import axios from 'axios';
import { apiUrl } from '@/lib/api';

interface User {
  id: number;
  name: string;
  email: string;
  role: string;
}

interface AuditLog {
  id: number;
  user_id: number;
  user?: User;
  action: string;
  target: string;
  details: string;
  ip_address: string;
  created_at: string;
}

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const router = useRouter();

  useEffect(() => {
    const fetchAuditLogs = async () => {
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          router.push('/login');
          return;
        }

        if (localStorage.getItem('userRole') !== 'admin') {
          alert('Bạn không có quyền truy cập vào nhật ký hệ thống!');
          router.push('/posts');
          return;
        }

        const res = await axios.get(apiUrl('audit-logs'), {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        setLogs(res.data.data || []);
      } catch (err: any) {
        console.error('Lỗi lấy danh sách Audit Logs:', err);
        if (err.response?.status === 403) {
          setError('Bạn không có quyền xem nhật ký hệ thống.');
        } else {
          setError('Không thể tải nhật ký hệ thống. Vui lòng thử lại sau.');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchAuditLogs();
  }, [router]);

  return (
    <div className="min-h-screen bg-gray-50 p-6 text-gray-800">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">🛡️ Nhật ký hệ thống (Audit Logs)</h1>
            <p className="text-sm text-gray-500 mt-1">
              Theo dõi và kiểm soát toàn bộ hoạt động của người dùng trên CMS.
            </p>
          </div>
          <button
            onClick={() => router.push('/posts')}
            className="px-4 py-2 bg-white border border-gray-300 text-gray-700 rounded-md hover:bg-gray-100 text-sm font-medium transition cursor-pointer"
          >
            ← Quản lý bài viết
          </button>
        </div>

        {/* Thông báo lỗi */}
        {error && (
          <div className="p-4 mb-4 text-sm text-red-700 bg-red-100 rounded-lg border border-red-200">
            {error}
          </div>
        )}

        {/* Bảng Dữ Liệu */}
        {loading ? (
          <div className="bg-white rounded-lg shadow p-8 text-center text-gray-500">
            Đang tải dữ liệu nhật ký hệ thống...
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow overflow-hidden border border-gray-200">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-sm">
                <thead className="bg-gray-100 text-gray-600 uppercase text-xs font-semibold">
                  <tr>
                    <th className="px-4 py-3 text-left">ID</th>
                    <th className="px-4 py-3 text-left">Thời gian</th>
                    <th className="px-4 py-3 text-left">Người thực hiện</th>
                    <th className="px-4 py-3 text-left">Hành động</th>
                    <th className="px-4 py-3 text-left">Đối tượng</th>
                    <th className="px-4 py-3 text-left">Chi tiết</th>
                    <th className="px-4 py-3 text-left">IP Address</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 bg-white">
                  {logs.map((log) => (
                    <tr key={log.id} className="hover:bg-gray-50 transition">
                      <td className="px-4 py-3 font-mono text-xs text-gray-400">#{log.id}</td>
                      <td className="px-4 py-3 text-xs text-gray-500 whitespace-nowrap">
                        {new Date(log.created_at).toLocaleString('vi-VN')}
                      </td>
                      <td className="px-4 py-3">
                        <div className="font-medium text-gray-900">{log.user?.name || 'Hệ thống'}</div>
                        <div className="text-xs text-gray-400">{log.user?.email}</div>
                      </td>
                      <td className="px-4 py-3">
                        {/* Phân màu sắc theo đúng yêu cầu */}
                        <span className={`inline-block px-2.5 py-1 text-xs font-bold rounded border ${
                          log.action === 'DELETE_POST' 
                            ? 'bg-red-50 text-red-700 border-red-200'         // Đỏ cho Xóa
                            : log.action === 'UPDATE_POST' 
                            ? 'bg-yellow-50 text-yellow-700 border-yellow-200' // Vàng cho Sửa
                            : log.action === 'LOGIN' 
                            ? 'bg-green-50 text-green-700 border-green-200'    // Xanh lá cho Đăng nhập
                            : log.action === 'CREATE_POST' 
                            ? 'bg-blue-50 text-blue-700 border-blue-200'       // Xanh dương cho Tạo mới
                            : 'bg-gray-50 text-gray-700 border-gray-200'
                        }`}>
                          {log.action}
                        </span>
                      </td>
                      <td className="px-4 py-3 font-medium text-gray-700">{log.target}</td>
                      <td className="px-4 py-3 text-xs text-gray-600 max-w-xs truncate" title={log.details}>
                        {log.details || '-'}
                      </td>
                      <td className="px-4 py-3 font-mono text-xs text-gray-500">{log.ip_address}</td>
                    </tr>
                  ))}

                  {logs.length === 0 && (
                    <tr>
                      <td colSpan={7} className="px-4 py-8 text-center text-gray-400">
                        Chưa có bản ghi nhật ký hoạt động nào.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}