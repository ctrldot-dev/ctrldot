import React, { useState, useEffect, useRef } from 'react';
import { guiClient, type NamespaceOption } from '../adapters/GuiClient.js';

const RECENT_MAX = 5;
const RECENT_KEY = 'web-dot-recent-namespaces';

function getRecent(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as string[];
    return Array.isArray(parsed) ? parsed.slice(0, RECENT_MAX) : [];
  } catch {
    return [];
  }
}

function pushRecent(namespaceId: string) {
  const recent = getRecent().filter((id) => id !== namespaceId);
  recent.unshift(namespaceId);
  localStorage.setItem(RECENT_KEY, JSON.stringify(recent.slice(0, RECENT_MAX)));
}

type NamespaceSelectorProps = {
  onSelect: (ns: NamespaceOption) => void;
  openNamespaceIds: Set<string>;
};

export default function NamespaceSelector({ onSelect, openNamespaceIds }: NamespaceSelectorProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [data, setData] = useState<{ namespaces: NamespaceOption[]; grouped: Record<string, NamespaceOption[]> } | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (open && !data) {
      setLoading(true);
      setError(null);
      guiClient
        .namespaces()
        .then(setData)
        .catch((e) => setError(e instanceof Error ? e.message : 'Failed to load'))
        .finally(() => setLoading(false));
    }
  }, [open, data]);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const recentIds = getRecent();
  const recent = (data?.namespaces ?? []).filter((ns) => recentIds.includes(ns.id));
  const q = query.trim().toLowerCase();
  const filtered =
    q === ''
      ? data?.namespaces ?? []
      : (data?.namespaces ?? []).filter(
          (ns) =>
            ns.id.toLowerCase().includes(q) ||
            ns.label.toLowerCase().includes(q) ||
            (ns.prefix && ns.prefix.toLowerCase().includes(q))
        );
  const grouped = data?.grouped ?? {};
  const prefixOrder = Object.keys(grouped).sort((a, b) => a.localeCompare(b));

  const handleChoose = (ns: NamespaceOption) => {
    pushRecent(ns.id);
    onSelect(ns);
    setOpen(false);
    setQuery('');
  };

  return (
    <div ref={containerRef} style={{ position: 'relative', minWidth: 220 }}>
      <div
        role="combobox"
        aria-expanded={open}
        aria-haspopup="listbox"
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          padding: '0.35rem 0.75rem',
          background: '#f8f9fa',
          border: '1px solid #dee2e6',
          borderRadius: '6px',
          cursor: 'pointer',
          fontSize: '0.875rem',
        }}
        onClick={() => {
          setOpen((o) => !o);
          if (!open) inputRef.current?.focus();
        }}
      >
        <span style={{ color: '#6c757d', whiteSpace: 'nowrap' }}>Ledger / Namespace</span>
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => setOpen(true)}
          placeholder="Select…"
          style={{
            flex: 1,
            minWidth: 0,
            border: 'none',
            background: 'transparent',
            fontSize: 'inherit',
            outline: 'none',
          }}
          aria-autocomplete="list"
          aria-controls="namespace-listbox"
          aria-activedescendant={undefined}
        />
      </div>
      {open && (
        <div
          id="namespace-listbox"
          role="listbox"
          style={{
            position: 'absolute',
            top: '100%',
            left: 0,
            right: 0,
            marginTop: 4,
            maxHeight: 320,
            overflowY: 'auto',
            background: 'white',
            border: '1px solid #dee2e6',
            borderRadius: '8px',
            boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
            zIndex: 100,
          }}
        >
          {loading && (
            <div style={{ padding: '0.75rem 1rem', color: '#6c757d', fontSize: '0.875rem' }}>
              Loading…
            </div>
          )}
          {error && (
            <div style={{ padding: '0.75rem 1rem', color: '#dc3545', fontSize: '0.875rem' }}>
              {error}
            </div>
          )}
          {!loading && !error && (
            <>
              {recent.length > 0 && query === '' && (
                <div style={{ padding: '0.5rem 1rem 0 0', fontSize: '0.75rem', color: '#6c757d', fontWeight: 600 }}>
                  Recent
                </div>
              )}
              {recent.length > 0 &&
                query === '' &&
                recent.map((ns) => (
                  <div
                    key={ns.id}
                    role="option"
                    aria-selected={openNamespaceIds.has(ns.id)}
                    onClick={() => handleChoose(ns)}
                    style={{
                      padding: '0.5rem 1rem',
                      cursor: 'pointer',
                      fontSize: '0.875rem',
                      background: openNamespaceIds.has(ns.id) ? '#e7f1ff' : 'transparent',
                    }}
                  >
                    <div style={{ fontWeight: 500 }}>{ns.label}</div>
                    <div style={{ fontSize: '0.75rem', color: '#6c757d' }}>{ns.id}</div>
                  </div>
                ))}
              {query === ''
                ? prefixOrder.map((prefix) => {
                    const list = grouped[prefix] ?? [];
                    if (list.length === 0) return null;
                    return (
                      <div key={prefix}>
                        <div
                          style={{
                            padding: '0.5rem 1rem 0 0',
                            fontSize: '0.75rem',
                            color: '#6c757d',
                            fontWeight: 600,
                            textTransform: 'uppercase',
                          }}
                        >
                          {prefix}
                        </div>
                        {list.map((ns) => (
                          <div
                            key={ns.id}
                            role="option"
                            aria-selected={openNamespaceIds.has(ns.id)}
                            onClick={() => handleChoose(ns)}
                            style={{
                              padding: '0.5rem 1rem',
                              cursor: 'pointer',
                              fontSize: '0.875rem',
                              background: openNamespaceIds.has(ns.id) ? '#e7f1ff' : 'transparent',
                            }}
                          >
                            <div style={{ fontWeight: 500 }}>{ns.label}</div>
                            <div style={{ fontSize: '0.75rem', color: '#6c757d' }}>{ns.id}</div>
                          </div>
                        ))}
                      </div>
                    );
                  })
                : filtered.map((ns) => (
                    <div
                      key={ns.id}
                      role="option"
                      aria-selected={openNamespaceIds.has(ns.id)}
                      onClick={() => handleChoose(ns)}
                      style={{
                        padding: '0.5rem 1rem',
                        cursor: 'pointer',
                        fontSize: '0.875rem',
                        background: openNamespaceIds.has(ns.id) ? '#e7f1ff' : 'transparent',
                      }}
                    >
                      <div style={{ fontWeight: 500 }}>{ns.label}</div>
                      <div style={{ fontSize: '0.75rem', color: '#6c757d' }}>{ns.id}</div>
                    </div>
                  ))}
              {query !== '' && filtered.length === 0 && (
                <div style={{ padding: '0.75rem 1rem', color: '#6c757d', fontSize: '0.875rem' }}>
                  No matching namespace
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}
