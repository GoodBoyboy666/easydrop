import { defineConfig } from 'vite'
import { devtools } from '@tanstack/devtools-vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import tsconfigPaths from 'vite-tsconfig-paths'
import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const config = defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          const normalizedId = id.replace(/\\/g, '/')

          if (!normalizedId.includes('node_modules/')) {
            return
          }

          if (
            normalizedId.includes('/react/') ||
            normalizedId.includes('/react-dom/') ||
            normalizedId.includes('/scheduler/')
          ) {
            return 'react-vendor'
          }

          if (normalizedId.includes('/@tanstack/')) {
            return 'tanstack-vendor'
          }

          if (
            normalizedId.includes('/@uiw/react-md-editor/') ||
            normalizedId.includes('/@uiw/react-markdown-preview/') ||
            normalizedId.includes('/@uiw/react-codemirror/') ||
            normalizedId.includes('/@codemirror/') ||
            normalizedId.includes('/prismjs/')
          ) {
            return 'editor-vendor'
          }

          if (
            normalizedId.includes('/react-markdown/') ||
            normalizedId.includes('/remark-') ||
            normalizedId.includes('/rehype-') ||
            normalizedId.includes('/hast-') ||
            normalizedId.includes('/hast-util-') ||
            normalizedId.includes('/mdast-') ||
            normalizedId.includes('/micromark') ||
            normalizedId.includes('/parse5/') ||
            normalizedId.includes('/entities/') ||
            normalizedId.includes('/unified/') ||
            normalizedId.includes('/property-information/') ||
            normalizedId.includes('/character-entities') ||
            normalizedId.includes('/decode-named-character-reference') ||
            normalizedId.includes('/vfile') ||
            normalizedId.includes('/space-separated-tokens/') ||
            normalizedId.includes('/comma-separated-tokens/') ||
            normalizedId.includes('/web-namespaces/') ||
            normalizedId.includes('/zwitch/') ||
            normalizedId.includes('/trough/')
          ) {
            return 'markdown-vendor'
          }
        },
      },
    },
  },
  plugins: [
    devtools(),
    tsconfigPaths({ projects: ['./tsconfig.json'] }),
    tailwindcss(),
    tanstackRouter({
      target: 'react',
      autoCodeSplitting: true,
    }),
    viteReact(),
  ],
})

export default config
