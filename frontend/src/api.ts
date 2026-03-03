const BASE_URL = import.meta.env.VITE_API_BASE_URL as string;
const TOKEN = import.meta.env.VITE_API_TOKEN as string;

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${TOKEN}`,
      ...(init?.headers ?? {}),
    },
  });

  if (!res.ok) {
    const txt = await res.text().catch(() => "");
    throw new Error(txt || `HTTP ${res.status}`);
  }

  const text = await res.text();
  return (text ? JSON.parse(text) : null) as T;
}

export const api = {
  listUrls: (params: { page: number; pageSize: number; sort?: string; dir?: string }) => {
    const qs = new URLSearchParams({
      page: String(params.page),
      page_size: String(params.pageSize),
      ...(params.sort ? { sort: params.sort } : {}),
      ...(params.dir ? { dir: params.dir } : {}),
    });
    return request<{ items: UrlItem[]; total: number }>(`/urls?${qs.toString()}`);
  },

  createUrls: (urls: string[]) =>
    request<{ ids?: number[] }>(`/urls`, {
      method: "POST",
      body: JSON.stringify({ urls }),
    }),

  startUrls: (ids: number[]) =>
    request<void>(`/urls/start`, { method: "POST", body: JSON.stringify({ ids }) }),

  stopUrls: (ids: number[]) =>
    request<void>(`/urls/stop`, { method: "POST", body: JSON.stringify({ ids }) }),
};


export type UrlItem = {
  id: number;
  url: string;
  status: "created" | "queued" | "running" | "done" | "failed" | "stopped" | "expired";

  html_version?: string;
  title?: string;

  links_count?: number;
  internal_links_count?: number;
  external_links_count?: number;
  inaccessible_links_count?: number;

  has_login_form?: boolean;

  h1_count?: number;
  h2_count?: number;
  h3_count?: number;
  h4_count?: number;
  h5_count?: number;
  h6_count?: number;

  created_at?: string;
  updated_at?: string;
};