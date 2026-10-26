'use client';

import axios from 'axios';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { apiUrl } from '@/lib/api';

export default function RegisterPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError('');
    setLoading(true);

    try {
      await axios.post(apiUrl('auth/register'), { name, email, password });
      window.alert('Tài khoản đã được tạo. Vui lòng đăng nhập.');
      router.push('/login');
    } catch (registrationError) {
      if (axios.isAxiosError(registrationError)) {
        setError(registrationError.response?.data?.error || 'Không thể tạo tài khoản.');
      } else {
        setError('Đã xảy ra lỗi không xác định khi tạo tài khoản.');
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <section className="max-w-md w-full bg-white rounded-xl shadow-md p-8">
        <h1 className="text-2xl font-bold text-gray-900 text-center">Tạo tài khoản</h1>
        <p className="text-sm text-gray-500 text-center mt-2">
          Tài khoản đăng ký mới sẽ có vai trò Author.
        </p>
        {error && (
          <p role="alert" className="mt-5 bg-red-50 border border-red-200 text-red-700 p-3 rounded-lg text-sm">
            {error}
          </p>
        )}
        <form onSubmit={handleSubmit} className="mt-6 space-y-4">
          <label className="block text-sm font-medium text-gray-700">
            Tên hiển thị
            <input
              autoComplete="name"
              required
              value={name}
              onChange={(event) => setName(event.target.value)}
              className="mt-1 w-full px-4 py-2 border border-gray-300 rounded-lg text-gray-900"
            />
          </label>
          <label className="block text-sm font-medium text-gray-700">
            Email
            <input
              type="email"
              autoComplete="email"
              required
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              className="mt-1 w-full px-4 py-2 border border-gray-300 rounded-lg text-gray-900"
            />
          </label>
          <label className="block text-sm font-medium text-gray-700">
            Mật khẩu (ít nhất 6 ký tự)
            <input
              type="password"
              autoComplete="new-password"
              minLength={6}
              maxLength={72}
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              className="mt-1 w-full px-4 py-2 border border-gray-300 rounded-lg text-gray-900"
            />
          </label>
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white py-2.5 rounded-lg font-semibold disabled:opacity-50"
          >
            {loading ? 'Đang tạo tài khoản...' : 'Đăng ký'}
          </button>
        </form>
        <p className="mt-6 text-center text-sm text-gray-600">
          Đã có tài khoản?{' '}
          <Link href="/login" className="font-semibold text-blue-600 hover:underline">
            Đăng nhập
          </Link>
        </p>
      </section>
    </main>
  );
}
