module.exports = {
  mode: 'jit',
  content: {
    files: ['src/**/*.rs', 'index.html', 'src/**/*.html']
  },
  darkMode: 'media',
  theme: {
    extend: {}
  },
  variants: {
    extend: {}
  },
  plugins: [require('daisyui')]
}
