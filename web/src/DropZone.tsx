import React, { useState, DragEvent } from "react";

type Props = {
  onFile: (file: File) => void;
};

export const DropZone: React.FC<Props> = ({ onFile }) => {
  const [hover, setHover] = useState(false);

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setHover(false);

    const file = e.dataTransfer.files?.[0];
    if (file && file.name.endsWith(".mid")) {
      onFile(file);
    }
  };

  return (
    <div
      onDragOver={e => {
        e.preventDefault();
        setHover(true);
      }}
      onDragLeave={() => setHover(false)}
      onDrop={handleDrop}
      style={{
        border: hover ? "3px dashed #4ecdc4" : "3px dashed #666",
        padding: 40,
        borderRadius: 12,
        textAlign: "center",
        color: "#aaa",
        transition: "0.2s",
        margin: "20px 0",
        cursor: "pointer"
      }}
    >
      <h3 style={{ margin: 0 }}>Drop a .mid session file here</h3>
      <p style={{ fontSize: 12, marginTop: 8 }}>to visualize the temporal evolution</p>
    </div>
  );
};
