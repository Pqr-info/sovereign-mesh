import React from "react";

type Props = {
  coords?: { x1: number; x2: number; x3: number; x4: number; x5: number };
  size?: number;
};

export const Radar5D: React.FC<Props> = ({ coords, size = 200 }) => {
  if (!coords) return null;
  const values = [coords.x1, coords.x2, coords.x3, coords.x4, coords.x5];
  const max = 1023;

  const radius = size / 2;
  const center = size / 2;

  const points = values.map((v, i) => {
    const angle = (Math.PI * 2 * i) / 5 - Math.PI / 2;
    const r = (v / max) * radius;
    return [
      center + r * Math.cos(angle),
      center + r * Math.sin(angle),
    ].join(",");
  });

  const labels = ["Hydrophobicity", "Complexity", "Stability", "Epoch", "Phase"];

  return (
    <svg width={size + 40} height={size + 40} style={{ margin: "auto", display: "block" }}>
      <g transform={`translate(20, 20)`}>
        {/* axes */}
        {[0, 1, 2, 3, 4].map(i => {
          const angle = (Math.PI * 2 * i) / 5 - Math.PI / 2;
          const labelAngle = angle;
          return (
            <g key={i}>
              <line
                x1={center}
                y1={center}
                x2={center + radius * Math.cos(angle)}
                y2={center + radius * Math.sin(angle)}
                stroke="#555"
              />
              <text
                x={center + (radius + 15) * Math.cos(labelAngle)}
                y={center + (radius + 15) * Math.sin(labelAngle) + 4}
                fill="#888"
                fontSize={9}
                textAnchor="middle"
              >
                {labels[i]}
              </text>
            </g>
          );
        })}

        {/* polygon */}
        <polygon
          points={points.join(" ")}
          fill="rgba(80, 200, 255, 0.3)"
          stroke="#4ecdc4"
          strokeWidth={2}
        />
      </g>
    </svg>
  );
};
