const colors = require('tailwindcss/colors')

module.exports = {
  content: [
    './views/**/*.gohtml',
    './views/**/**/*.gohtml'
  ],
  theme: {
    extend: {
      colors: {
        mbrblue: '#82a7bf',
        mbrbrown: '#6e6a68',
        mbrdarkbrown: '#825758',
        mbrrosa: '#fcc8b3',
        mbrbad:'#ff9d9d',
        mbrok: '#ffc078',
        mbrgood: '#83cbfa',
        green: colors.emerald,
        yellow: colors.amber,
        purple: colors.violet,
      },
      fontFamily: {
        sanchez: ['"Sanchez"', 'serif']
      }
    },
  },
  plugins: [],
  future: {
    removeDeprecatedGapUtilities: true,
  },
}
