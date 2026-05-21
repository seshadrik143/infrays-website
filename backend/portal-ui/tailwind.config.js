/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: {
          900: "#0b0f17",
          800: "#111727",
          700: "#1a2236",
          600: "#283149",
        },
        accent: {
          500: "#7c5cff",
          400: "#9d83ff",
        },
      },
    },
  },
  plugins: [],
};
