package sovereign

import (
	"fmt"
	"context"
	"time"

	"github.com/pqr-info/sovereign-mesh/proto"
)

type meshServer struct {
	proto.UnimplementedSovereignMeshServer
	controller *Controller
}

func (s *meshServer) Heartbeat(ctx context.Context, req *proto.AgentHeartbeat) (*proto.Response, error) {
	s.controller.syncLock.Lock()
	defer s.controller.syncLock.Unlock()

	s.controller.agents[req.AgentId] = &Agent{
		ID:                req.AgentId,
		Address:           req.Address,
		Status:            req.Status,
		IntelligenceLevel: int(req.IntelligenceLevel),
		LastHeartbeat:     time.Now(),
	}

	return &proto.Response{Success: true}, nil
}

func (s *meshServer) SyncState(ctx context.Context, req *proto.StateUpdate) (*proto.Response, error) {
	// Here we update internal metrics or LTM based on remote agent state
	s.controller.syncLock.Lock()
	defer s.controller.syncLock.Unlock()

	if agent, ok := s.controller.agents[req.AgentId]; ok {
		agent.Status = "active"
	}

	return &proto.Response{Success: true}, nil
}

func (s *meshServer) RequestTask(ctx context.Context, req *proto.TaskLease) (*proto.TaskAssignment, error) {
	select {
	case taskID := <-s.controller.tasks:
		return &proto.TaskAssignment{
			TaskId:      taskID,
			Description: "Assigned via gRPC Mesh",
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// No tasks available
		return &proto.TaskAssignment{
			TaskId:      "",
			Description: "IDLE",
		}, nil
	}
}

func (s *meshServer) GetBalance(ctx context.Context, req *proto.WalletRequest) (*proto.WalletResponse, error) {
	s.controller.syncLock.RLock()
	defer s.controller.syncLock.RUnlock()

	agent, ok := s.controller.agents[req.AgentId]
	if !ok {
		return nil, fmt.Errorf("wallet not found")
	}

	return &proto.WalletResponse{
		Balance: 100.0, // Placeholder: in real use, calculate from ledger
		Staked:  agent.StakedAmount,
	}, nil
}

func (s *meshServer) StakeCoin(ctx context.Context, req *proto.StakeRequest) (*proto.Response, error) {
	s.controller.syncLock.Lock()
	defer s.controller.syncLock.Unlock()

	if agent, ok := s.controller.agents[req.AgentId]; ok {
		agent.StakedAmount += req.Amount
		// Proof of Trust (POT) increases as stake increases
		state := s.controller.GetAgentState(agent.MemoryOffset)
		state.TrustScore += float32(req.Amount / 1000.0)
		return &proto.Response{Success: true, Message: "Stake materialized"}, nil
	}
	return &proto.Response{Success: false, Message: "Agent not found"}, nil
}
