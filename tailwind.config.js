module.exports = {
  purge: [
    './views/**/*.html',
    './views/**/**/*.html'
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
      },
      fontFamily: {
        sanchez: ['"Sanchez"', 'serif']
      }
    },
  },
  variants: {},
  plugins: [],
  future: {
    removeDeprecatedGapUtilities: true,
  },
}
