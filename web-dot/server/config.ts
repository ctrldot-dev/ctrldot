export const config = {
  // Kernel server URL
  kernelUrl: process.env.KERNEL_URL || 'http://localhost:8080',
  
  // Fixed namespace for this demo
  namespace: 'ProductLedger:/Kesteron',

  // v0.2 pinned roots (from context/materials-product-tree-product-ledger-demo-ui.yaml)
  pinnedRoots: {
    label: 'Kesteron Product Ledger',
    products: {
      fieldServe: 'node:306a477f-de8d-4049-b36f-7b3007866381',
      assetLink: 'node:bcabd26c-1276-4db7-9092-5a238d14a57a',
    },
    intent: {
      so1: 'node:b985f1da-70a9-4e4b-9093-a15bb45cfc78',
      ao1: 'node:b1aa950e-492f-4fab-976a-c744cfee54d7',
      tt1: 'node:27f9d326-fcb7-4924-abf7-a48ae329af6b',
    },
  },
  
  // Web Dot server port
  port: parseInt(process.env.PORT || '3000', 10),
};
