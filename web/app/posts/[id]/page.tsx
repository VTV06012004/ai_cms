'use client';

import { useEffect, useState, Suspense } from 'react';
import { useRouter, useParams } from 'next/navigation';
import axios from 'axios';
import { apiUrl, mediaUrl } from '@/lib/api';

function PostDetailContent() {
  const [post, setPost] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  const router = useRouter();
  const params = useParams();
  const { id } = params;

  useEffect(() => {
    const fetchPostDetail = async () => {
      try {
        const res = await axios.get(apiUrl(`posts/${id}`));
        const postData = res.data.data || res.data;
        setPost(postData);
        setLoading(false);
      } catch (err) {
        console.error('Lỗi tải bài viết:', err);
        setError('Không tìm thấy bài viết hoặc bài viết đã bị xóa.');
        setLoading(false);
      }
    };

    if (id) {
      fetchPostDetail();
    }
  }, [id]);

  if (loading) {
    return <div className="text-center py-20 text-gray-500 font-medium">Đang tải nội dung bài viết...</div>;
  }

  if (error || !post) {
    return (
      <div className="max-w-2xl mx-auto mt-20 p-6 bg-red-50 border border-red-200 rounded-lg text-center text-red-700">
        <p className="mb-4">{error || 'Không có dữ liệu bài viết.'}</p>
        <button
          onClick={() => router.push('/posts')}
          className="px-4 py-2 bg-gray-800 text-white rounded hover:bg-black"
        >
          Quay lại danh sách
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-10 px-4 text-black">
      <div className="max-w-3xl mx-auto bg-white p-8 rounded-xl shadow-sm border border-gray-200">
        <button
          onClick={() => router.push('/posts')}
          className="text-sm text-blue-600 hover:underline mb-6 inline-block cursor-pointer"
        >
          ← Quay lại danh sách quản trị
        </button>

        <h1 className="text-3xl font-extrabold text-gray-900 mb-4">{post.title}</h1>

        <div className="flex flex-wrap items-center gap-4 text-sm text-gray-500 pb-6 mb-6 border-b border-gray-100">
          <span>Tác giả: <strong className="text-gray-700">{post.author?.name || post.author?.username || 'Admin'}</strong></span>
          <span>•</span>
          <span>Danh mục: <strong className="text-blue-600">{post.category?.name || '---'}</strong></span>
          <span>•</span>
          <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${
            post.status === 'published' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
          }`}>
            {post.status}
          </span>
        </div>

        <div className="prose max-w-none text-gray-800 leading-relaxed whitespace-pre-wrap">
        {/* Chia nhỏ nội dung để tìm cú pháp hình ảnh markdown ![alt](url) và render thành ảnh */}
        {post.content.split('\n').map((line: string, index: number) => {
            const imgMatch = line.match(/!\[.*?\]\((.*?)\)/);
            if (imgMatch) {
            const imgUrl = mediaUrl(imgMatch[1]);
            return (
                <div key={index} className="my-4">
                <img src={imgUrl} alt="Post content" className="max-h-96 rounded-lg shadow-sm border" />
                </div>
            );
            }
            return <p key={index} className="mb-2">{line}</p>;
        })}
        </div>

        <div className="mt-10 pt-6 border-t border-gray-100 flex gap-3">
          <button
            onClick={() => router.push(`/posts/edit/${post.ID || post.id}`)}
            className="px-4 py-2 bg-yellow-500 text-white rounded-md hover:bg-yellow-600 text-sm font-medium cursor-pointer"
          >
            Chỉnh sửa bài viết này
          </button>
        </div>
      </div>
    </div>
  );
}

export default function PostDetailPage() {
  return (
    <Suspense fallback={<div className="text-center mt-20">Đang tải...</div>}>
      <PostDetailContent />
    </Suspense>
  );
}