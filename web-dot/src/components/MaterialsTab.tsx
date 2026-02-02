import React, { useState, useEffect } from 'react';
import { guiClient } from '../adapters/GuiClient.js';

type MaterialsTabProps = {
  contextNodeId: string | null;
  contextNamespaceId: string | null;
};

export default function MaterialsTab({ contextNodeId, contextNamespaceId }: MaterialsTabProps) {
  const [categories, setCategories] = useState<Array<{
    category: string;
    items: Array<{
      material_id: string;
      node_id: string;
      title: string;
      content_ref: string;
      media_type: string;
      category: string;
      meta: Record<string, unknown>;
    }>;
  }>>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedMaterial, setSelectedMaterial] = useState<{
    material_id: string;
    node_id: string;
    title: string;
    content_ref: string;
    media_type: string;
  } | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);

  useEffect(() => {
    if (!contextNamespaceId) {
      setCategories([]);
      return;
    }
    loadMaterials();
  }, [contextNodeId, contextNamespaceId]);

  const loadMaterials = async () => {
    if (!contextNamespaceId) return;
    setLoading(true);
    setError(null);
    try {
      const result = await guiClient.materials(contextNamespaceId, contextNodeId ?? undefined);
      setCategories(result.categories);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load materials');
      console.error('Error loading materials:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleMaterialClick = (material: {
    material_id: string;
    node_id: string;
    title: string;
    content_ref: string;
    media_type: string;
  }) => {
    setSelectedMaterial(material);
    // Emit custom event for App to open markdown tool tab
    window.dispatchEvent(new CustomEvent('openMaterial', { detail: material }));
  };

  // Filter categories and items based on search and category filter
  const filteredCategories = React.useMemo(() => {
    if (!searchQuery && !selectedCategory) return categories;
    
    return categories
      .filter(cat => !selectedCategory || cat.category === selectedCategory)
      .map(cat => ({
        ...cat,
        items: cat.items.filter(item => {
          if (searchQuery) {
            const query = searchQuery.toLowerCase();
            const matchesTitle = item.title.toLowerCase().includes(query);
            const matchesType = item.media_type.toLowerCase().includes(query);
            const matchesCategory = item.category.toLowerCase().includes(query);
            if (!matchesTitle && !matchesType && !matchesCategory) return false;
          }
          return true;
        })
      }))
      .filter(cat => cat.items.length > 0);
  }, [categories, searchQuery, selectedCategory]);

  const allCategories = React.useMemo(() => {
    return Array.from(new Set(categories.map(cat => cat.category))).sort();
  }, [categories]);

  return (
    <div style={{ display: 'flex', height: '100%', gap: '1rem', flexDirection: 'column' }}>
      <div style={{ padding: '1rem', borderBottom: '1px solid #e0e0e0' }}>
        <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
          <input
            type="text"
            placeholder="Search by title, type, or category..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{
              flex: 1,
              padding: '0.5rem',
              border: '1px solid #e0e0e0',
              borderRadius: '4px',
              fontSize: '0.875rem',
            }}
          />
          <select
            value={selectedCategory || ''}
            onChange={(e) => setSelectedCategory(e.target.value || null)}
            style={{
              padding: '0.5rem',
              border: '1px solid #e0e0e0',
              borderRadius: '4px',
              fontSize: '0.875rem',
              minWidth: '150px',
            }}
          >
            <option value="">All Categories</option>
            {allCategories.map(cat => (
              <option key={cat} value={cat}>{cat}</option>
            ))}
          </select>
        </div>
      </div>
      <div style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
        {loading && <div className="loading">Loading materials...</div>}
        {error && <div className="error">Error: {error}</div>}
        {!loading && !error && filteredCategories.length === 0 && (
          <div style={{ padding: '1rem', color: '#555' }}>
            <p style={{ marginBottom: '0.5rem' }}>
              No materials found{searchQuery || selectedCategory ? ' matching filters' : ''}.
            </p>
            {contextNamespaceId && (
              <p style={{ fontSize: '0.875rem', marginTop: '0.5rem' }}>
                Materials are stored in the kernel and linked to nodes. To add materials for this namespace, run the materials seed script (e.g. <code style={{ fontSize: '0.8rem', background: '#f0f0f0', padding: '0.125rem 0.25rem' }}>seed-kesteron-materials.sh</code>) with <code style={{ fontSize: '0.8rem', background: '#f0f0f0', padding: '0.125rem 0.25rem' }}>NS=&quot;{contextNamespaceId}&quot;</code>.
              </p>
            )}
          </div>
        )}
        {!loading && !error && filteredCategories.length > 0 && (
          <div>
            {filteredCategories.map((cat) => (
              <div key={cat.category} style={{ marginBottom: '2rem' }}>
                <h3 style={{ marginBottom: '1rem', borderBottom: '2px solid #e0e0e0', paddingBottom: '0.5rem' }}>
                  {cat.category} ({cat.items.length})
                </h3>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {cat.items.map((item) => (
                    <div
                      key={item.material_id}
                      onClick={() => handleMaterialClick(item)}
                      style={{
                        padding: '0.75rem',
                        border: '1px solid #e0e0e0',
                        borderRadius: '4px',
                        cursor: 'pointer',
                        background: selectedMaterial?.material_id === item.material_id ? '#e3f2fd' : 'white',
                      }}
                    >
                      <div style={{ fontWeight: '500' }}>{item.title}</div>
                      <div style={{ fontSize: '0.875rem', color: '#666', marginTop: '0.25rem' }}>
                        {item.media_type} • {item.content_ref}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
