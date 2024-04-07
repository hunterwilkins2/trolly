/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./components/**/*.{html,templ}"],
    theme: {
        extend: {
            colors: {
                'logoYellow': '#fbc466',
                'logoDarkYellow': '#fbd490',
            },
        },
        fontFamily: {
            'pacifico': ['"Pacifico"', 'cursive'],
        },

    },
    plugins: [],
}