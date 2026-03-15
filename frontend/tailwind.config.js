/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        base: {
          DEFAULT: '#22272e',
          surface: '#2d333b',
          elevated: '#373e47',
        },
        border: '#373e47',
        text: {
          DEFAULT: '#adbac7',
          muted: '#768390',
        },
        accent: {
          DEFAULT: '#6cb6ff',
          muted: '#539bf5',
        },
        success: '#57ab5a',
        warning: '#c69026',
        error: '#f47067',
      },
      animation: {
        'fade-in': 'fadeIn 150ms ease-out forwards',
        'fade-out': 'fadeOut 100ms ease-in forwards',
        'slide-in-right': 'slideInRight 200ms ease-out forwards',
        'slide-out-right': 'slideOutRight 150ms ease-in forwards',
        'pulse-soft': 'pulseSoft 2s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0', transform: 'translateY(-4px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        fadeOut: {
          '0%': { opacity: '1', transform: 'translateY(0)' },
          '100%': { opacity: '0', transform: 'translateY(-4px)' },
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' },
        },
        slideOutRight: {
          '0%': { opacity: '1', transform: 'translateX(0)' },
          '100%': { opacity: '0', transform: 'translateX(20px)' },
        },
        pulseSoft: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.5' },
        },
      },
      transitionDuration: {
        '150': '150ms',
      },
    },
  },
  plugins: [
    function({ addUtilities }) {
      addUtilities({
        '.active-scale': {
          'transition': 'transform 100ms ease',
          '&:active': {
            'transform': 'scale(0.98)',
          },
        },
        '.transition-height': {
          'transition-property': 'height',
          'transition-timing-function': 'ease-in-out',
          'transition-duration': '200ms',
        },
      })
    },
  ],
}
