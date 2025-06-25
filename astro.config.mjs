import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';
import icon from 'astro-icon';
// https://astro.build/config
export default defineConfig({
	output: 'static',
	srcDir: './astro',
	cacheDir: './.cache',
	site: 'https://world-dns-resolver.cyberjake.xyz',
	integrations: [
		sitemap(),
		icon({
			include: {
				'line-md': ['*'],
			},
		}),
	],
});
