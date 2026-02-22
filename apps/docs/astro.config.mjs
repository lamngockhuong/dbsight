import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://dbsight.khuong.dev',
  base: '/',
  integrations: [
    starlight({
      title: 'DBSight Docs',
      defaultLocale: 'en',
      locales: {
        en: { label: 'English' },
        vi: { label: 'Tiếng Việt', lang: 'vi-VN' },
      },
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/lamngockhuong/dbsight' },
      ],
      sidebar: [
        {
          label: 'Getting Started',
          translations: { 'vi-VN': 'Bắt đầu' },
          autogenerate: { directory: 'getting-started' },
        },
        {
          label: 'User Guide',
          translations: { 'vi-VN': 'Hướng dẫn sử dụng' },
          autogenerate: { directory: 'user-guide' },
        },
        {
          label: 'SQL Reference',
          translations: { 'vi-VN': 'Tham chiếu SQL' },
          collapsed: true,
          autogenerate: { directory: 'sql-reference' },
        },
        {
          label: 'Optimization Guide',
          translations: { 'vi-VN': 'Hướng dẫn tối ưu' },
          collapsed: true,
          autogenerate: { directory: 'optimization-guide' },
        },
        {
          label: 'Developer Guide',
          translations: { 'vi-VN': 'Hướng dẫn phát triển' },
          collapsed: true,
          autogenerate: { directory: 'developer-guide' },
        },
        {
          label: 'Deployment',
          translations: { 'vi-VN': 'Triển khai' },
          collapsed: true,
          autogenerate: { directory: 'deployment' },
        },
      ],
    }),
  ],
});
