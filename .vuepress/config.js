const fs = require('fs');
const path = require('path');

module.exports = {
  base: '/',
  lang: 'en-EN',
  title: 'tacoscript',
  description: 'Tacoscript library provides functionality for provisioning of remote servers and local machines running on any OS.',
  head: [
    [
      'link',
      {
        rel: 'icon',
        type: 'image/png',
        sizes: '16x16',
        href: `/favicon/favicon-16x16.png`,
      },
    ],
    [
      'link',
      {
        rel: 'icon',
        type: 'image/png',
        sizes: '32x32',
        href: `/favicon/favicon-32x32.png`,
      },
    ],
    ['link', { rel: 'manifest', href: '/favicon/site.webmanifest' }],
    ['meta', { name: 'application-name', content: 'docs' }],
    ['meta', { name: 'apple-mobile-web-app-title', content: 'docs' }],
    [
      'meta',
      { name: 'apple-mobile-web-app-status-bar-style', content: 'black' },
    ],
    [
      'link',
      { rel: 'apple-touch-icon', href: `/favicon/apple-touch-icon.png` },
    ],
    ['meta', { name: 'msapplication-TileColor', content: '#0473e7' }],
    ['meta', { name: 'theme-color', content: '#0473e7' }],
  ],
  editLink: false,

  plugins: [ ],

  // additional global constants
  define: { },

  themeConfig: {
    contributors: false,
    editLink: false,
    logo: 'logo/tacoscript-img-text.svg',
    lastUpdated: false,
    navbar: [
      {
        text: 'Documentation',
        link: '/docs/',
      },
      {
        text: 'Download',
        link: 'https://github.com/cloudradar-monitoring/tacoscript/releases',
      },
    ],
    repo: 'cloudradar-monitoring/tacoscript',
    repoLabel: 'Github-Repo',
    sidebar: {
      '/docs/': [
        {
          isGroup: true,
          text: 'Documentation',
          children: getSideBar('docs'),
        },
      ],
    },
  },
};

function getSideBar(folder) {
  const extension = [".md"];

  return fs
    .readdirSync(path.join(`${__dirname}/../${folder}`))
    .filter(
      (item) =>
        //item.toLowerCase() != "readme.md"  &&
        fs.statSync(path.join(`${__dirname}/../${folder}`, item)).isFile() &&
        extension.includes(path.extname(item))
    );
}
