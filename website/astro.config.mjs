import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://k3desktop.github.io',
  base: '/',
  integrations: [
    starlight({
      title: 'K3Desktop',
      description: 'Desktop GUI for managing k3d Kubernetes clusters on Windows, macOS, and Linux.',
      logo: {
        src: './public/favicon.svg',
      },
      customCss: ['./src/styles/custom.css'],
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/k3desktop/k3desktop' },
      ],
      defaultLocale: 'en',
      locales: {
        en: { label: 'English', lang: 'en' },
        it: { label: 'Italiano', lang: 'it' },
        es: { label: 'Español', lang: 'es' },
        fr: { label: 'Français', lang: 'fr' },
        de: { label: 'Deutsch', lang: 'de' },
      },
      sidebar: [
        {
          label: 'Get Started',
          translations: {
            it: 'Inizia qui',
            es: 'Primeros pasos',
            fr: 'Commencer',
            de: 'Erste Schritte',
          },
          items: [{ autogenerate: { directory: 'getting-started' } }],
        },
        {
          label: 'Features',
          translations: {
            it: 'Funzionalità',
            es: 'Funcionalidades',
            fr: 'Fonctionnalités',
            de: 'Funktionen',
          },
          items: [{ autogenerate: { directory: 'guides' } }],
        },
      ],
    }),
  ],
});
