export type TimelineEvent = {
  tick: number;
  realTimeMs: number;
  sessionId: string;
  agentId: string;
  page: number;
  data: string;
  coords5d?: { x1: number; x2: number; x3: number; x4: number; x5: number };
  evolutionary_version: number;
};

export type AgentLane = {
  agentId: string;
  color: string;
};
