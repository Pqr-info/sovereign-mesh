import React, { useEffect, useRef, useState } from "react";

type Coords5D = { x1: number; x2: number; x3: number; x4: number; x5: number };

type Props = {
  coords: Coords5D;
  history: Coords5D[];
  size?: number;
};

export const Manifold5D: React.FC<Props> = ({ coords, history, size = 300 }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const thetaRef = useRef(0);
  const [hoveredNode, setHoveredNode] = useState<Coords5D | null>(null);
  const [hoverPos, setHoverPos] = useState<{x: number, y: number} | null>(null);
  const mouseRef = useRef<{x: number, y: number} | null>(null);
  
  // Animation state
  const prevRef = useRef<Coords5D>(coords);
  const animRef = useRef<Coords5D>(coords);
  const tRef = useRef(1); // 1 = settled

  // Start fold when coords change
  useEffect(() => {
    prevRef.current = animRef.current;
    tRef.current = 0; // start animation
  }, [coords]);

  const project5Dto3D = (c: Coords5D) => {
    const x = c.x1 * 0.6 + c.x4 * 0.4;
    const y = c.x2 * 0.7 + c.x5 * 0.3;
    const z = c.x3;
    return { x, y, z };
  };

  const rotate3D = (p: any, theta: number) => {
    const cosT = Math.cos(theta);
    const sinT = Math.sin(theta);
    return {
      x: p.x * cosT - p.z * sinT,
      y: p.y,
    };
  };

  const offset = (c: Coords5D) => ({
    x1: c.x1 + (Math.random() - 0.5) * 200,
    x2: c.x2 + (Math.random() - 0.5) * 200,
    x3: c.x3 + (Math.random() - 0.5) * 200,
    x4: c.x4 + (Math.random() - 0.5) * 200,
    x5: c.x5 + (Math.random() - 0.5) * 200,
  });

  const bezier5D = (p0: Coords5D, p1: Coords5D, p2: Coords5D, p3: Coords5D, t: number) => {
    const u = 1 - t;
    const uu = u * u;
    const tt = t * t;

    return {
      x1: uu * u * p0.x1 + 3 * uu * t * p1.x1 + 3 * u * tt * p2.x1 + tt * t * p3.x1,
      x2: uu * u * p0.x2 + 3 * uu * t * p1.x2 + 3 * u * tt * p2.x2 + tt * t * p3.x2,
      x3: uu * u * p0.x3 + 3 * uu * t * p1.x3 + 3 * u * tt * p2.x3 + tt * t * p3.x3,
      x4: uu * u * p0.x4 + 3 * uu * t * p1.x4 + 3 * u * tt * p2.x4 + tt * t * p3.x4,
      x5: uu * u * p0.x5 + 3 * uu * t * p1.x5 + 3 * u * tt * p2.x5 + tt * t * p3.x5,
    };
  };

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    const center = size / 2;

    let animationId: number;
    let c1 = offset(prevRef.current);
    let c2 = offset(coords);

    const animate = () => {
      thetaRef.current += 0.01;
      const theta = thetaRef.current;

      ctx.clearRect(0, 0, size, size);

      // Interpolate animation
      if (tRef.current < 1) {
        tRef.current = Math.min(1, tRef.current + 0.02); // animation speed
        const t = tRef.current * tRef.current * (3 - 2 * tRef.current); // cubic ease

        // Regenerate control points if we started a new animation
        if (tRef.current === 0.02) {
            c1 = offset(prevRef.current);
            c2 = offset(coords);
        }

        animRef.current = bezier5D(prevRef.current, c1, c2, coords, t);
      } else {
        animRef.current = coords;
      }

      // Draw historical path / trail
      history.forEach((h, i) => {
        const p3 = project5Dto3D(h);
        const p2 = rotate3D(p3, theta);
        const alpha = Math.max(0.1, i / history.length); // fade older history

        ctx.fillStyle = `rgba(80, 200, 255, ${alpha * 0.4})`;
        ctx.beginPath();
        ctx.arc(center + p2.x / 4, center - p2.y / 4, 3, 0, Math.PI * 2);
        ctx.fill();

        if (i > 0) {
           const prevH = history[i-1];
           const prevP3 = project5Dto3D(prevH);
           const prevP2 = rotate3D(prevP3, theta);
           ctx.strokeStyle = `rgba(80, 200, 255, ${alpha * 0.2})`;
           ctx.lineWidth = 1;
           ctx.beginPath();
           ctx.moveTo(center + prevP2.x / 4, center - prevP2.y / 4);
           ctx.lineTo(center + p2.x / 4, center - p2.y / 4);
           ctx.stroke();
        }
      });

      // Draw current animated event node
      const p3 = project5Dto3D(animRef.current);
      const p2 = rotate3D(p3, theta);

      // Glow intensifies during fold
      const glowIntensity = tRef.current < 1 ? 15 + Math.sin(tRef.current * Math.PI) * 10 : 10;
      
      ctx.fillStyle = "rgba(255, 255, 255, 0.9)";
      ctx.shadowColor = "#4ecdc4";
      ctx.shadowBlur = glowIntensity;
      ctx.beginPath();
      ctx.arc(center + p2.x / 4, center - p2.y / 4, 6, 0, Math.PI * 2);
      ctx.fill();
      
      ctx.shadowBlur = 0;

      // ----------------------------------------------------
      // Node Collision Detection
      // ----------------------------------------------------
      const m = mouseRef.current;
      let closestNode: Coords5D | null = null;
      let minDistance = 15; // 15px hover radius

      if (m) {
        // Check historical nodes
        for (const h of history) {
            const hp3 = project5Dto3D(h);
            const hp2 = rotate3D(hp3, theta);
            const screenX = center + hp2.x / 4;
            const screenY = center - hp2.y / 4;
            const dist = Math.hypot(m.x - screenX, m.y - screenY);
            if (dist < minDistance) {
                minDistance = dist;
                closestNode = h;
            }
        }

        // Check current node
        const currScreenX = center + p2.x / 4;
        const currScreenY = center - p2.y / 4;
        const currDist = Math.hypot(m.x - currScreenX, m.y - currScreenY);
        if (currDist < minDistance) {
            closestNode = animRef.current;
        }
      }
      
      setHoveredNode(closestNode);
      if (closestNode && m) {
        setHoverPos(m);
      } else {
        setHoverPos(null);
      }

      animationId = requestAnimationFrame(animate);
    };

    animate();

    return () => {
      cancelAnimationFrame(animationId);
    };
  }, [coords, history, size]);

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
      const rect = e.currentTarget.getBoundingClientRect();
      mouseRef.current = {
          x: e.clientX - rect.left,
          y: e.clientY - rect.top,
      };
  };

  const handleMouseLeave = () => {
      mouseRef.current = null;
  };

  return (
    <div 
        style={{ position: "relative", width: size, height: size, margin: "auto" }}
        onMouseMove={handleMouseMove}
        onMouseLeave={handleMouseLeave}
    >
        <canvas ref={canvasRef} width={size} height={size} style={{ display: "block" }} />
        {hoveredNode && hoverPos && (
            <div style={{
                position: "absolute",
                left: hoverPos.x + 15,
                top: hoverPos.y + 15,
                backgroundColor: "rgba(20, 20, 20, 0.9)",
                border: "1px solid #4ecdc4",
                padding: "8px 12px",
                borderRadius: 6,
                color: "#fff",
                fontFamily: "monospace",
                fontSize: "12px",
                pointerEvents: "none",
                zIndex: 10,
                boxShadow: "0 0 10px rgba(78, 205, 196, 0.3)"
            }}>
                <div style={{ color: "#4ecdc4", marginBottom: 4, fontWeight: "bold" }}>Node Coordinate</div>
                <div>x₁: {hoveredNode.x1.toFixed(1)}</div>
                <div>x₂: {hoveredNode.x2.toFixed(1)}</div>
                <div>x₃: {hoveredNode.x3.toFixed(1)}</div>
                <div>x₄: {hoveredNode.x4.toFixed(1)}</div>
                <div>x₅: {hoveredNode.x5.toFixed(1)}</div>
            </div>
        )}
    </div>
  );
};
