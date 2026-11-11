'use client';

import { useCallback, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { isAxiosError } from 'axios';
import { apiClient, apiUrl } from '@/lib/api';
import { readSession } from '@/lib/session';

interface Category {
  id: number;
  name: string;
  slug: string;
  description: string;
  parent_id: number | null;
}

interface Tag {
  id: number;
  name: string;
  slug: string;
}

interface MenuItemDraft {
  title: string;
  type: 'custom' | 'category';
  target_id: number;
  url: string;
}

interface Menu {
  id: number;
  name: string;
  location: string;
  menu_items?: { title: string; type: string; target_id: number; url: string; sort_order: number }[];
}

type Notice = { type: 'ok' | 'error'; text: string } | null;
type Tab = 'categories' | 'tags' | 'menus';

function errorMessage(err: unknown): string {
  if (isAxiosError(err)) {
    return err.response?.data?.error || 'Thao tác thất bại';
  }
  return 'Thao tác thất bại';
}

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/đ/g, 'd')
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

const inputClass =
  'w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none bg-white';
const primaryButton =
  'px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-lg transition cursor-pointer disabled:opacity-50';
const secondaryButton =
  'px-3 py-1.5 bg-gray-100 hover:bg-gray-200 text-gray-700 text-xs font-semibold rounded-lg transition cursor-pointer';
const dangerButton =
  'px-3 py-1.5 bg-red-50 hover:bg-red-100 text-red-600 text-xs font-semibold rounded-lg transition cursor-pointer';

function NoticeBar({ notice }: { notice: Notice }) {
  if (!notice) return null;
  const color =
    notice.type === 'ok'
      ? 'bg-green-50 text-green-700 border-green-200'
      : 'bg-red-50 text-red-700 border-red-200';
  return <div className={`mb-4 p-3 text-sm rounded-lg border ${color}`}>{notice.text}</div>;
}

