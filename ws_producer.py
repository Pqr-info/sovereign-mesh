import asyncio
import websockets
import json
import random
import time

mutation_scale = 1.0
coupling_gain = 1.0
input_band = "mid"

async def handler(websocket):
    global mutation_scale, coupling_gain, input_band
    print("[Producer] Client connected.")
    
    async def event_generator():
        tick = 0
        while True:
            await asyncio.sleep(0.05) # 20fps
            tick += 1
            # Generate a mock synthetic event string
            # Just sending empty to trigger ticks
            # In real system this is a valid byte stream or json
            # But the mesh requires `IngestEvent` format.
            # We will send a valid JSON string that parses to TimelineEvent
            event = [{
                "tick": tick,
                "realTimeMs": int(time.time()*1000),
                "sessionId": "mock-session",
                "agentId": f"agent-{random.randint(1,5)}",
                "page": 0,
                "coords5d": {
                    "x1": random.randint(-100, 100),
                    "x2": random.randint(-100, 100),
                    "x3": random.randint(-100, 100),
                    "x4": random.randint(-100, 100),
                    "x5": random.randint(-100, 100)
                },
                "evolutionary_version": random.randint(1,10)
            }]
            try:
                await websocket.send(json.dumps(event))
            except websockets.exceptions.ConnectionClosed:
                break

    async def control_listener():
        global mutation_scale, coupling_gain, input_band
        try:
            async for message in websocket:
                try:
                    msg = json.loads(message)
                    if msg.get("type") == "control" and msg.get("intent") == "adjust_regime":
                        p = msg.get("payload", {})
                        mutation_scale = max(0.5, min(1.5, p.get("mutation_scale", 1.0)))
                        coupling_gain = max(0.5, min(1.5, p.get("coupling_gain", 1.0)))
                        input_band = p.get("input_band", "mid")
                        
                        regime = p.get("regime", "neutral")
                        print(f"\\n[Organism Intent] {regime.upper()}")
                        print(f" > Mutation: {mutation_scale:.2f} | Coupling: {coupling_gain:.2f} | Band: {input_band}")
                except Exception as e:
                    pass
        except websockets.exceptions.ConnectionClosed:
            pass

    await asyncio.gather(
        event_generator(),
        control_listener()
    )
    print("[Producer] Client disconnected.")

async def main():
    async with websockets.serve(handler, "localhost", 8080):
        print("Mock Synthetic Physics Generator listening on ws://localhost:8080/live")
        await asyncio.Future()  # run forever

if __name__ == "__main__":
    asyncio.run(main())
