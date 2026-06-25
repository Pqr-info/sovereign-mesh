import React, { useEffect, useState } from "react";
import { initWasm, parseSmfToTimeline } from "./wasm";
import { TimelineEvent } from "./types";
import { TimelineView } from "./TimelineView";
import { DropZone } from "./DropZone";

export const App: React.FC = () => {
  const [ready, setReady] = useState(false);
  const [events, setEvents] = useState<TimelineEvent[]>([]);

  useEffect(() => {
    initWasm().then(() => setReady(true));
  }, []);

  return (
    <div className="app" style={{ padding: 20, fontFamily: "sans-serif", backgroundColor: "#1e1e1e", color: "#eee", minHeight: "100vh" }}>
      <header>
        <h1 style={{ color: "#4ecdc4" }}>Sovereign Mesh Time Machine</h1>
        <DropZone onFile={async (file) => {
          if (!ready) return;
          const buf = new Uint8Array(await file.arrayBuffer());
          const timeline = parseSmfToTimeline(buf) as TimelineEvent[];
          setEvents(timeline);
        }} />
      </header>

      <main>
        {events.length === 0 ? (
          <p>Load a .mid session to view the timeline.</p>
        ) : (
          <TimelineView events={events} />
        )}
      </main>
    </div>
  );
};

export default App;
