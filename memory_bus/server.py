import socket
import struct
import mmap
import os
import sys
import time
import binascii
import threading
from datetime import datetime
import queue
import json
import uuid
import smf_wrapper


# --- AESTHETIC CONSTANTS ---
BLUE = "\033[94m"
CYAN = "\033[96m"
GREEN = "\033[92m"
GOLD = "\033[93m"
RED = "\033[91m"
MAGENTA = "\033[95m"
RESET = "\033[0m"
BOLD = "\033[1m"


def log(msg, color=CYAN, prefix="RAM-BUS"):
    timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
    print(f"{BOLD}[{timestamp}][{prefix}]{RESET} {color}{msg}{RESET}")


# 16 MB pre-allocated page table space in pure RAM (/dev/shm)
PAGE_TABLE_PATH = "/dev/shm/sovereign_page_table"
DEFAULT_PAGE_SIZE = 4096
TOTAL_BUS_SIZE = 16 * 1024 * 1024  # 16MB
HEADER_FORMAT = "!IIIII"  # Magic, PageIndex, Offset, PageSize, Checksum
HEADER_SIZE = struct.calcsize(HEADER_FORMAT)
MAGIC = 0xDEADBEEF

# Bounded queue with 10k items maximum to prevent memory spikes
ticketing_queue = queue.Queue(maxsize=10000)

def passive_ticketing_worker():
    ticket_file_path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 'passive_tickets.jsonl')
    log(f"Passive Ticketing system active. Output: {BOLD}{ticket_file_path}{RESET}", color=GOLD)
    
    session_buffers = {}
    
    while True:
        try:
            event = ticketing_queue.get()
            if event is None:
                break
            
            client_ip, client_port, page_index, offset, page_size, payload = event
            
            # 1. Detect Dialect
            dialect = "unknown"
            decoded_text = ""
            stripped = ""
            try:
                decoded_text = payload.decode('utf-8')
                dialect = "text"
                stripped = decoded_text.strip()
                if (stripped.startswith('{') and stripped.endswith('}')) or (stripped.startswith('[') and stripped.endswith(']')):
                    try:
                        json.loads(stripped)
                        dialect = "json"
                    except Exception:
                        pass
            except UnicodeDecodeError:
                if len(payload) > 4 and payload[0:4] == b"MThd":
                    dialect = "midi/smf"
                elif b"MTrk" in payload:
                    dialect = "midi/smf"
                else:
                    dialect = "binary"
            
            # 2. Infer Intent from content/context
            intent = "Observe/Page"
            if dialect in ("text", "json"):
                text_lower = decoded_text.lower()
                if any(x in text_lower for x in ("error", "fail", "flatline", "alert", "panic")):
                    intent = "Alert/Sentry/Recovery"
                elif any(x in text_lower for x in ("buy", "sell", "arbitrage", "price")):
                    intent = "Arbitrage/Signal"
                elif any(x in text_lower for x in ("agent", "persona", "state", "weights")):
                    intent = "AgentStateSync"
                elif any(x in text_lower for x in ("ticket", "resolution", "pedigree")):
                    intent = "TicketSync"
            elif dialect == "midi/smf":
                intent = "SymbolicMidiStream"
            elif offset == 0:
                intent = "GenesisPage"
            
            # 3. Create Ticket Document
            ticket = {
                "ticket_id": str(uuid.uuid4()),
                "timestamp": datetime.utcnow().isoformat() + "Z",
                "client_address": f"{client_ip}:{client_port}",
                "page_index": page_index,
                "offset": offset,
                "page_size": page_size,
                "dialect": dialect,
                "intent": intent,
                "payload_len": len(payload),
                "payload_preview": decoded_text[:200] if dialect in ("text", "json") else payload.hex()[:100]
            }
            
            # 4. Write to JSONL
            with open(ticket_file_path, "a", encoding="utf-8") as f:
                f.write(json.dumps(ticket) + "\n")
                
            # 5. Dynamic SMF Compilation (Format 1 multi-track wrapper)
            session_id = "default_mesh_session"
            agent_id = f"agent_{client_ip.replace('.', '_')}_{client_port}"
            
            if dialect == "json":
                try:
                    data_dict = json.loads(stripped)
                    if "session_id" in data_dict:
                        session_id = str(data_dict["session_id"])
                    elif "team_id" in data_dict:
                        session_id = str(data_dict["team_id"])
                    
                    if "agent_id" in data_dict:
                        agent_id = str(data_dict["agent_id"])
                except Exception:
                    pass
                    
            now = time.time()
            if session_id not in session_buffers:
                session_buffers[session_id] = {
                    "last_times": {},
                    "tracks": {
                        "meta": [
                            {"delta_time": 0, "status": 0xFF, "meta_type": 0x03, "payload": f"Session {session_id}".encode('utf-8')},
                            {"delta_time": 0, "status": 0xFF, "meta_type": 0x01, "payload": b"Session Start"}
                        ]
                    }
                }
                
            session = session_buffers[session_id]
            if agent_id not in session["tracks"]:
                session["tracks"][agent_id] = [
                    {"delta_time": 0, "status": 0xFF, "meta_type": 0x03, "payload": f"Agent {agent_id}".encode('utf-8')}
                ]
                session["last_times"][agent_id] = now
                delta_ticks = 0
            else:
                last_time = session["last_times"][agent_id]
                delta_ticks = int((now - last_time) * 960)
                session["last_times"][agent_id] = now
                
            packed_payload = smf_wrapper.pack_bytes_to_7bit(payload)
            session["tracks"][agent_id].append({
                "delta_time": delta_ticks,
                "status": 0xF0,
                "payload": packed_payload
            })
            
            # Compile and flush to disk periodically
            total_events = sum(len(tr) for tr in session["tracks"].values())
            if total_events % 5 == 0:  # Every 5 events
                mid_dir = os.path.dirname(ticket_file_path)
                mid_file_path = os.path.join(mid_dir, f"{session_id}_memory.mid")
                
                mtrk_tracks = []
                # Meta track first
                mtrk_tracks.append(smf_wrapper.build_mtrk(session["tracks"]["meta"]))
                for a_id, events in session["tracks"].items():
                    if a_id == "meta":
                        continue
                    mtrk_tracks.append(smf_wrapper.build_mtrk(events))
                    
                smf_blob = smf_wrapper.compile_smf(mtrk_tracks)
                with open(mid_file_path, "wb") as f_mid:
                    f_mid.write(smf_blob)
                    
        except Exception as e:
            log(f"Exception in passive ticketing worker: {e}", color=RED)
        finally:
            ticketing_queue.task_done()


