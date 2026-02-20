// @ts-check

const isProd = process.env.NODE_ENV === 'production';

const config = {
  title: 'Warchaeology',
  tagline: 'WARC tools',
  url: 'https://nationallibraryofnorway.github.io',
  baseUrl: process.env.DOCS_BASEURL || (isProd ? '/warchaeology/' : '/'),
  organizationName: 'nationallibraryofnorway',
  projectName: 'warchaeology',
  trailingSlash: false,
  onBrokenLinks: 'throw',
  favicon: 'logo_small.png',
  markdown: {
    hooks:
      {
        onBrokenMarkdownLinks: 'warn'
      }
  },
  presets: [
    [
      'classic',
      {
        docs: {
          path: '../docs',
          routeBasePath: '/',
          sidebarPath: require.resolve('./sidebars.js')
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css')
        }
      }
    ]
  ],

  themeConfig: {
    image: 'logo.png',
    navbar: {
      title: 'Warchaeology',
      logo: {
        alt: 'Warchaeology logo',
        src: 'logo_small.png'
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs'
        },
        {
          href: 'https://github.com/nationallibraryofnorway/warchaeology',
          label: 'GitHub',
          position: 'right'
        }
      ]
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Repository',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/nationallibraryofnorway/warchaeology'
            }
          ]
        }
      ]
    }
  }
};

module.exports = config;
