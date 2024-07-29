/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './internal/templates/pdf/*.templ'
  ],
  theme: {
    extend: {
      fontSize: {
        'md': ['1rem', '1.5rem'],
        '2xs': ['.6rem', '.8rem'],
        '3xs': ['.45rem', '.6rem'],
      },
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
    require("daisyui"),
  ],
}