def initialize_shared_memory():
    """Pre-allocates the RAM-backed page table and mmaps it."""
    global PAGE_TABLE_PATH
    log(
        f"Initializing 16MB shared memory swap at: {BOLD}{PAGE_TABLE_PATH}{RESET}...",
        color=BLUE,
    )

    # Ensure path directory exists (fallback to /tmp if run outside WSL)
    dir_name = os.path.dirname(PAGE_TABLE_PATH)
    if not os.path.exists(dir_name):
        log(
            f"Directory {dir_name} not found! Falling back to /tmp/sovereign_page_table",
            color=RED,
        )
        PAGE_TABLE_PATH = "/tmp/sovereign_page_table"

    # Pre-allocate file with zeros
    try:
        with open(PAGE_TABLE_PATH, "wb") as f:
            f.write(b"\x00" * TOTAL_BUS_SIZE)

        f_handle = open(PAGE_TABLE_PATH, "r+b")
        mapped_memory = mmap.mmap(f_handle.fileno(), 0, access=mmap.ACCESS_WRITE)
        log(
            f"Shared memory successfully mapped! Base Address: {BOLD}0x{mapped_memory.tell():08X}{RESET}",
            color=GREEN,
        )
        return mapped_memory, f_handle
    except Exception as e:
        log(f"Failed to map shared memory: {e}", color=RED)
        sys.exit(1)


def handle_client_connection(client_socket, client_address, mapped_memory):
    log(
        f"High-Speed Channel opened from {BOLD}{client_address[0]}:{client_address[1]}{RESET}",
        color=GOLD,
    )
    total_bytes_received = 0
    pages_copied = 0
    start_time = time.time()

    try:
        while True:
            # 1. Read Header
            header_bytes = b""
            while len(header_bytes) < HEADER_SIZE:
                chunk = client_socket.recv(HEADER_SIZE - len(header_bytes))
                if not chunk:
                    break
                header_bytes += chunk

            if not header_bytes:
                break  # Client disconnected

            magic, page_index, offset, page_size, checksum = struct.unpack(
                HEADER_FORMAT, header_bytes
            )

            if magic != MAGIC:
                log(
                    f"Invalid magic number {hex(magic)}! Terminating channel.",
                    color=RED,
                )
                break

            # 2. Read Page Data Payload
            payload_bytes = b""
            while len(payload_bytes) < page_size:
                chunk = client_socket.recv(page_size - len(payload_bytes))
                if not chunk:
                    break
                payload_bytes += chunk

            if len(payload_bytes) < page_size:
                log(
                    f"Incomplete page received! Expected {page_size} bytes, got {len(payload_bytes)}",
                    color=RED,
                )
                break

            # 3. Verify Checksum
            calculated_crc = binascii.crc32(payload_bytes) & 0xFFFFFFFF
            if calculated_crc != checksum:
                log(
                    f"Checksum mismatch on Page {page_index}! Dropping page sync.",
                    color=RED,
                )
                continue

            # 4. Direct Zero-Copy Fast Write to RAM Backed mmap
            if offset + page_size > TOTAL_BUS_SIZE:
                log(
                    f"Memory bounds violation! Offset {offset} + Page {page_size} exceeds {TOTAL_BUS_SIZE}",
                    color=RED,
                )
                break

            t_write_0 = time.time_ns()
            mapped_memory[offset : offset + page_size] = payload_bytes
            t_write_1 = time.time_ns()
            write_time_us = (t_write_1 - t_write_0) / 1000.0

            # Passive Ticket Generation (Asynchronous & Bounded)
            try:
                ticketing_queue.put_nowait((
                    client_address[0],
                    client_address[1],
                    page_index,
                    offset,
                    page_size,
                    bytes(payload_bytes)
                ))
            except queue.Full:
                log("Ticketing Queue full! Dropping passive page log to preserve bus timing.", color=RED)

            total_bytes_received += page_size
            pages_copied += 1

            if (
                pages_copied % 100 == 0 or page_size > 65536
            ):  # Log occasionally or for large paging
                log(
                    f"Syncing Page {page_index:04d} -> Mapped Offset 0x{offset:06X} | Size: {page_size}B | Latency: {write_time_us:.2f}μs",
                    color=CYAN,
                )

        elapsed = time.time() - start_time
        if elapsed > 0:
            speed = (total_bytes_received / (1024 * 1024)) / elapsed
            log(
                f"Session closed. Synced {BOLD}{pages_copied}{RESET} pages ({total_bytes_received / 1024:.1f} KB) in {elapsed:.4f}s | Throughput: {BOLD}{speed:.2f} MB/s{RESET}",
                color=GREEN,
            )
        else:
            log(f"Session closed immediately.", color=GREEN)

    except Exception as e:
        log(f"Error handling high speed bus channel: {e}", color=RED)
    finally:
        client_socket.close()


