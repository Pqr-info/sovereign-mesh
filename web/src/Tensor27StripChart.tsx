import React, { useEffect, useRef, useState } from "react";

type Tensor27Sample = {
  t: number;
  pos: number;
  neg: number;
  zero: number;
};

interface Tensor27StripChartProps {
  getTensor27Summary: () => { pos: number; neg: number; zero: number } | null;
  durationMs?: number; // window length
}

export const Tensor27StripChart: React.FC<Tensor27StripChartProps> = ({
  getTensor27Summary,
  durationMs = 10000,
}) => {
  const [samples, setSamples] = useState<Tensor27Sample[]>([]);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    const loop = () => {
      const now = performance.now();
      const summary = getTensor27Summary();
      if (summary) {
        setSamples((prev) => {
          const next: Tensor27Sample[] = [
            ...prev,
            { t: now, pos: summary.pos, neg: summary.neg, zero: summary.zero },
          ];
          const cutoff = now - durationMs;
          return next.filter((s) => s.t >= cutoff);
        });
      }
      rafRef.current = requestAnimationFrame(loop);
    };
    rafRef.current = requestAnimationFrame(loop);
    return () => {
      if (rafRef.current !== null) cancelAnimationFrame(rafRef.current);
    };
  }, [getTensor27Summary, durationMs]);

  const width = 240;
  const height = 60;

  if (samples.length === 0) {
    return (
      <div style={{ width, height, fontSize: 10, color: "#888" }}>
        Tensor27: waiting for samples…
      </div>
    );
  }

  const tMin = samples[0].t;
  const tMax = samples[samples.length - 1].t || tMin + 1;
  const maxVal =
    samples.reduce(
      (m, s) => Math.max(m, Math.abs(s.pos), Math.abs(s.neg), Math.abs(s.zero)),
      1
    ) || 1;

  const xScale = (t: number) =>
    ((t - tMin) / Math.max(1, tMax - tMin)) * width;
  const yScale = (v: number) => height - ((v / maxVal) * height) / 2 - height / 4;

  const buildPath = (key: "pos" | "neg" | "zero") => {
    return samples
      .map((s, i) => {
        const x = xScale(s.t);
        const y = yScale(s[key]);
        return `${i === 0 ? "M" : "L"} ${x.toFixed(1)} ${y.toFixed(1)}`;
      })
      .join(" ");
  };

  return (
    <div
      style={{
        width,
        height: height + 16,
        display: "flex",
        flexDirection: "column",
      }}
    >
      <div style={{ fontSize: 10, color: "#ccc", marginBottom: 2 }}>
        Tensor27 Metabolic Strip
      </div>
      <svg width={width} height={height} style={{ background: "#050509", border: "1px solid #222" }}>
        {/* zero line */}
        <line
          x1={0}
          y1={height / 2}
          x2={width}
          y2={height / 2}
          stroke="#333"
          strokeWidth={0.5}
        />
        {/* pos: orange */}
        <path
          d={buildPath("pos")}
          stroke="#ffb347"
          strokeWidth={1}
          fill="none"
        />
        {/* neg: blue */}
        <path
          d={buildPath("neg")}
          stroke="#4aa3ff"
          strokeWidth={1}
          fill="none"
        />
        {/* zero: gray */}
        <path
          d={buildPath("zero")}
          stroke="#888888"
          strokeWidth={0.7}
          fill="none"
        />
      </svg>
    </div>
  );
};
