export const config = {
  // Kernel server URL
  kernelUrl: process.env.KERNEL_URL || 'http://localhost:8080',

  // Active namespace (set NAMESPACE env for FinLedger or other ledger)
  namespace: process.env.NAMESPACE || 'ProductLedger:/Kesteron/FieldServe',

  // v0.2 pinned roots (node IDs from seed; used by narratives and legacy references)
  pinnedRoots: {
    label: 'Kesteron Product Ledger',
    products: {
      fieldServe: '', // Set after seeding ProductLedger:/Kesteron/FieldServe
      assetLink: '',  // Set after seeding ProductLedger:/Kesteron/AssetLink
    },
    intent: {
      so1: '',
      ao1: '',
      tt1: '',
    },
  },

  // Demo namespaces: one per ledger (nav is namespace-driven; kernel returns these)
  availableNamespaces: [
    { id: 'ProductLedger:/Kesteron/AssetLink', label: 'Kesteron AssetLink' },
    { id: 'ProductLedger:/Kesteron/FieldServe', label: 'Kesteron FieldServe' },
    { id: 'FinLedger:/Kesteron/Treasury', label: 'Kesteron Treasury' },
    { id: 'FinLedger:/Kesteron/StablecoinReserves', label: 'Kesteron Stablecoin Reserves' },
  ],

  // Web Dot server port
  port: parseInt(process.env.PORT || '3000', 10),
};
