import React, { useEffect, useRef } from "react";

type Props = {
  versions: number[];   // evolutionary_version history
  color: string;
  size?: number;
  helixPointRef?: React.MutableRefObject<{x: number, y: number, z: number} | null>;
  manifoldPointRef?: React.MutableRefObject<{x: number, y: number, z: number} | null>;
};

export const HelixEvolution: React.FC<Props> = ({ 
  versions, 
  color, 
  size = 300,
  helixPointRef,
  manifoldPointRef
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const thetaRef = useRef(0);

  const project = (v: number) => {
    const phase = v * 0.35;
    const radius = 20;
    const height = v * 4;

    const x = radius * Math.cos(phase);
    const y = radius * Math.sin(phase);
    const z = height;

    return { x, y, z };
  };

  useEffect(() => {
    const canvas = canvasRef.current!;
    const ctx = canvas.getContext("2d")!;
    const center = size / 2;

    let animationId: number;

    const animate = () => {
      thetaRef.current += 0.01;
      const theta = thetaRef.current;

      ctx.clearRect(0, 0, size, size);

      versions.forEach((v, i) => {
        const p3 = project(v);
        
        const coupling = 0.15;
        if (manifoldPointRef?.current && i === versions.length - 1) {
          p3.x = p3.x * (1 - coupling) + manifoldPointRef.current.x * coupling;
          p3.y = p3.y * (1 - coupling) + manifoldPointRef.current.y * coupling;
          p3.z = p3.z * (1 - coupling) + manifoldPointRef.current.z * coupling;
        }

        if (helixPointRef && i === versions.length - 1) {
          helixPointRef.current = p3;
        }

        const cosT = Math.cos(theta);
        const sinT = Math.sin(theta);
        const p = {
          x: p3.x * cosT - p3.z * sinT,
          y: p3.y
        };

        const alpha = (i + 1) / versions.length;

        ctx.fillStyle = `rgba(${hexToRgb(color)}, ${alpha})`;
        ctx.beginPath();
        // Adjust center Y based on the maximum height so the helix grows upward but stays in view
        // Alternatively, just draw it normally
        const maxV = Math.max(...versions, 10);
        const yOffset = center + (maxV * 2) - p.y - v*4; // approximate centering
        
        ctx.arc(center + p.x, center + p.y + (maxV*2 - v*4), 4, 0, Math.PI * 2);
        ctx.fill();
        
        if (i > 0) {
            const prevV = versions[i-1];
            const prevP3 = project(prevV);
            const prevP = {
                x: prevP3.x * cosT - prevP3.z * sinT,
                y: prevP3.y
            };
            ctx.strokeStyle = `rgba(${hexToRgb(color)}, ${alpha * 0.5})`;
            ctx.lineWidth = 2;
            ctx.beginPath();
            ctx.moveTo(center + prevP.x, center + prevP.y + (maxV*2 - prevV*4));
            ctx.lineTo(center + p.x, center + p.y + (maxV*2 - v*4));
            ctx.stroke();
        }
      });

      animationId = requestAnimationFrame(animate);
    };

    animate();
    
    return () => cancelAnimationFrame(animationId);
  }, [versions, color, size]);

  return <canvas ref={canvasRef} width={size} height={size} style={{ display: "block", margin: "auto" }} />;
};

function hexToRgb(hex: string): string {
  if (!hex) return "255,255,255";
  const h = hex.replace("#", "");
  const bigint = parseInt(h, 16);
  const r = (bigint >> 16) & 255;
  const g = (bigint >> 8) & 255;
  const b = bigint & 255;
  return `${r}, ${g}, ${b}`;
}
