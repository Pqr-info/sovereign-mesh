import React from "react";

type Props = {
  width: number;
  height?: number;
  density: number[];
  viewStart: number;
  viewEnd: number;
  onJump: (tick: number) => void;
};

export const Minimap: React.FC<Props> = ({
  width,
  height = 60,
  density,
  viewStart,
  viewEnd,
  onJump,
}) => {
  const numBins = density.length;

  const xScale = (i: number) => (i / numBins) * width;
  const viewX1 = Math.max(0, viewStart * width);
  const viewX2 = Math.min(width, viewEnd * width);
  const viewWidth = Math.max(0, viewX2 - viewX1);

  return (
    <svg
      width={width}
      height={height}
      style={{ border: "1px solid #333", cursor: "pointer", backgroundColor: "#0a0a0a", borderRadius: 8, marginTop: 8 }}
      onClick={(e) => {
        const rect = e.currentTarget.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const tick = x / width;
        onJump(tick);
      }}
    >
      {/* density bars */}
      {density.map((d, i) => (
        <rect
          key={i}
          x={xScale(i)}
          y={height - d * height}
          width={width / numBins}
          height={d * height}
          fill="rgba(255, 200, 0, 0.3)"
        />
      ))}

      {/* viewport */}
      <rect
        x={viewX1}
        y={0}
        width={viewWidth}
        height={height}
        fill="rgba(80, 200, 255, 0.25)"
        stroke="#4ecdc4"
        strokeWidth={2}
      />
    </svg>
  );
};
