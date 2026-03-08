// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  modules: [
    '@bg-dev/nuxt-naiveui',
    '@pinia/nuxt',
  ],

  css: [
    '~/assets/css/main.css',
  ],

  app: {
    head: {
      title: '1Drive - Your Cloud, Your Control',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: '1Drive - Secure file management powered by Backblaze B2 Cloud Storage' },
        { name: 'theme-color', content: '#0a0a0f' },
      ],
      link: [
        {
          rel: 'preconnect',
          href: 'https://fonts.googleapis.com',
        },
        {
          rel: 'preconnect',
          href: 'https://fonts.gstatic.com',
          crossorigin: '',
        },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap',
        },
      ],
    },
  },

  runtimeConfig: {
    public: {
      apiBase: 'http://localhost:8080',
    },
  },

  naiveui: {
    colorModePreference: 'dark',
    themeConfig: {},
  },
})
