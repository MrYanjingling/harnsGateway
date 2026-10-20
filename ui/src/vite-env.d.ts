/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_USE_MOCK?: string
  readonly VITE_GATEWAY_TARGET?: string
  readonly VITE_DATA_TARGET?: string
  readonly VITE_GATEWAY_BASE?: string
  readonly VITE_DATA_BASE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
