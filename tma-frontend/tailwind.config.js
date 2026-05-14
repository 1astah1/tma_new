/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        tg: {
          bg: 'var(--tg-bg, #ffffff)',
          text: 'var(--tg-text, #000000)',
          button: 'var(--tg-button, #40a7e3)',
          'button-text': 'var(--tg-button-text, #ffffff)',
          hint: 'var(--tg-hint, #999999)',
          secondary: 'var(--tg-secondary, #f4f4f5)',
        },
      },
    },
  },
  plugins: [],
}