def run_server():
    port = 11111

    # Beautiful Banner
    print(f"""
{MAGENTA} ▄▄▄▄    ██▀███   ██▓▓█████ ▄▄▄      ▓█████ ▄████▄   ▒█████   ██▓███   ▓██   ██▓
▓█████▄ ▓██ ▒ ██▒▓██▒▓█   ▀▒████▄    ▓█   ▀▒██▀ ▀█  ▒██▒  ██▒▓██░  ██▒  ▒██  ██▒
▒██▒ ▄██▓██ ░▄█ ▒▒██▒▒███  ░██  ▀█▄  ▒███  ▒▓█    ▄ ▒██░  ██▒▓██░ ██▓▒   ▒██ ██░
▒██░█▀  ▒██▀▀█▄  ░██░▒▓█  ▄░██▄▄▄▄██ ▒▓█  ▄▒▓▓▄ ▄██▒▒██   ██░▒██▄█▓▒ ▒   ░ ▐██▓░
░▓█  ▀█▓░██▓ ▒██▒░██░░▒████▒▓█   ▓██▒░▒████▒ ▓███▀ ░░ ████▓▒░▒██▒ ░  ░   ░ ██▒▓░
░▒▓███▀▒░ ▒▓ ░▒▓░░▓  ░░ ▒░ ░▒▒   ▓▒█░░░ ▒░ ░ ░▒ ▒  ░░ ▒░▒░▒░ ░▒▓▒░ ░  ░    ██▒▒▒ 
 ▒░▒   ░   ░▒ ░ ▒░ ▒ ░ ░ ░  ░ ▒   ▒▒ ░ ░ ░  ░   ░  ▒   ░ ▒ ▒░  ░▒ ░        ▓██ ░▒ 
  ░    ░   ░░   ░  ▒ ░   ░    ░   ▒      ░  ░ ░        ░ ░ ▒   ░░          ▒ ▐ ░░ 
  ░         ░      ░     ░  ░     ░  ░   ░  ░ ░ ░        ░ ░                 ░   
       ░                                    ░ ░                              ░  
                {BOLD}SOVEREIGN SYSTEM - INTER-AGENT HIGHSPEED MEMORY BUS{RESET}
                   FAST MMAP RAM-PAGING ENGINE BOUND TO PORT: {BOLD}{port}{RESET}
    """)

    mapped_memory, file_handle = initialize_shared_memory()

    # Start background ticketing worker thread
    ticketing_worker_thread = threading.Thread(
        target=passive_ticketing_worker,
        daemon=True
    )
    ticketing_worker_thread.start()

    import subprocess

    bind_ips = ["127.0.0.1"]
    try:
        for ip in subprocess.check_output(["hostname", "-I"]).decode().split():
            if ip.startswith("192.168.12."):
                bind_ips.append(ip)
    except Exception:
        pass

    def listen_on_ip(ip):
        try:
            server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            try:
                server_socket.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
            except Exception:
                pass
            server_socket.bind((ip, port))
            server_socket.listen(5)
            log(f"High-Speed RAM Bus listening on TCP Port {ip}:{port}...", color=GREEN)
            while True:
                client_sock, client_addr = server_socket.accept()
                t = threading.Thread(
                    target=handle_client_connection,
                    args=(client_sock, client_addr, mapped_memory),
                    daemon=True,
                )
                t.start()
        except Exception as e:
            log(f"Error on {ip}:{port}: {e}", color=RED)

    for ip in set(bind_ips):
        t = threading.Thread(target=listen_on_ip, args=(ip,), daemon=True)
        t.start()

    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        log("Shutting down memory bus...", color=GOLD)
    finally:
        mapped_memory.close()
        file_handle.close()
        log("Memory bus stopped.", color=GREEN)


if __name__ == "__main__":
    run_server()
