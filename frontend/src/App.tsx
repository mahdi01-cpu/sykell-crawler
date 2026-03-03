import { useEffect, useMemo, useState } from "react";
import { api } from "./api";
import type { UrlItem } from "./api";
import "./App.css";

type SortKey =
  | "created_at"
  | "page_title"
  | "internal_links_count"
  | "external_links_count"
  | "inaccessible_links_count"
  | "status";

export default function App() {
  const [items, setItems] = useState<UrlItem[]>([]);
  const [total, setTotal] = useState(0);

  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);

  const [sort, setSort] = useState<SortKey>("created_at");
  const [dir, setDir] = useState<"asc" | "desc">("desc");

  const [selected, setSelected] = useState<Record<number, boolean>>({});
  const selectedIDs = useMemo(
    () => Object.entries(selected).filter(([, v]) => v).map(([k]) => Number(k)),
    [selected]
  );

  const [newURL, setNewURL] = useState("");
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string>("");

  async function load() {
    setLoading(true);
    setErr("");
    try {
      const res = await api.listUrls({ page, pageSize, sort, dir });
      setItems(res.items);
      setTotal(res.total);
    } catch (e: any) {
      setErr(e?.message ?? String(e));
    } finally {
      setLoading(false);
    }
  }

  // initial + whenever query changes
  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, sort, dir]);

  // polling for "real-time-ish" progress
  useEffect(() => {
    const t = setInterval(() => {
      load();
    }, 2000);
    return () => clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, sort, dir]);

  async function onAdd() {
    const raw = newURL.trim();
    if (!raw) return;
    setErr("");
    try {
      await api.createUrls([raw]);
      setNewURL("");
      setPage(1);
      await load();
    } catch (e: any) {
      setErr(e?.message ?? String(e));
    }
  }

  async function onStart() {
    if (selectedIDs.length === 0) return;
    setErr("");
    try {
      await api.startUrls(selectedIDs);
      await load();
    } catch (e: any) {
      setErr(e?.message ?? String(e));
    }
  }

  async function onStop() {
    if (selectedIDs.length === 0) return;
    setErr("");
    try {
      await api.stopUrls(selectedIDs);
      await load();
    } catch (e: any) {
      setErr(e?.message ?? String(e));
    }
  }

  function toggleAll(checked: boolean) {
    const m: Record<number, boolean> = {};
    for (const it of items) m[it.id] = checked;
    setSelected(m);
  }

  function toggleOne(id: number, checked: boolean) {
    setSelected((prev) => ({ ...prev, [id]: checked }));
  }

  function changeSort(k: SortKey) {
    if (sort === k) setDir((d) => (d === "asc" ? "desc" : "asc"));
    else {
      setSort(k);
      setDir("desc");
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className="container">
      <h1>Sykell Crawler</h1>

      <div className="card">
        <h2>Add URL</h2>
        <div className="row">
          <input
            value={newURL}
            onChange={(e) => setNewURL(e.target.value)}
            placeholder="https://example.com"
          />
          <button onClick={onAdd}>Add</button>
        </div>

        <div className="row">
          <button disabled={selectedIDs.length === 0} onClick={onStart}>
            Start
          </button>
          <button disabled={selectedIDs.length === 0} onClick={onStop}>
            Stop
          </button>

          <div style={{ marginLeft: "auto", opacity: 0.7 }}>
            {loading ? "Loading..." : `Total: ${total}`}
          </div>
        </div>

        {err ? <div className="error">Error: {err}</div> : null}
      </div>

      <div className="card">
        <h2>Results</h2>

        <table>
          <thead>
            <tr>
              <th>
                <input
                  type="checkbox"
                  onChange={(e) => toggleAll(e.target.checked)}
                />
              </th>
              <th onClick={() => changeSort("status")} className="clickable">
                Status
              </th>
              <th>URL</th>
              <th onClick={() => changeSort("page_title")} className="clickable">
                Title
              </th>
              <th>
                HTML
              </th>
              <th onClick={() => changeSort("internal_links_count")} className="clickable">
                Internal
              </th>
              <th onClick={() => changeSort("external_links_count")} className="clickable">
                External
              </th>
              <th onClick={() => changeSort("inaccessible_links_count")} className="clickable">
                Broken
              </th>
              <th>Login</th>
              <th>H1/H2/H3</th>
            </tr>
          </thead>

          <tbody>
            {items.map((it) => (
              <tr key={it.id}>
                <td>
                  <input
                    type="checkbox"
                    checked={!!selected[it.id]}
                    onChange={(e) => toggleOne(it.id, e.target.checked)}
                  />
                </td>
                <td><span className={`status ${it.status}`}>{it.status}</span></td>
                <td className="mono">{it.url}</td>
                <td>{it.title ?? "-"}</td>
                <td>{it.html_version ?? "-"}</td>
                <td>{it.internal_links_count ?? 0}</td>
                <td>{it.external_links_count ?? 0}</td>
                <td>{it.inaccessible_links_count ?? 0}</td>
                <td>{it.has_login_form ? "yes" : "no"}</td>
                <td>
                  {(it.h1_count ?? 0)}/{(it.h2_count ?? 0)}/{(it.h3_count ?? 0)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        <div className="row" style={{ marginTop: 12 }}>
          <button disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Prev
          </button>
          <div style={{ padding: "0 12px" }}>
            Page {page} / {totalPages}
          </div>
          <button disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            Next
          </button>
        </div>
      </div>
    </div>
  );
}