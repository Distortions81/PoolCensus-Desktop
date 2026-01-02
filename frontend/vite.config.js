import {defineConfig} from 'vite';

export default defineConfig({
  build: {
    // Safari 12 (macOS 10.14) doesn't support optional chaining, etc.
    target: 'es2018',
  },
});

