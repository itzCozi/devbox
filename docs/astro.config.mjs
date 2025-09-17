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
			favicon: '/favicon.svg',
			logo: {
				replacesTitle: true,
        light: './src/assets/logo-dark.png',
        dark: './src/assets/logo.png',
      },
			components: {
				Footer: './src/components/CustomFooter.astro',
			},
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/itzCozi/devbox' },
				{ icon: 'discord', label: 'Discord', href: '/discord' }
			],
			editLink: {
        baseUrl: 'https://github.com/itzcozi/devbox/edit/main/docs/',
      },
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'intro' },
						{ label: 'Quick Start', slug: 'start' },
						{ label: 'Installation', slug: 'install' },
					],
				},
				{
					label: 'Configuration',
					items: [
						{ label: 'Configuration Files', slug: 'configuration' },
						{ label: 'Templates & Setup', slug: 'templates' },
					],
				},
				{
					label: 'Maintenance',
					items: [
						{ label: 'Cleanup & Maintenance', slug: 'cleanup-maintenance' },
						{ label: 'Troubleshooting', slug: 'troubleshooting' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'CLI Commands', slug: 'cli' },
					],
				},
			],
			plugins: [starlightThemeRapide()],
		}),
	],
});
