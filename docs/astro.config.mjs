
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeRapide from 'starlight-theme-rapide'


export default defineConfig({
	site: 'https://devbox.ar0.eu',
	integrations: [
		starlight({
			title: 'devbox',
			description: 'Isolated development environments using Docker containers',
			favicon: '/favicon.svg',
			logo: {
				replacesTitle: true,
        light: './src/assets/logo-dark.png',
        dark: './src/assets/logo.png',
      },
      lastUpdated: true,
			components: {
				Footer: './src/components/CustomFooter.astro',
			},
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/itzcozi/devbox' },
				{ icon: 'telegram', label: 'Telegram', href: 'http://t.me/devboxcli' }
			],
			editLink: {
        baseUrl: 'https://github.com/itzcozi/devbox/edit/main/docs/',
      },
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'docs/intro' },
						{ label: 'Quick Start', slug: 'docs/start' },
						{ label: 'Installation', slug: 'docs/install' },
					],
				},
				{
					label: 'Configuration',
					collapsed: true,
					items: [
						{ label: 'Configuration Files', slug: 'docs/configuration' },
						{ label: 'Templates & Setup', slug: 'docs/templates' },
					],
				},
				{
					label: 'Maintenance',
					collapsed: true,
					items: [
						{ label: 'Cleanup & Maintenance', slug: 'docs/cleanup-maintenance' },
						{ label: 'Troubleshooting', slug: 'docs/troubleshooting' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'CLI Commands', slug: 'docs/cli' },
					],
				},
			],
			plugins: [starlightThemeRapide()],
		}),
	],
});
