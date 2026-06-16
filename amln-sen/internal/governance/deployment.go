package governance

// NodeDeploymentClass specifies the hardware resource tiers
type NodeDeploymentClass string

const (
	ClassEdgeNodeNode   NodeDeploymentClass = "Tier-3 Edge Node"
	ClassRegionalServer NodeDeploymentClass = "Tier-2 Regional Server"
	ClassGlobalCoreNode NodeDeploymentClass = "Tier-1 Global Core Node"
)

// HardwareProfile outlines target node resource specs
type HardwareProfile struct {
	MinRAMGigs   int    `json:"min_ram_gigs"`
	StorageType  string `json:"storage_type"`
	NetworkClass string `json:"network_class"`
}

type SovereignNodeDeployment struct {
	Class   NodeDeploymentClass `json:"class"`
	Profile HardwareProfile     `json:"profile"`
	MeshID  string              `json:"mesh_id"`
}

func GetHardwareProfile(class NodeDeploymentClass) HardwareProfile {
	switch class {
	case ClassGlobalCoreNode:
		return HardwareProfile{
			MinRAMGigs:   64,
			StorageType:  "NVMe SSD",
			NetworkClass: "10 Gbps Backbone",
		}
	case ClassRegionalServer:
		return HardwareProfile{
			MinRAMGigs:   16,
			StorageType:  "SATA SSD",
			NetworkClass: "1 Gbps Dedicated",
		}
	default:
		return HardwareProfile{
			MinRAMGigs:   4,
			StorageType:  "eMMC Flash",
			NetworkClass: "Broadband Mesh",
		}
	}
}
