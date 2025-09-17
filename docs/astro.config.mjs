// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeRapide from 'starlight-theme-rapide'

// https://astro.build/config
export default defineConfig({
	site: 'https://devbox.ar0.eu',
	integrations: [
		starlight({
			title: 'devbox',
			description: 'Isolated development environments using Docker containers',
			favicon: './src/assets/logo.svg',
			logo: {
				replacesTitle: true,
        light: './src/assets/logo-dark.png',
        dark: './src/assets/logo.png',
      },
			components: {
				Footer: './src/components/CustomFooter.astro',
			},
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/itzCozi/devbox' }
			],
			editLink: {
        baseUrl: 'https://github.com/itzcozi/devbox/edit/main/docs/',
      },
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'index' },
						{ label: 'Installation', slug: 'guides/install' },
						{ label: 'Quick Start', slug: 'guides/intro' },
					],
				},
				{
					label: 'Configuration',
					items: [
						{ label: 'Configuration Files', slug: 'reference/configuration' },
						{ label: 'Templates & Setup', slug: 'reference/templates' },
					],
				},
				{
					label: 'Maintenance',
					items: [
						{ label: 'Cleanup & Maintenance', slug: 'reference/cleanup-maintenance' },
						{ label: 'Troubleshooting', slug: 'reference/troubleshooting' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'CLI Commands', slug: 'reference/cli' },
					],
				},
			],
			plugins: [starlightThemeRapide()],
		}),
	],
});
