// Polyfill localStorage for Node.js v25+ (needed by @typescript/vfs during SSR)
// Node.js v25 adds experimental localStorage but with broken implementation
// that @typescript/vfs doesn't handle correctly.
// https://github.com/shikijs/twoslash/issues/191

if (typeof globalThis.localStorage === "undefined" || typeof globalThis.localStorage.getItem !== "function") {
  const storage = new Map();
  globalThis.localStorage = {
    getItem: (key) => storage.get(key) ?? null,
    setItem: (key, value) => storage.set(key, String(value)),
    removeItem: (key) => storage.delete(key),
    clear: () => storage.clear(),
    get length() {
      return storage.size;
    },
    key: (index) => [...storage.keys()][index] ?? null,
  };
}
