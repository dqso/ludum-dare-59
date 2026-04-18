package entity

type NPC interface {
	Type() NPCType
}

type NPCType string

const (
	NPCRecruiter       = NPCType("recruiter")
	NPCTechInterviewer = NPCType("tech interviewer")
)
