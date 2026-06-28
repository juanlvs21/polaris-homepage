// Wrapper mínimo de fetch con timeout y manejo de errores consistente.

export async function getJSON(url, { timeout = 12000 } = {}) {
  const ctrl = new AbortController();
  const id = setTimeout(() => ctrl.abort(), timeout);
  try {
    const res = await fetch(url, { signal: ctrl.signal, headers: { Accept: "application/json" } });
    if (!res.ok) {
      throw new Error(`HTTP ${res.status} en ${url}`);
    }
    return await res.json();
  } finally {
    clearTimeout(id);
  }
}
