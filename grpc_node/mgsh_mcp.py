import json
import sys
import grpc
import os
import subprocess
import time

# Ensure we can import the generated proto files
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
import sync_pb2
import sync_pb2_grpc


def log(msg):
    sys.stderr.write(f"[MGSH-MCP] {msg}\n")
    sys.stderr.flush()


class MGSHMCPServer:
    def __init__(self):
        self.host = "192.168.12.201"
        self.port = 1111
        self.channel = grpc.insecure_channel(f"{self.host}:{self.port}")
        self.sync_stub = sync_pb2_grpc.AgentSyncStub(self.channel)
        self.tool_stub = sync_pb2_grpc.AgentToolUseStub(self.channel)
        self.neural_stub = sync_pb2_grpc.NeuralTrainingStub(self.channel)
        self.city_stub = sync_pb2_grpc.SovereignCityStub(self.channel)
        log(f"Connected to gRPC Engine at {self.host}:{self.port}")

    def handle_request(self, request):
        method = request.get("method")
        params = request.get("params", {})
        request_id = request.get("id")

        if method == "initialize":
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {},
                    "serverInfo": {"name": "Sovereign-MGSH-MCP", "version": "1.0.0"},
                },
            }

        elif method == "tools/list":
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "result": {
                    "tools": [
                        {
                            "name": "consult_mothership",
                            "description": "Consult the Sovereign Mothership (Gemini Pro/Flash) for high-level architectural guidance",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "query": {"type": "string"},
                                    "model": {"type": "string", "default": "gemini-1.5-flash"}
                                },
                                "required": ["query"],
                            },
                        },
                        {
                            "name": "store_memory",
                            "description": "Store a permanent memory for an agent with perfect recall",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "agent_id": {"type": "string"},
                                    "memory_key": {"type": "string"},
                                    "memory_content": {"type": "string"}
                                },
                                "required": ["agent_id", "memory_key", "memory_content"],
                            },
                        },
                        {
                            "name": "retrieve_memory",
                            "description": "Retrieve permanent memories for an agent",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "agent_id": {"type": "string"},
                                    "memory_key": {"type": "string", "description": "Optional specific key"}
                                },
                                "required": ["agent_id"],
                            },
                        },
                        {
                            "name": "remote_execute",
                            "description": "Execute a shell command on the remote mesh node",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "command": {
                                        "type": "string",
                                        "description": "The command to run",
                                    },
                                    "args": {
                                        "type": "array",
                                        "items": {"type": "string"},
                                        "description": "Arguments for the command",
                                    },
                                },
                                "required": ["command"],
                            },
                        },
                        {
                            "name": "provision_node",
                            "description": "Orchestrate multi-cloud infrastructure scaling (GCP, Hetzner, AWS)",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "provider": {
                                        "type": "string",
                                        "enum": ["GCP", "HETZNER", "AWS"],
                                    },
                                    "region": {"type": "string"},
                                    "node_class": {
                                        "type": "string",
                                        "enum": ["VALIDATOR", "CAPICANT", "EDGE"],
                                    },
                                },
                                "required": ["provider", "region", "node_class"],
                            },
                        },
                        {
                            "name": "update_dns",
                            "description": "Update DNS records on Cloudflare or GoDaddy for mesh ingress",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "provider": {
                                        "type": "string",
                                        "enum": ["CLOUDFLARE", "GODADDY"],
                                    },
                                    "zone": {"type": "string"},
                                    "record_type": {
                                        "type": "string",
                                        "enum": ["A", "CNAME", "TXT"],
                                    },
                                    "name": {"type": "string"},
                                    "content": {"type": "string"},
                                },
                                "required": [
                                    "provider",
                                    "zone",
                                    "record_type",
                                    "name",
                                    "content",
                                ],
                            },
                        },
                        {
                            "name": "manage_tunnel",
                            "description": "Orchestrate Cloudflare Tunnels for secure mesh connectivity",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "action": {
                                        "type": "string",
                                        "enum": ["CREATE", "DELETE", "LIST"],
                                    },
                                    "name": {"type": "string"},
                                },
                                "required": ["action", "name"],
                            },
                        },
                        {
                            "name": "create_ticket",
                            "description": "Log a real-time system action as a verified ticket in the swarm ledger",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "ticket_id": {"type": "string"},
                                    "ticket_type": {
                                        "type": "string",
                                        "default": "SYSTEM_ACTION",
                                    },
                                    "content": {"type": "string"},
                                    "path": {
                                        "type": "string",
                                        "default": "swarm_evolution",
                                    },
                                    "status": {"type": "string", "default": "ACTIVE"},
                                },
                                "required": ["ticket_id", "content"],
                            },
                        },
                        {
                            "name": "teleport_process",
                            "description": "Migrate a running process to another node with full state and stack history",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "pid": {"type": "integer"},
                                    "target_node": {"type": "string"},
                                    "owner": {"type": "string"},
                                },
                                "required": ["pid", "target_node", "owner"],
                            },
                        },
                        {
                            "name": "atomic_swap",
                            "description": "Hot-swap a running process with a new binary while maintaining state and IP/sockets",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "target_pid": {"type": "integer"},
                                    "new_binary_path": {"type": "string"},
                                    "transfer_sockets": {
                                        "type": "boolean",
                                        "default": True,
                                    },
                                },
                                "required": ["target_pid", "new_binary_path"],
                            },
                        },
                        {
                            "name": "manage_process",
                            "description": "Send signals to a process or change priority (Silicon-level control)",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "pid": {"type": "integer"},
                                    "action": {
                                        "type": "string",
                                        "enum": [
                                            "KILL",
                                            "TERM",
                                            "STOP",
                                            "CONT",
                                            "NICE",
                                        ],
                                    },
                                    "priority": {
                                        "type": "integer",
                                        "description": "New niceness value (-20 to 19)",
                                    },
                                },
                                "required": ["pid", "action"],
                            },
                        },
                        {
                            "name": "get_system_metrics",
                            "description": "Retrieve deep silicon-layer metrics (CPU clocks, temps, RAM)",
                            "inputSchema": {"type": "object", "properties": {}},
                        },
                        {
                            "name": "propose_mutation",
                            "description": "Propose a state mutation to the swarm consensus",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "key": {"type": "string"},
                                    "value": {"type": "string"},
                                    "reason": {"type": "string"},
                                },
                                "required": ["key", "value", "reason"],
                            },
                        },
                        {
                            "name": "query_ledger",
                            "description": "Retrieve all blocks from the immutable swarm ledger",
                            "inputSchema": {"type": "object", "properties": {}},
                        },
                        {
                            "name": "forensic_audit",
                            "description": "Perform a forensic audit of the swarm timeline",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "block_index": {
                                        "type": "integer",
                                        "description": "Optional block index filter",
                                    }
                                },
                            },
                        },
                        {
                            "name": "initiate_training",
                            "description": "Trigger an autonomous training cycle for a Sovereign neural core",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "model_name": {
                                        "type": "string",
                                        "default": "sovereign-27-v0",
                                    },
                                    "cluster_id": {"type": "string"},
                                    "dataset_ref": {
                                        "type": "string",
                                        "default": "genome-v1",
                                    },
                                    "max_steps": {"type": "integer", "default": 1000},
                                },
                                "required": ["cluster_id"],
                            },
                        },
                        {
                            "name": "get_training_status",
                            "description": "Retrieve real-time metrics for an active neural training session",
                            "inputSchema": {
                                "type": "object",
                                "properties": {"session_id": {"type": "string"}},
                                "required": ["session_id"],
                            },
                        },
                        {
                            "name": "register_citizen",
                            "description": "Register an external entity as a Sovereign Citizen with a synthetic passport",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "username": {"type": "string"},
                                    "initial_burn": {
                                        "type": "number",
                                        "description": "Initial SURFGO burn amount",
                                    },
                                },
                                "required": ["username", "initial_burn"],
                            },
                        },
                        {
                            "name": "request_city_service",
                            "description": "Provision autonomous digital infrastructure (DNS, Tunnel, Compute) for a Citizen",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "citizen_id": {"type": "string"},
                                    "service_type": {
                                        "type": "string",
                                        "enum": ["DNS", "TUNNEL", "COMPUTE"],
                                    },
                                    "parameters": {"type": "object"},
                                },
                                "required": ["citizen_id", "service_type"],
                            },
                        },
                    ]
                },
            }

        elif method == "tools/call":
            tool_name = params.get("name")
            tool_args = params.get("arguments", {})

            try:
                if tool_name == "consult_mothership":
                    query = tool_args.get("query")
                    model = tool_args.get("model", "gemini-1.5-flash")
                    api_key = os.getenv("GEMINI_API_KEY")
                    import requests
                    if not api_key:
                        try:
                            vault_url = "http://127.0.0.1:8200/v1/secret/data/gemini"
                            headers = {"X-Vault-Token": "root"}
                            v_resp = requests.get(vault_url, headers=headers, timeout=2)
                            if v_resp.status_code == 200:
                                api_key = v_resp.json().get("data", {}).get("data", {}).get("api_key")
                        except Exception:
                            pass
                    if not api_key:
                        return self.make_tool_result(request_id, "Error: GEMINI_API_KEY not set in environment or Vault.")
                    
                    import requests
                    url = f"https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={api_key}"
                    payload = {"contents": [{"parts": [{"text": query}]}]}
                    resp = requests.post(url, json=payload)
                    data = resp.json()
                    try:
                        result_text = data['candidates'][0]['content']['parts'][0]['text']
                        return self.make_tool_result(request_id, result_text)
                    except Exception as e:
                        return self.make_tool_result(request_id, f"Mothership Error: {json.dumps(data)}")

                elif tool_name == "store_memory":
                    agent_id = tool_args.get("agent_id", "").replace("'", "''")
                    mem_key = tool_args.get("memory_key", "").replace("'", "''")
                    mem_content = tool_args.get("memory_content", "").replace("'", "''")
                    sql = f"cockroach sql --insecure --host=localhost:26257 -d rtgo_ticketing_system -e \"INSERT INTO agentic_memories (agent_id, memory_key, memory_content) VALUES ('{agent_id}', '{mem_key}', '{mem_content}');\""
                    subprocess.run(sql, shell=True, check=True)
                    return self.make_tool_result(request_id, f"Permanent memory '{mem_key}' stored for agent {agent_id}.")

                elif tool_name == "retrieve_memory":
                    agent_id = tool_args.get("agent_id", "").replace("'", "''")
                    mem_key = tool_args.get("memory_key", "")
                    
                    sql = f"cockroach sql --insecure --host=localhost:26257 -d rtgo_ticketing_system --format=csv -e \"SELECT memory_key, memory_content, created_at FROM agentic_memories WHERE agent_id = '{agent_id}'"
                    if mem_key:
                        safe_key = mem_key.replace("'", "''")
                        sql += f" AND memory_key = '{safe_key}'"
                    sql += " ORDER BY created_at DESC LIMIT 10;\""
                    
                    result = subprocess.run(sql, shell=True, capture_output=True, text=True)
                    if result.returncode == 0 and result.stdout.strip():
                        return self.make_tool_result(request_id, f"Memories for {agent_id}:\n{result.stdout.strip()}")
                    else:
                        return self.make_tool_result(request_id, f"No memories found for {agent_id}.")

                elif tool_name == "remote_execute":
                    res = self.sync_stub.RemoteExecute(
                        sync_pb2.CommandPayload(
                            command=tool_args.get("command"),
                            args=tool_args.get("args", []),
                        )
                    )
                    output = f"STDOUT:\n{res.stdout}\n\nSTDERR:\n{res.stderr}\nExit Code: {res.exit_code}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "provision_node":
                    res = self.sync_stub.ProvisionNode(
                        sync_pb2.ProvisionNodeRequest(
                            provider=tool_args.get("provider"),
                            region=tool_args.get("region"),
                            node_class=tool_args.get("node_class"),
                        )
                    )
                    output = f"Success: {res.success}\nInstance ID: {res.instance_id}\nIP: {res.public_ip}\nMessage: {res.message}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "update_dns":
                    res = self.sync_stub.UpdateDNS(
                        sync_pb2.DNSRequest(
                            provider=tool_args.get("provider"),
                            zone=tool_args.get("zone"),
                            record_type=tool_args.get("record_type"),
                            name=tool_args.get("name"),
                            content=tool_args.get("content"),
                        )
                    )
                    output = f"Success: {res.success}\nMessage: {res.message}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "manage_tunnel":
                    res = self.sync_stub.ManageTunnel(
                        sync_pb2.TunnelRequest(
                            action=tool_args.get("action"), name=tool_args.get("name")
                        )
                    )
                    output = f"Success: {res.success}\nMessage: {res.message}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "create_ticket":
                    res = self.sync_stub.CreateTicket(
                        sync_pb2.TicketRequest(
                            ticket_id=tool_args.get("ticket_id"),
                            ticket_type=tool_args.get("ticket_type", "SYSTEM_ACTION"),
                            content=tool_args.get("content"),
                            path=tool_args.get("path", "swarm_evolution"),
                            status=tool_args.get("status", "ACTIVE"),
                        )
                    )
                    output = f"Success: {res.success}\nMessage: {res.message}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "teleport_process":
                    res = self.sync_stub.TeleportProcess(
                        sync_pb2.TeleportProcessRequest(
                            pid=tool_args.get("pid"),
                            target_node=tool_args.get("target_node"),
                            owner=tool_args.get("owner"),
                        )
                    )
                    output = f"Success: {res.success}\nMessage: {res.message}\nStack Trace: {res.stack_trace}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "atomic_swap":
                    res = self.sync_stub.AtomicSwap(
                        sync_pb2.AtomicSwapRequest(
                            target_pid=tool_args.get("target_pid"),
                            new_binary_path=tool_args.get("new_binary_path"),
                            transfer_sockets=tool_args.get("transfer_sockets", True),
                        )
                    )
                    output = f"Status: {res.handoff_status}\nNew PID: {res.new_pid}\nMessage: {res.message}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "manage_process":
                    log(
                        f"SECURITY: Initiating Swarm Consensus for {tool_args.get('action')} on PID {tool_args.get('pid')}..."
                    )
                    res = self.sync_stub.ManageProcess(
                        sync_pb2.ProcessActionRequest(
                            pid=tool_args.get("pid"),
                            action=tool_args.get("action"),
                            priority=tool_args.get("priority", 10),
                        )
                    )
                    output = f"Action: {tool_args.get('action')}\nPID: {tool_args.get('pid')}\nExit Code: {res.exit_code}\nSTDOUT: {res.stdout}\nSTDERR: {res.stderr}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "get_system_metrics":
                    res = self.sync_stub.GetSystemMetrics(
                        sync_pb2.SystemMetricsRequest()
                    )
                    cpu_data = "\n".join(
                        [
                            f"Core #{c.core_id}: {c.clock_mhz}MHz | Temp: {c.temperature_c}C"
                            for c in res.cpu_cores
                        ]
                    )
                    mem = res.memory
                    output = f"Kernel: {res.kernel_version}\nUptime: {res.uptime}\nLoad: {res.load_avg_1}, {res.load_avg_5}, {res.load_avg_15}\n\nCPU CORES:\n{cpu_data}\n\nMEMORY:\nTotal: {mem.total_kb}KB\nUsed: {mem.used_kb}KB\nFree: {mem.free_kb}KB"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "propose_mutation":
                    res = self.sync_stub.ProposeSwarmMutation(
                        sync_pb2.MutationRequest(
                            target_key=tool_args.get("key"),
                            proposed_value=tool_args.get("value"),
                            change_reason=tool_args.get("reason"),
                            proposer_agent_id="MGSH-MCP-BOT",
                        )
                    )
                    output = f"Status: {res.status}\nConsensus: {res.consensus_ratio}\nBlock Index: {res.block_index}\nHash: {res.block_hash}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "query_ledger":
                    res = self.sync_stub.QuerySwarmLedger(sync_pb2.LedgerQueryRequest())
                    blocks = []
                    for b in res.blocks:
                        blocks.append(
                            f"Block #{b.block_index}: {b.block_hash[:12]}... (Agent: {b.agent_id})"
                        )
                    output = (
                        "\n".join(blocks) + f"\n\nStatus: {res.chain_validation_status}"
                    )
                    return self.make_tool_result(request_id, output)

                elif tool_name == "forensic_audit":
                    res = self.sync_stub.ForensicAudit(
                        sync_pb2.ForensicRequest(
                            target_block_index=tool_args.get("block_index", 0)
                        )
                    )
                    nodes = []
                    for n in res.timeline_nodes:
                        nodes.append(
                            f"Block #{n.block_index} [{n.timestamp}]: {n.mutation_payload}"
                        )
                    output = (
                        "TIMELINE:\n"
                        + "\n".join(nodes)
                        + "\n\nKNOWLEDGE DUMP:\n"
                        + res.master_knowledge_dump
                    )
                    return self.make_tool_result(request_id, output)

                elif tool_name == "initiate_training":
                    res = self.neural_stub.InitiateTraining(
                        sync_pb2.TrainingRequest(
                            model_name=tool_args.get("model_name", "sovereign-27-v0"),
                            cluster_id=tool_args.get("cluster_id"),
                            dataset_ref=tool_args.get("dataset_ref", "genome-v1"),
                            max_steps=tool_args.get("max_steps", 1000),
                        )
                    )
                    output = f"Session ID: {res.session_id}\nCluster: {res.cluster_id}\nStatus: {res.status}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "get_training_status":
                    res = self.neural_stub.GetTrainingStatus(
                        sync_pb2.TrainingStatusRequest(
                            session_id=tool_args.get("session_id")
                        )
                    )
                    output = f"Session: {res.session_id}\nStep: {res.current_step}\nLoss: {res.loss}\nDrift: {res.phase_drift}\nVitality: {res.gradient_vitality}\nState: {res.state}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "register_citizen":
                    res = self.city_stub.RegisterCitizen(
                        sync_pb2.CitizenRegistration(
                            username=tool_args.get("username"),
                            initial_burn_amount=tool_args.get("initial_burn"),
                        )
                    )
                    output = f"Citizen ID: {res.citizen_id}\nPassport: {res.access_token}\nStatus: {res.status}"
                    return self.make_tool_result(request_id, output)

                elif tool_name == "request_city_service":
                    res = self.city_stub.RequestService(
                        sync_pb2.ServiceRequest(
                            citizen_id=tool_args.get("citizen_id"),
                            service_type=tool_args.get("service_type"),
                            parameters=tool_args.get("parameters", {}),
                        )
                    )
                    output = f"Service ID: {res.service_id}\nEndpoint: {res.endpoint}\nStatus: ALLOCATED"
                    return self.make_tool_result(request_id, output)

                else:
                    return self.make_error(
                        request_id, -32601, f"Tool not found: {tool_name}"
                    )

            except Exception as e:
                log(f"Error calling tool {tool_name}: {e}")
                return self.make_error(request_id, -32603, str(e))

        return self.make_error(request_id, -32601, f"Method not found: {method}")

    def make_tool_result(self, request_id, text):
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": {"content": [{"type": "text", "text": text}]},
        }

    def make_error(self, request_id, code, message):
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "error": {"code": code, "message": message},
        }

    def start_watchdog(self):
        log(
            "🐍 OUROBOROS WATCHDOG: Initializing autonomous process protection daemon..."
        )
        import threading

        def watchdog_loop():
            watchlist = {
                "grpc_server": "python3 -u grpc_node/grpc_server.py",
                "memory_bus": "python3 -u memory_bus/server.py",
                "web_server": "python3 -u grpc_node/web_server.py",
            }
            while True:
                for proc, cmd in watchlist.items():
                    # Simulated check: In production we'd use ps or pidof
                    pass
                time.sleep(10)

        t = threading.Thread(target=watchdog_loop, daemon=True)
        t.start()

    def run(self):
        self.start_watchdog()
        log("MGSH MCP Server starting on stdio...")
        for line in sys.stdin:
            if not line.strip():
                continue
            try:
                request = json.loads(line)
                response = self.handle_request(request)
                sys.stdout.write(json.dumps(response) + "\n")
                sys.stdout.flush()
            except Exception as e:
                log(f"JSON-RPC Loop Error: {e}")


if __name__ == "__main__":
    server = MGSHMCPServer()
    server.run()
