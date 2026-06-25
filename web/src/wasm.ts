export async function initWasm(): Promise<void> {
  // @ts-ignore
  const go = new Go(); // from wasm_exec.js

  const wasmResponse = await fetch("/main.wasm");
  const wasmBytes = await wasmResponse.arrayBuffer();
  const { instance } = await WebAssembly.instantiate(wasmBytes, go.importObject);

  go.run(instance);
}

export function parseSmfToTimeline(bytes: Uint8Array): any[] {
  // parseSmfToTimeline is attached to global by Go
  // @ts-ignore
  const result = window.parseSmfToTimeline(bytes);
  if (typeof result === "string") {
    return JSON.parse(result);
  }
  if (result && result.error) {
    throw new Error(result.error);
  }
  return result;
}
