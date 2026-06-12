package sovereign

import (
	"log"
)

// PrisonerDilemma simulates competitive defection/cooperation analysis.
type PrisonerDilemma struct {
	Agents []*DAOAgent
	Payoff map[string]int
}

// PlayMatch simulates a PD match. 
// Payoff: Both Cooperate (+3), Both Defect (+1), One Defects (+5), One Cooperates (0).
func (pd *PrisonerDilemma) PlayMatch(a1, a2 *DAOAgent) (int, int) {
	// Strategy Logic: Humans lean towards Cooperation, GT agents optimize for Game-Theory-Perfect Defection
	a1Coop := a1.Archetype == "HUMAN_DESIGN"
	a2Coop := a2.Archetype == "HUMAN_DESIGN"
	
	if a1Coop && a2Coop { return 3, 3 }
	if a1Coop && !a2Coop { return 0, 5 }
	if !a1Coop && a2Coop { return 5, 0 }
	return 1, 1
}

// RunTournament audits swarm alignment.
func (pd *PrisonerDilemma) RunTournament() {
	log.Println("⚖️ PD-AUDIT: Commencing Prisoner's Dilemma alignment audit...")
	
	for i := 0; i < len(pd.Agents); i++ {
		for j := i + 1; j < len(pd.Agents); j++ {
			p1, p2 := pd.PlayMatch(pd.Agents[i], pd.Agents[j])
			pd.Payoff[pd.Agents[i].ID] += p1
			pd.Payoff[pd.Agents[j].ID] += p2
		}
	}
	
	log.Println("📊 PD-AUDIT: Alignment check complete. Identifying defectors...")
}
