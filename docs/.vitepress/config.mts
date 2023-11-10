import {defineConfig} from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
    title: "Wireset",
    description: "a collection of useful wireset for a next project",
    themeConfig: {
        nav: [
            {text: 'Guide', link: '/guide/'},
            {text: 'Wiresets', link: '/wiresets/'}
        ],

        sidebar: [
            {
                text: 'Guide',
                items: [
                    {text: 'Getting started', link: '/guide/getting-started.md'},
                    {text: 'Normal app', link: '/guide/normal-app.md'},
                    {text: 'Shopify app', link: '/guide/shopify-app.md'},
                ]
            },
            {
                text: 'Wiresets',
                items: [
                    {text: 'Configuration', link: '/wiresets/configuration'},
                    {text: 'Asset Handling', link: '/wiresets/assets'},
                    {text: 'Markdown Extensions', link: '/wiresets/markdown'},
                    {text: 'Using Vue in Markdown', link: '/wiresets/using-vue'},
                    {text: 'Deploying', link: '/wiresets/deploy'}
                ]
            }
        ],

        socialLinks: [
            {icon: 'github', link: 'https://github.com/vuejs/vitepress'}
        ]
    }
})
