package raft

// 是本地程序与其他节点程序传输数据的桥梁
type Transporter interface {
	//	SendVoteRequest(server Server, peer *Peer, req *RequestVoteRequest) *RequestVoteResponse
	//	SendAppendEntriesRequest(server Server, peer *Peer, req *AppendEntriesRequest) *AppendEntriesResponse
	SendSnapshotRequest(server Server, peer *Peer, req *SnapshotRequest) *SnapshotResponse
	SendSnapshotRecoveryRequest(server Server, peer *Peer, req *SnapshotRecoveryRequest) *SnapshotRecoveryResponse
}
