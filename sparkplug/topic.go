package sparkplug

import "fmt"

// Sparkplug B v2.2 topic namespace root.
const Namespace = "spBv1.0"

// Message type tokens (Section 6.1).
const (
	NBIRTH = "NBIRTH"
	NDEATH = "NDEATH"
	NDATA  = "NDATA"
	NCMD   = "NCMD"
	DBIRTH = "DBIRTH"
	DDEATH = "DDEATH"
	DDATA  = "DDATA"
	DCMD   = "DCMD"
	STATE  = "STATE"
)

// NodeTopic returns spBv1.0/{group}/{msgType}/{eonNode}.
func NodeTopic(group, msgType, eonNode string) string {
	return fmt.Sprintf("%s/%s/%s/%s", Namespace, group, msgType, eonNode)
}

// DeviceTopic returns spBv1.0/{group}/{msgType}/{eonNode}/{device}.
func DeviceTopic(group, msgType, eonNode, device string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", Namespace, group, msgType, eonNode, device)
}