function CategoriesPanel() {
  const [items, setItems] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [notice, setNotice] = useState<Notice>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [slugTouched, setSlugTouched] = useState(false);
  const [description, setDescription] = useState('');
  const [parentId, setParentId] = useState('');

  const load = useCallback(async () => {
    try {
      const res = await apiClient.get(apiUrl('categories'));
      setItems(res.data.data || []);
    } catch (err) {
      setNotice({ type: 'error', text: errorMessage(err) });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // Tải dữ liệu khi vào trang; load() đặt trạng thái loading đồng bộ.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [load]);

  const resetForm = () => {
    setEditingId(null);
    setName('');
    setSlug('');
    setSlugTouched(false);
    setDescription('');
    setParentId('');
  };

  const startEdit = (item: Category) => {
    setEditingId(item.id);
    setName(item.name);
    setSlug(item.slug);
    setSlugTouched(true);
    setDescription(item.description || '');
    setParentId(item.parent_id ? String(item.parent_id) : '');
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      name,
      slug,
      description,
      parent_id: parentId ? Number(parentId) : null,
    };
    try {
      if (editingId === null) {
        await apiClient.post(apiUrl('categories'), payload);
        setNotice({ type: 'ok', text: 'Đã tạo danh mục' });
      } else {
        await apiClient.put(apiUrl(`categories/${editingId}`), payload);
        setNotice({ type: 'ok', text: 'Đã cập nhật danh mục' });
      }
      resetForm();
      await load();
    } catch (err) {
      setNotice({ type: 'error', text: errorMessage(err) });
    }
  };

  const remove = async (item: Category) => {
    if (!confirm(`Xóa danh mục "${item.name}"?`)) return;
    try {
      await apiClient.delete(apiUrl(`categories/${item.id}`));
      setNotice({ type: 'ok', text: 'Đã xóa danh mục' });
      if (editingId === item.id) resetForm();
      await load();
    } catch (err) {
      setNotice({ type: 'error', text: errorMessage(err) });
    }
  };

  const parentName = (id: number | null) =>
    id ? items.find((c) => c.id === id)?.name || `#${id}` : '—';

  return (
    <div>
      <NoticeBar notice={notice} />
      <form onSubmit={submit} className="bg-white border rounded-xl p-5 mb-6 grid gap-3 md:grid-cols-2">
        <h2 className="md:col-span-2 font-bold text-gray-900">
          {editingId === null ? 'Thêm danh mục' : `Sửa danh mục #${editingId}`}
        </h2>
        <input
          className={inputClass}
          placeholder="Tên danh mục"
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            if (!slugTouched) setSlug(slugify(e.target.value));
          }}
          required
        />
        <input
          className={inputClass}
          placeholder="Slug"
          value={slug}
          onChange={(e) => {
            setSlug(e.target.value);
            setSlugTouched(true);
          }}
          required
        />
        <input
          className={inputClass}
          placeholder="Mô tả (không bắt buộc)"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <select className={inputClass} value={parentId} onChange={(e) => setParentId(e.target.value)}>
          <option value="">Không có danh mục cha</option>
          {items
            .filter((c) => c.id !== editingId)
            .map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
        </select>
        <div className="md:col-span-2 flex gap-2">
          <button type="submit" className={primaryButton}>
            {editingId === null ? 'Thêm' : 'Lưu thay đổi'}
          </button>
          {editingId !== null && (
            <button type="button" onClick={resetForm} className={secondaryButton}>
              Hủy sửa
            </button>
          )}
        </div>
      </form>

      <div className="bg-white border rounded-xl overflow-hidden">
        <table className="w-full text-left text-sm">
          <thead className="bg-gray-50 text-xs uppercase text-gray-500">
            <tr>
              <th className="py-3 px-4">ID</th>
              <th className="py-3 px-4">Tên</th>
              <th className="py-3 px-4">Slug</th>
              <th className="py-3 px-4">Danh mục cha</th>
              <th className="py-3 px-4 text-right">Hành động</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {loading && (
              <tr>
                <td colSpan={5} className="py-6 text-center text-gray-400">
                  Đang tải...
                </td>
              </tr>
            )}
            {!loading && items.length === 0 && (
              <tr>
                <td colSpan={5} className="py-6 text-center text-gray-400">
                  Chưa có danh mục nào.
                </td>
              </tr>
            )}
            {items.map((item) => (
              <tr key={item.id}>
                <td className="py-3 px-4 font-mono text-gray-500">#{item.id}</td>
                <td className="py-3 px-4 font-medium text-gray-900">{item.name}</td>
                <td className="py-3 px-4 text-gray-600">{item.slug}</td>
                <td className="py-3 px-4 text-gray-600">{parentName(item.parent_id)}</td>
                <td className="py-3 px-4 text-right space-x-2">
                  <button onClick={() => startEdit(item)} className={secondaryButton}>
                    Sửa
                  </button>
                  <button onClick={() => remove(item)} className={dangerButton}>
                    Xóa
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function TagsPanel() {
  const [items, setItems] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [notice, setNotice] = useState<Notice>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [slugTouched, setSlugTouched] = useState(false);

  const load = useCallback(async () => {
    try {
      const res = await apiClient.get(apiUrl('tags'));
      setItems(res.data.data || []);
    } catch (err) {
      setNotice({ type: 'error', text: errorMessage(err) });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // Tải dữ liệu khi vào trang; load() đặt trạng thái loading đồng bộ.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [load]);

  const resetForm = () => {
    setEditingId(null);
    setName('');
    setSlug('');
    setSlugTouched(false);
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingId === null) {
        await apiClient.post(apiUrl('tags'), { name, slug });
        setNotice({ type: 'ok', text: 'Đã tạo thẻ' });
      } else {
        await apiClient.put(apiUrl(`tags/${editingId}`), { name, slug });
        setNotice({ type: 'ok', text: 'Đã cập nhật thẻ' });
      }
      resetForm();
      await load();
    } catch (err) {
      setNotice({ type: 'error', text: errorMessage(err) });
    }
  };

  const remove = async (item: Tag) => {
    if (!confirm(`Xóa thẻ "${item.name}"? Thẻ sẽ được gỡ khỏi các bài viết đang dùng.`)) return;
    try {
      await apiClient.delete(apiUrl(`tags/${item.id}`));
      setNotice({ type: 'ok', text: 'Đã xóa thẻ' });
      if (editingId === item.id) resetForm();
      await load();
    } catch (err) {
      setNotice({ type: 'error', text: errorMessage(err) });
    }
  };

  return (
    <div>
      <NoticeBar notice={notice} />
      <form onSubmit={submit} className="bg-white border rounded-xl p-5 mb-6 grid gap-3 md:grid-cols-2">
        <h2 className="md:col-span-2 font-bold text-gray-900">
          {editingId === null ? 'Thêm thẻ' : `Sửa thẻ #${editingId}`}
        </h2>
        <input
          className={inputClass}
          placeholder="Tên thẻ"
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            if (!slugTouched) setSlug(slugify(e.target.value));
          }}
          required
        />
        <input
          className={inputClass}
          placeholder="Slug"
          value={slug}
          onChange={(e) => {
            setSlug(e.target.value);
            setSlugTouched(true);
          }}
          required
        />
        <div className="md:col-span-2 flex gap-2">
          <button type="submit" className={primaryButton}>
            {editingId === null ? 'Thêm' : 'Lưu thay đổi'}
          </button>
          {editingId !== null && (
            <button type="button" onClick={resetForm} className={secondaryButton}>
              Hủy sửa
            </button>
          )}
        </div>
      </form>

      <div className="bg-white border rounded-xl p-5">
        {loading && <p className="text-gray-400 text-sm">Đang tải...</p>}
        {!loading && items.length === 0 && <p className="text-gray-400 text-sm">Chưa có thẻ nào.</p>}
        <div className="flex flex-wrap gap-3">
          {items.map((item) => (
            <div key={item.id} className="flex items-center gap-2 bg-gray-50 border rounded-full pl-4 pr-2 py-1.5">
              <span className="text-sm font-medium text-gray-800">{item.name}</span>
              <span className="text-xs text-gray-400">{item.slug}</span>
              <button
                onClick={() => {
                  setEditingId(item.id);
                  setName(item.name);
                  setSlug(item.slug);
                  setSlugTouched(true);
                }}
                className={secondaryButton}
              >
                Sửa
              </button>
              <button onClick={() => remove(item)} className={dangerButton}>
                Xóa
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function MenusPanel() {
  const [menus, setMenus] = useState<Menu[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [notice, setNotice] = useState<Notice>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [name, setName] = useState('');
  const [location, setLocation] = useState('');
  const [drafts, setDrafts] = useState<MenuItemDraft[]>([]);

  const load = useCallback(async () => {
    try {
      const [menuRes, catRes] = await Promise.all([
        apiClient.get(apiUrl('menus')),
        apiClient.get(apiUrl('categories')),
      ]);
      setMenus(menuRes.data.data || []);
      setCategories(catRes.data.data || []);
    } catch (err) {
      setNotice({ type: 'error', text: errorMessage(err) });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // Tải dữ liệu khi vào trang; load() đặt trạng thái loading đồng bộ.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [load]);

  const resetForm = () => {
    setEditingId(null);
    setName('');
    setLocation('');
    setDrafts([]);
  };

  const startEdit = (menu: Menu) => {
    setEditingId(menu.id);
    setName(menu.name);
    setLocation(menu.location || '');
    setDrafts(
      (menu.menu_items || []).map((item) => ({
        title: item.title,
        type: item.type === 'category' ? 'category' : 'custom',
        target_id: item.target_id,
        url: item.url,
      }))
    );
  };

  const updateDraft = (index: number, changes: Partial<MenuItemDraft>) => {
    setDrafts((prev) => prev.map((item, i) => (i === index ? { ...item, ...changes } : item)));
  };

  const moveDraft = (index: number, delta: number) => {
    setDrafts((prev) => {
      const target = index + delta;
      if (target < 0 || target >= prev.length) return prev;
      const next = [...prev];
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      name,
      location,
      menu_items: drafts.map((item, index) => ({
        title: item.title,
        type: item.type,
        target_id: item.type === 'category' ? item.target_id : 0,
        url: item.type === 'custom' ? item.url : '',
        sort_order: index + 1,
      })),
    };
    try {
      if (editingId === null) {
        await apiClient.post(apiUrl('menus'), payload);
        setNotice({ type: 'ok', text: 'Đã tạo menu' });
      } else {
        await apiClient.put(apiUrl(`menus/${editingId}`), payload);
        setNotice({ type: 'ok', text: 'Đã cập nhật menu' });
      }
      resetForm();
      await load();
    } catch (err) {
      setNotice({ type: 'error', text: errorMessage(err) });
    }
  };

  const remove = async (menu: Menu) => {
    if (!confirm(`Xóa menu "${menu.name}"?`)) return;
    try {
      await apiClient.delete(apiUrl(`menus/${menu.id}`));
      setNotice({ type: 'ok', text: 'Đã xóa menu' });
      if (editingId === menu.id) resetForm();
      await load();
    } catch (err) {
      setNotice({ type: 'error', text: errorMessage(err) });
    }
  };

  return (
    <div>
      <NoticeBar notice={notice} />
      <form onSubmit={submit} className="bg-white border rounded-xl p-5 mb-6 grid gap-3">
        <h2 className="font-bold text-gray-900">
          {editingId === null ? 'Thêm menu' : `Sửa menu #${editingId}`}
        </h2>
        <div className="grid gap-3 md:grid-cols-2">
          <input
            className={inputClass}
            placeholder="Tên menu"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <input
            className={inputClass}
            placeholder="Vị trí (ví dụ: header, footer)"
            value={location}
            onChange={(e) => setLocation(e.target.value)}
          />
        </div>

        <div className="space-y-2">
          {drafts.map((item, index) => (
            <div key={index} className="grid gap-2 md:grid-cols-[1fr_140px_1fr_auto] items-center">
              <input
                className={inputClass}
                placeholder="Tiêu đề mục"
                value={item.title}
                onChange={(e) => updateDraft(index, { title: e.target.value })}
                required
              />
              <select
                className={inputClass}
                value={item.type}
                onChange={(e) => updateDraft(index, { type: e.target.value as 'custom' | 'category' })}
              >
                <option value="custom">Liên kết</option>
                <option value="category">Danh mục</option>
              </select>
              {item.type === 'custom' ? (
                <input
                  className={inputClass}
                  placeholder="/duong-dan hoặc https://..."
                  value={item.url}
                  onChange={(e) => updateDraft(index, { url: e.target.value })}
                />
              ) : (
                <select
                  className={inputClass}
                  value={item.target_id || ''}
                  onChange={(e) => {
                    const target = Number(e.target.value);
                    const category = categories.find((c) => c.id === target);
                    updateDraft(index, {
                      target_id: target,
                      title: item.title || category?.name || '',
                    });
                  }}
                >
                  <option value="">Chọn danh mục</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              )}
              <div className="flex gap-1">
                <button type="button" onClick={() => moveDraft(index, -1)} className={secondaryButton}>
                  ↑
                </button>
                <button type="button" onClick={() => moveDraft(index, 1)} className={secondaryButton}>
                  ↓
                </button>
                <button
                  type="button"
                  onClick={() => setDrafts((prev) => prev.filter((_, i) => i !== index))}
                  className={dangerButton}
                >
                  Bỏ
                </button>
              </div>
            </div>
          ))}
          <button
            type="button"
            onClick={() =>
              setDrafts((prev) => [...prev, { title: '', type: 'custom', target_id: 0, url: '' }])
            }
            className={secondaryButton}
          >
            + Thêm mục
          </button>
        </div>

        <div className="flex gap-2">
          <button type="submit" className={primaryButton}>
            {editingId === null ? 'Tạo menu' : 'Lưu thay đổi'}
          </button>
          {editingId !== null && (
            <button type="button" onClick={resetForm} className={secondaryButton}>
              Hủy sửa
            </button>
          )}
        </div>
      </form>

      <div className="space-y-3">
        {loading && <p className="text-gray-400 text-sm">Đang tải...</p>}
        {!loading && menus.length === 0 && <p className="text-gray-400 text-sm">Chưa có menu nào.</p>}
        {menus.map((menu) => (
          <div key={menu.id} className="bg-white border rounded-xl p-5">
            <div className="flex items-center justify-between mb-2">
              <div>
                <span className="font-bold text-gray-900">{menu.name}</span>
                <span className="ml-2 text-xs text-gray-400">{menu.location || 'chưa đặt vị trí'}</span>
              </div>
              <div className="space-x-2">
                <button onClick={() => startEdit(menu)} className={secondaryButton}>
                  Sửa
                </button>
                <button onClick={() => remove(menu)} className={dangerButton}>
                  Xóa
                </button>
              </div>
            </div>
            <ol className="list-decimal list-inside text-sm text-gray-600 space-y-0.5">
              {(menu.menu_items || []).map((item, index) => (
                <li key={index}>
                  {item.title} <span className="text-gray-400">→ {item.url || '(chưa có liên kết)'}</span>
                </li>
              ))}
            </ol>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function CatalogAdminPage() {
  const router = useRouter();
  const [role, setRole] = useState<string | null>(null);
  const [tab, setTab] = useState<Tab>('categories');

  useEffect(() => {
    const { signedIn, role: storedRole } = readSession();
    if (!signedIn || (storedRole !== 'admin' && storedRole !== 'editor')) {
      alert('Bạn không có quyền truy cập trang này!');
      router.push('/posts');
      return;
    }
    // Chỉ chạy một lần sau khi mount: đọc phiên từ localStorage rồi chặn/chuyển hướng.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setRole(storedRole);
  }, [router]);

  if (!role) return null;

  const tabs: { key: Tab; label: string }[] = [
    { key: 'categories', label: 'Danh mục' },
    { key: 'tags', label: 'Thẻ' },
  ];
  if (role === 'admin') tabs.push({ key: 'menus', label: 'Menu' });

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-5xl mx-auto">
        <div className="flex items-center justify-between mb-6 bg-white p-6 rounded-xl shadow-sm border">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">🗂️ Danh mục, thẻ và menu</h1>
            <p className="text-sm text-gray-500 mt-1">Quản lý cách nội dung được phân loại và điều hướng.</p>
          </div>
          <button
            onClick={() => router.push('/posts')}
            className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 text-sm font-semibold rounded-lg transition cursor-pointer"
          >
            ← Quay lại Bài viết
          </button>
        </div>

        <div className="flex gap-2 mb-6">
          {tabs.map((item) => (
            <button
              key={item.key}
              onClick={() => setTab(item.key)}
              className={`px-4 py-2 rounded-lg text-sm font-semibold transition cursor-pointer ${
                tab === item.key
                  ? 'bg-blue-600 text-white'
                  : 'bg-white text-gray-700 border hover:bg-gray-100'
              }`}
            >
              {item.label}
            </button>
          ))}
        </div>

        {tab === 'categories' && <CategoriesPanel />}
        {tab === 'tags' && <TagsPanel />}
        {tab === 'menus' && role === 'admin' && <MenusPanel />}
      </div>
    </div>
  );
}