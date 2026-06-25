export function fftReal(input: Float32Array): { re: Float32Array; im: Float32Array } {
  const n = input.length;
  const re = new Float32Array(n);
  const im = new Float32Array(n);
  re.set(input);

  // Cooley–Tukey radix-2 (assume n is power of 2)
  for (let i = 1, j = 0; i < n; i++) {
    let bit = n >> 1;
    for (; j & bit; bit >>= 1) j ^= bit;
    j ^= bit;
    if (i < j) {
      [re[i], re[j]] = [re[j], re[i]];
      [im[i], im[j]] = [im[j], im[i]];
    }
  }

  for (let len = 2; len <= n; len <<= 1) {
    const ang = (-2 * Math.PI) / len;
    const wlenRe = Math.cos(ang);
    const wlenIm = Math.sin(ang);
    for (let i = 0; i < n; i += len) {
      let wRe = 1, wIm = 0;
      for (let j = 0; j < len / 2; j++) {
        const uRe = re[i + j];
        const uIm = im[i + j];
        const vRe = re[i + j + len / 2] * wRe - im[i + j + len / 2] * wIm;
        const vIm = re[i + j + len / 2] * wIm + im[i + j + len / 2] * wRe;

        re[i + j] = uRe + vRe;
        im[i + j] = uIm + vIm;
        re[i + j + len / 2] = uRe - vRe;
        im[i + j + len / 2] = uIm - vIm;

        const nwRe = wRe * wlenRe - wIm * wlenIm;
        const nwIm = wRe * wlenIm + wIm * wlenRe;
        wRe = nwRe;
        wIm = nwIm;
      }
    }
  }

  return { re, im };
}

export function magnitudeSpectrum(re: Float32Array, im: Float32Array): Float32Array {
  const n = re.length;
  const mag = new Float32Array(n / 2);
  let max = 1e-6;
  for (let i = 0; i < mag.length; i++) {
    const m = Math.hypot(re[i], im[i]);
    mag[i] = m;
    if (m > max) max = m;
  }
  for (let i = 0; i < mag.length; i++) mag[i] /= max;
  return mag;
}
