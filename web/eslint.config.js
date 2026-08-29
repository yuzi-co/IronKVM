import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import eslintConfigPrettier from 'eslint-config-prettier'

// The React Compiler rules that eslint-plugin-react-hooks 7 added to its
// recommended set. They report 47 real findings in this tree: effects that call
// setState synchronously, refs read during render, and props mutated in place.
// Each one is worth fixing, and none of them is a dependency bump. They are
// warnings here so that they stay visible without turning every `pnpm lint`
// into a failure, and the backlog is recorded in
// docs/plans/2026-08-29-external-component-audit.md.
//
// rules-of-hooks stays an error. It reports nothing today, and it is the rule
// whose violations break a render rather than slow one down.
const compilerRulesAsWarnings = {
    'react-hooks/set-state-in-effect': 'warn',
    'react-hooks/immutability': 'warn',
    'react-hooks/refs': 'warn',
    'react-hooks/purity': 'warn',
}

export default tseslint.config(
    { ignores: ['dist', 'public/mockServiceWorker.js'] },
    {
        extends: [js.configs.recommended, ...tseslint.configs.recommended],
        files: ['**/*.{ts,tsx}'],
        languageOptions: {
            ecmaVersion: 2020,
            globals: globals.browser,
        },
        plugins: {
            'react-hooks': reactHooks,
            'react-refresh': reactRefresh,
        },
        rules: {
            ...reactHooks.configs.recommended.rules,
            ...compilerRulesAsWarnings,
            '@typescript-eslint/no-explicit-any': 'off',
            'react-hooks/exhaustive-deps': 'warn',
            'react-refresh/only-export-components': [
                'warn',
                { allowConstantExport: true },
            ],
        },
    },
    eslintConfigPrettier
)
