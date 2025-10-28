package service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/p4wl/lovp-shared-models/events"
)

type CommandHandlerService struct {
	logger *slog.Logger
	nm     *NetManager
	data   chan []byte
}

func NewCommandHandlerService(logger *slog.Logger, data chan []byte, nm *NetManager) *CommandHandlerService {
	return &CommandHandlerService{
		logger: logger,
		nm:     nm,
		data:   data,
	}
}

func (s *CommandHandlerService) HandleRawCmd(stop chan bool) {
	s.logger.Info("Handling event")
	go func() {
		for {
			select {
			case <-stop:
				s.logger.Info("Stopping service")
			default:
				msg := <-s.data
				s.logger.Info("NetworkService got message", "message", msg)

				var envelope events.CmdEnvelope
				if err := json.Unmarshal(msg, &envelope); err != nil {
					s.logger.Error(fmt.Sprintf("Could not unmarshal event envelope %s", string(msg)))
					continue
				}

				switch envelope.EventType {
				case events.CREATE_NETWORK:
					var cmd events.NetworkCreateCmd
					if err := json.Unmarshal(envelope.Data, &cmd); err != nil {
						s.logger.Error(fmt.Sprintf("Could not unmarshal event command %s", string(msg)))
						continue
					}

					_, err := s.nm.CreateNetwork(cmd.NetworkName, cmd.Owner, cmd.Hosts)
					if err != nil {
						s.logger.Error("Failed to create network", "error", err)
						continue
					}
				case events.DELETE_NETWORK:
					var cmd events.NetworkDeleteCmd
					if err := json.Unmarshal(envelope.Data, &cmd); err != nil {
						s.logger.Error(fmt.Sprintf("Could not unmarshal event command %s", string(msg)))
						continue
					}

					// create and call s.nm.DeleteNetwork()

				case events.CREATE_USER:
					// unmarshal as events.UserCreateCmd
					var cmd events.UserCreateCmd
					if err := json.Unmarshal(envelope.Data, &cmd); err != nil {
						s.logger.Error(fmt.Sprintf("Could not unmarshal event command %s", string(msg)))
						continue
					}

					err := s.nm.RegisterUser(cmd.Username, cmd.Email)
					if err != nil {
						s.logger.Error("Failed to register user", "error", err)
						continue
					}
				case events.ADD_USER_TO_NETWORK:
				default:
					s.logger.Error("Unknown command")
				}

			}
		}
	}()
}
