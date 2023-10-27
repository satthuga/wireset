import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Wireset",
  description: "a collection of useful wireset for a next project",
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Guide', link: '/guide/' },
      { text: 'Wiresets', link: '/wiresets/' }
    ],

    sidebar: [
      {
        text: 'Guide',
        items: [
            { text: 'Installation', link: '/guide/installation.md' },
        ]
      },
      {
        text: 'Wiresets',
        items: [
            { text: 'Getting Started', link: '/wiresets/' },
            { text: 'Configuration', link: '/wiresets/configuration' },
            { text: 'Asset Handling', link: '/wiresets/assets' },
            { text: 'Markdown Extensions', link: '/wiresets/markdown' },
            { text: 'Using Vue in Markdown', link: '/wiresets/using-vue' },
            { text: 'Deploying', link: '/wiresets/deploy' }
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/vuejs/vitepress' }
    ]
  }
})
