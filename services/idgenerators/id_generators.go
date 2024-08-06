package idgenerators

import (
	idgenerator "github.com/ecumenos-social/id-generator"
	"github.com/ecumenos-social/toolkitfx/fxidgenerator"
)

type HoldersIDGeneratorConfig fxidgenerator.Config

type HoldersIDGenerator idgenerator.Generator

func NewHoldersIDGenerator(config *HoldersIDGeneratorConfig) (HoldersIDGenerator, error) {
	return idgenerator.New(&idgenerator.NodeID{
		Top: config.TopNodeID,
		Low: config.LowNodeID,
	})
}

type HolderSessionsIDGeneratorConfig fxidgenerator.Config

type HolderSessionsIDGenerator idgenerator.Generator

func NewHolderSessionsIDGenerator(config *HolderSessionsIDGeneratorConfig) (HolderSessionsIDGenerator, error) {
	return idgenerator.New(&idgenerator.NodeID{
		Top: config.TopNodeID,
		Low: config.LowNodeID,
	})
}

type SentEmailsIDGeneratorConfig fxidgenerator.Config

type SentEmailsIDGenerator idgenerator.Generator

func NewSentEmailsIDGenerator(config *SentEmailsIDGeneratorConfig) (SentEmailsIDGenerator, error) {
	return idgenerator.New(&idgenerator.NodeID{
		Top: config.TopNodeID,
		Low: config.LowNodeID,
	})
}

type NetworkNodesIDGeneratorConfig fxidgenerator.Config

type NetworkNodesIDGenerator idgenerator.Generator

func NewNetworkNodesIDGenerator(config *NetworkNodesIDGeneratorConfig) (NetworkNodesIDGenerator, error) {
	return idgenerator.New(&idgenerator.NodeID{
		Top: config.TopNodeID,
		Low: config.LowNodeID,
	})
}

type PersonalDataNodesIDGeneratorConfig fxidgenerator.Config

type PersonalDataNodesIDGenerator idgenerator.Generator

func NewPersonalDataNodesIDGenerator(config *PersonalDataNodesIDGeneratorConfig) (PersonalDataNodesIDGenerator, error) {
	return idgenerator.New(&idgenerator.NodeID{
		Top: config.TopNodeID,
		Low: config.LowNodeID,
	})
}

type NetworkWardensIDGeneratorConfig fxidgenerator.Config

type NetworkWardensIDGenerator idgenerator.Generator

func NewNetworkWardensIDGenerator(config *NetworkWardensIDGeneratorConfig) (NetworkWardensIDGenerator, error) {
	return idgenerator.New(&idgenerator.NodeID{
		Top: config.TopNodeID,
		Low: config.LowNodeID,
	})
}
