import React, { useMemo, useState, useRef, useEffect } from "react";
import { TimelineEvent, AgentLane } from "./types";
import { Radar5D } from "./Radar5D";
import { Minimap } from "./Minimap";
import { Tensor27StripChart } from "./Tensor27StripChart";
import { WebGLManifold5D, AgentPath } from "./WebGLManifold5D";
import { fftReal, magnitudeSpectrum } from "./fft";
import { HelixEvolution } from "./HelixEvolution";

type Props = {
  events: TimelineEvent[];
};

export const TimelineView: React.FC<Props> = ({ events }) => {
  const [selected, setSelected] = useState<TimelineEvent | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playIndex, setPlayIndex] = useState(0);
  const [isLive, setIsLive] = useState(true);
  const [latestTick, setLatestTick] = useState<number | null>(null);
  const [playbackSpeed, setPlaybackSpeed] = useState(500);

  const WINDOW = 1024;
  const densityRef = useRef(new Float32Array(WINDOW));
  const spectrumRef = useRef(new Float32Array(WINDOW / 2));
  const writeIndexRef = useRef(0);
  const frameCountRef = useRef(0);

  // NEW: navigation state
  const [scaleX, setScaleX] = useState(1);
  const [translateX, setTranslateX] = useState(0);
  const [playhead, setPlayhead] = useState<number | null>(null);
  const [selectedAgents, setSelectedAgents] = useState<string[]>([]);
  
  const helixPointRef = useRef<{x: number, y: number, z: number} | null>(null);
  const manifoldPointRef = useRef<{x: number, y: number, z: number} | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const dragging = useRef(false);
  const lastX = useRef(0);

  // Lanes
  const lanes = useMemo(() => {
    const ids = Array.from(new Set(events.map(e => e.agentId))).sort();
    return ids.map((agentId, i) => ({
      agentId,
      color: laneColor(i),
    }));
  }, [events]);

  useEffect(() => {
    if (selectedAgents.length === 0 && lanes.length > 0) {
      setSelectedAgents(lanes.map(l => l.agentId));
    }
  }, [lanes]);

  
  useEffect(() => {
    const ws = new WebSocket("ws://localhost:8080/live");
    ws.binaryType = "arraybuffer";
    wsRef.current = ws;

    const enc = new TextEncoder();
    ws.onmessage = (ev) => {
      if (typeof ev.data === "string") {
        const buf = enc.encode(ev.data as string);
        if ((window as any).IngestEvent) (window as any).IngestEvent(buf);
      } else {
        const buf = new Uint8Array(ev.data as ArrayBuffer);
        if ((window as any).IngestEvent) (window as any).IngestEvent(buf);
      }
    };

    return () => ws.close();
  }, []);

  const pollControlDecision = React.useCallback(() => {
    try {
      if ((window as any).GetControlDecisionJSON) {
        return (window as any).GetControlDecisionJSON();
      }
    } catch {
      // ignore
    }
    return null;
  }, []);


  type LiveAgent = {
    agent_id: string;
    coords5d: { x1:number; x2:number; x3:number; x4:number; x5:number };
    evolutionary_version: number;
    tick: number;
  };

  
  const getTensor27Summary = React.useCallback(() => {
    try {
      const summaryArr = (window as any).GetTensor27Summary ? (window as any).GetTensor27Summary() : undefined;
      if (summaryArr && summaryArr.length === 3) {
        return {
          pos: summaryArr[0],
          neg: summaryArr[1],
          zero: summaryArr[2]
        };
      }
    } catch {
      // ignore
    }
    return null;
  }, []);

  const [liveAgents, setLiveAgents] = useState<LiveAgent[]>([]);

  useEffect(() => {
    let rafId: number;
    const dec = new TextDecoder();

    const loop = () => {
      const arr = (window as any).GetAgentsSnapshot ? (window as any).GetAgentsSnapshot() as Uint8Array | undefined : undefined;
      if (arr && arr.byteLength > 0) {
        const json = dec.decode(arr);
        const parsed = JSON.parse(json) as LiveAgent[];
        setLiveAgents(parsed);

        let maxTick = latestTick ?? 0;
        for (const a of parsed) {
          if (a.tick > maxTick) maxTick = a.tick;
        }
        
        if (maxTick > 0) {
            setLatestTick(maxTick);
            // record density
            densityRef.current[writeIndexRef.current] = parsed.length;
            writeIndexRef.current = (writeIndexRef.current + 1) % WINDOW;
            
            frameCountRef.current++;
            if (frameCountRef.current % 6 === 0) {
                const rotated = new Float32Array(WINDOW);
                for (let j = 0; j < WINDOW; j++) {
                    rotated[j] = densityRef.current[(writeIndexRef.current + j) % WINDOW];
                }
                const { re, im } = fftReal(rotated);
                spectrumRef.current = magnitudeSpectrum(re, im);
            }
        }
      }
      const controlJson = pollControlDecision();
      if (controlJson && wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(controlJson);
      }
      rafId = requestAnimationFrame(loop);
    };

    loop();
    return () => cancelAnimationFrame(rafId);
  }, [latestTick]);

  useEffect(() => {
    if (!isLive || latestTick == null) return;
    const x = (latestTick * scaleX) + translateX;
    const center = 1200 / 2;
    setTranslateX(t => t + (center - x));
  }, [isLive, latestTick, scaleX]);

  useEffect(() => {
    if (!isPlaying || events.length === 0) return;

    const id = setInterval(() => {
      setPlayIndex(i => {
        const next = i + 1;
        return next >= events.length ? 0 : next;
      });

      // Update Density Buffer
      densityRef.current[writeIndexRef.current] = 1; // Assuming 1 event processed
      writeIndexRef.current = (writeIndexRef.current + 1) % WINDOW;
      
      frameCountRef.current++;
      if (frameCountRef.current % 6 === 0) {
        const rotated = new Float32Array(WINDOW);
        for (let j = 0; j < WINDOW; j++) {
            rotated[j] = densityRef.current[(writeIndexRef.current + j) % WINDOW];
        }
        const { re, im } = fftReal(rotated);
        spectrumRef.current = magnitudeSpectrum(re, im);
      }
    }, playbackSpeed);

    return () => clearInterval(id);
  }, [isPlaying, events.length, playbackSpeed]);

  useEffect(() => {
    if (events.length > 0) {
      const ev = events[playIndex];
      setSelected(ev);
      if (isPlaying) {
        setPlayhead(ev.tick);
        setTranslateX(t => {
          const x = (ev.tick * scaleX) + t;
          if (x < 200 || x > 1200 - 200) { // width is 1200
            return t - (x - 1200 / 2);
          }
          return t;
        });
      }
    }
  }, [playIndex, isPlaying, events, scaleX]);

  const maxTick = useMemo(
    () => events.reduce((m, e) => Math.max(m, e.tick), 0),
    [events]
  );

  const laneHeight = 40;
  const width = 1200;
  const height = Math.max(400, lanes.length * laneHeight + 80); // Ensure minimum height

  // Transform tick → screen X
  const xScale = (tick: number) => tick * scaleX + translateX;

  const yForAgent = (agentId: string) => {
    const idx = lanes.findIndex(l => l.agentId === agentId);
    return 40 + idx * laneHeight + laneHeight / 2;
  };

  const numBins = 100;
  const density = useMemo(() => {
    if (events.length === 0 || maxTick === 0) return new Array(numBins).fill(0);
    const bins = new Array(numBins).fill(0);
    events.forEach(ev => {
      let bin = Math.floor((ev.tick / maxTick) * numBins);
      if (bin === numBins) bin = numBins - 1;
      bins[bin]++;
    });
    return bins;
  }, [events, maxTick]);

  const normalizedDensity = useMemo(() => {
    const maxD = Math.max(...density, 1);
    return density.map(d => d / maxD);
  }, [density]);

  const viewStartTick = (-translateX) / scaleX;
  const viewEndTick = (width - translateX) / scaleX;

  // -----------------------------
  // ZOOM (mouse wheel)
  // -----------------------------
  const onWheel = (e: React.WheelEvent<SVGSVGElement>) => {
    e.preventDefault();

    const cursorX = e.clientX - e.currentTarget.getBoundingClientRect().left;
    const zoomFactor = e.deltaY < 0 ? 1.1 : 0.9;

    const newScale = Math.min(20, Math.max(0.1, scaleX * zoomFactor));

    const cursorTick = (cursorX - translateX) / scaleX;
    const newTranslate = cursorX - cursorTick * newScale;

    setScaleX(newScale);
    setTranslateX(newTranslate);
  };

  // -----------------------------
  // PAN (drag)
  // -----------------------------
  const onMouseDown = (e: React.MouseEvent) => {
    dragging.current = true;
    lastX.current = e.clientX;
  };

  const onMouseMove = (e: React.MouseEvent) => {
    if (!dragging.current) return;

    const dx = e.clientX - lastX.current;
    lastX.current = e.clientX;

    if (e.shiftKey) {
      // SCRUB MODE
      const svgRect = e.currentTarget.getBoundingClientRect();
      const cursorX = e.clientX - svgRect.left;
      const tick = (cursorX - translateX) / scaleX;
      setPlayhead(tick);
    } else {
      // PAN MODE
      setTranslateX(t => t + dx);
    }
  };

  const onMouseUp = () => {
    dragging.current = false;
  };
  
  const onMouseLeave = () => {
    dragging.current = false;
  };

  const overlayAgents = selectedAgents.map((agentId) => {
    const laneIdx = lanes.findIndex(l => l.agentId === agentId);
    const color = laneIdx >= 0 ? lanes[laneIdx].color : "#fff";
    const agentEvents = events.filter(e => e.agentId === agentId && e.coords5d);
    const history = agentEvents.map(e => e.coords5d!);
    
    let current;
    if (selected && selected.agentId === agentId && selected.coords5d) {
      current = selected.coords5d;
    } else if (history.length > 0) {
      current = history[history.length - 1];
    } else {
      current = { x1: 0, x2: 0, x3: 0, x4: 0, x5: 0 };
    }

    return { agentId, color, current, history };
  });

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <div style={{ display: "flex", gap: 16, alignItems: "center", background: "#121212", padding: "8px 16px", borderRadius: 8, border: "1px solid #333" }}>
        <button 
          onClick={() => setIsPlaying(p => !p)}
          style={{ background: isPlaying ? "#ff6b6b" : "#4ecdc4", color: "#000", border: "none", padding: "8px 16px", borderRadius: 4, cursor: "pointer", fontWeight: "bold" }}
        >
          {isPlaying ? "Pause" : "Play"}
        </button>
        <button
          onClick={() => setIsLive(true)}
          style={{
            padding: "8px 16px",
            borderRadius: 8,
            border: "1px solid #333",
            cursor: "pointer",
            background: isLive ? "#ff4b4b" : "#222",
            color: "white",
            display: "flex",
            alignItems: "center",
            gap: 8,
            fontWeight: "bold"
          }}
        >
          <span
            style={{
              width: 8,
              height: 8,
              borderRadius: "50%",
              background: isLive ? "#ffdfdf" : "#777",
              boxShadow: isLive ? "0 0 8px #ffdfdf" : "none",
            }}
          />
          LIVE
        </button>
        <label style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 14 }}>
          Speed:
          <input
            type="range"
            min="10"
            max="2000"
            value={playbackSpeed}
            onChange={e => setPlaybackSpeed(Number(e.target.value))}
          />
          {playbackSpeed}ms
        </label>
      </div>

      <div className="timeline-container" style={{ display: "flex", gap: 16 }}>
        <div style={{ display: "flex", flexDirection: "column" }}>
        <svg
          width={width}
          height={height}
          style={{ border: "1px solid #333", backgroundColor: "#0a0a0a", borderRadius: 8, cursor: dragging.current ? (playhead !== null ? "ew-resize" : "grabbing") : "grab" }}
          onWheel={onWheel}
          onMouseDown={onMouseDown}
          onMouseMove={onMouseMove}
          onMouseUp={onMouseUp}
          onMouseLeave={onMouseLeave}
        >
          {/* density glow */}
          {normalizedDensity.map((d, i) => {
            const x = (i / numBins) * width;
            const barWidth = width / numBins;

            return (
              <rect
                key={i}
                x={xScale(x * maxTick / width)}
                y={0}
                width={barWidth * scaleX}
                height={height}
                fill="rgba(255, 200, 0, 0.15)"
                opacity={d}
              />
            );
          })}

          {/* lanes */}
        {lanes.map((lane, i) => (
          <g key={lane.agentId}>
            <line
              x1={0}
              x2={width}
              y1={40 + i * laneHeight + laneHeight / 2}
              y2={40 + i * laneHeight + laneHeight / 2}
              stroke="#333"
              strokeDasharray="4 4"
            />
            <text
              x={10}
              y={40 + i * laneHeight + laneHeight / 2 + 4}
              fill={lane.color}
              fontSize={12}
            >
              {lane.agentId}
            </text>
          </g>
        ))}

        {/* events */}
        {events.map((ev, idx) => {
          const x = xScale(ev.tick);
          const y = yForAgent(ev.agentId);
          const lane = lanes.find(l => l.agentId === ev.agentId)!;

          const size = 8;
          const shape = ev.page % 2 === 0 ? "circle" : "square";

          if (shape === "circle") {
            return (
              <circle
                key={idx}
                cx={x}
                cy={y}
                r={size}
                fill={lane.color}
                onClick={() => {
                  setSelected(ev);
                  setPlayIndex(idx);
                }}
                style={{ cursor: "pointer", transition: "r 0.1s" }}
                onMouseOver={(e) => (e.currentTarget.r.baseVal.value = size * 1.5)}
                onMouseOut={(e) => (e.currentTarget.r.baseVal.value = size)}
              />
            );
          }

          return (
            <rect
              key={idx}
              x={x - size}
              y={y - size}
              width={size * 2}
              height={size * 2}
              fill={lane.color}
              onClick={() => {
                setSelected(ev);
                setPlayIndex(idx);
              }}
              style={{ cursor: "pointer", transition: "width 0.1s, height 0.1s" }}
            />
          );
        })}

        {/* playhead */}
        {playhead !== null && (
          <line
            x1={xScale(playhead)}
            x2={xScale(playhead)}
            y1={0}
            y2={height}
            stroke="#4ecdc4"
            strokeWidth={2}
            strokeDasharray="2 2"
          />
        )}
      </svg>
      
      <Minimap
        width={width}
        density={normalizedDensity}
        viewStart={maxTick > 0 ? viewStartTick / maxTick : 0}
        viewEnd={maxTick > 0 ? viewEndTick / maxTick : 1}
        onJump={(fraction) => {
          const targetTick = fraction * maxTick;
          const newTranslate = -targetTick * scaleX + width / 2;
          setTranslateX(newTranslate);
        }}
      />
      </div>

      {/* right panel */}
      <div style={{ minWidth: 360, backgroundColor: "#121212", padding: 16, borderRadius: 8, border: "1px solid #333", maxHeight: "100%", overflowY: "auto" }}>
        
        <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
          {lanes.map(lane => (
            <label key={lane.agentId} style={{ display: "flex", alignItems: "center", gap: 4, color: lane.color, fontSize: 12, cursor: "pointer" }}>
              <input 
                type="checkbox" 
                checked={selectedAgents.includes(lane.agentId)} 
                onChange={(e) => {
                  if (e.target.checked) setSelectedAgents([...selectedAgents, lane.agentId]);
                  else setSelectedAgents(selectedAgents.filter(a => a !== lane.agentId));
                }} 
              />
              {lane.agentId}
            </label>
          ))}
        </div>

        <h2 style={{ marginTop: 0, color: "#4ecdc4" }}>Event Details</h2>
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <WebGLManifold5D 
            agents={liveAgents.length > 0 ? liveAgents.map((a: any) => ({
              agentId: a.agent_id,
              color: laneColor(a.agent_id),
              current: a.coords5d,
              history: []
            })) : overlayAgents} 
            selectedAgentId={selected?.agentId}
            helixPointRef={helixPointRef}
            manifoldPointRef={manifoldPointRef}
            spectrumRef={spectrumRef}
            size={240} 
          />
          {selected ? (
            <>
              {selected.coords5d && (
                <div style={{ display: "flex", gap: 16, alignItems: "center", justifyContent: "center" }}>
                  <Radar5D coords={selected.coords5d} size={150} />
                  <HelixEvolution 
                    versions={events.filter(e => e.agentId === selected.agentId && e.evolutionary_version).map(e => e.evolutionary_version)} 
                    color={laneColor(lanes.findIndex(l => l.agentId === selected.agentId))} 
                    helixPointRef={helixPointRef}
                    manifoldPointRef={manifoldPointRef}
                    size={150} 
                  />
                </div>
              )}
              <pre style={{ fontSize: 12, color: "#ddd", whiteSpace: "pre-wrap", background: "#1e1e1e", padding: 8, borderRadius: 4 }}>
                {JSON.stringify(selected, null, 2)}
              </pre>
            </>
          ) : (
            <p style={{ color: "#888", fontSize: 14 }}>Click an event to inspect its specific memory page details.</p>
          )}
        </div>
        </div>
      </div>
    </div>
  );
};

function laneColor(i: number): string {
  const palette = ["#ff6b6b", "#4ecdc4", "#ffe66d", "#5f6caf", "#ff9f1c"];
  return palette[i % palette.length];
}
